// Package storage содержит функции для работы с базой данных свечей
// Market Loader
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"market-loader/internal/money"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
	"github.com/sirupsen/logrus"
)

// Candle структура для хранения данных свечи
type Candle struct {
	FIGI         string    `json:"figi"`
	Time         time.Time `json:"time"`
	OpenPrice    float64   `json:"open_price"`
	HighPrice    float64   `json:"high_price"`
	LowPrice     float64   `json:"low_price"`
	ClosePrice   float64   `json:"close_price"`
	Volume       int64     `json:"volume"`
	IntervalType string    `json:"interval_type"`
}

// GetLastLoadedTime получает время последней загрузки из таблицы candles
func GetLastLoadedTime(ctx context.Context, dbpool *pgxpool.Pool, figi, intervalType string) (time.Time, error) {
	query := `SELECT MAX(time) FROM candles WHERE figi = $1 AND interval_type = $2`

	var lastLoadedTime sql.NullTime
	err := dbpool.QueryRow(ctx, query, figi, intervalType).Scan(&lastLoadedTime)

	// Если нет данных (NULL) или ошибка
	if err != nil {
		return time.Time{}, fmt.Errorf("ошибка выполнения запроса к таблице candles: %w", err)
	}

	// Если MAX(time) вернул NULL (нет свечей)
	if !lastLoadedTime.Valid {
		return time.Time{}, nil // данных нет — это нормально
	}

	return lastLoadedTime.Time, nil
}

// GetEarliestCandle получает самую раннюю свечу
func GetEarliestCandle(dbpool *pgxpool.Pool, figi, intervalType string) (time.Time, error) {
	query := `SELECT MIN(time) FROM candles WHERE figi = $1 AND interval_type = $2`

	var earliestTime sql.NullTime
	err := dbpool.QueryRow(context.Background(), query, figi, intervalType).Scan(&earliestTime)

	if err == pgx.ErrNoRows || !earliestTime.Valid {
		return time.Time{}, nil
	}

	return earliestTime.Time, fmt.Errorf("ошибка сканирования самого раннего времени: %w", err)
}

// GetLastCandleTime возвращает время последней загруженной свечи для инструмента и интервала
func GetLastCandleTime(ctx context.Context, dbpool *pgxpool.Pool, figi, intervalType string) (time.Time, error) {
	query := `
		SELECT MAX("time") 
		FROM candles 
		WHERE figi = $1 AND interval_type = $2
	`

	var lastTime *time.Time
	err := dbpool.QueryRow(ctx, query, figi, intervalType).Scan(&lastTime)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return time.Time{}, nil // Нет данных
		}
		return time.Time{}, fmt.Errorf("ошибка получения времени последней свечи: %w", err)
	}

	if lastTime == nil {
		return time.Time{}, nil // Нет данных
	}

	return *lastTime, nil
}

// SaveCandles сохраняет свечи в базу данных батчами (с логгером)
func SaveCandles(dbpool *pgxpool.Pool, figi string, candles []*pb.HistoricCandle, intervalType string, logger *logrus.Logger) error {
	if len(candles) == 0 {
		return nil
	}

	//	const batchSize = 1000 // Размер батча

	// Логируем начало сохранения
	// logger.Debugf("Начинаем сохранение %d свечей батчами", len(candles))
	logger.Debugf("Начинаем сохранение %d свечей", len(candles))

	// Подготавливаем запрос
	query := `
		INSERT INTO candles (figi, time, open_price, high_price, low_price, close_price, volume, interval_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (figi, time, interval_type) DO UPDATE SET
			open_price = EXCLUDED.open_price,
			high_price = EXCLUDED.high_price,
			low_price = EXCLUDED.low_price,
			close_price = EXCLUDED.close_price,
			volume = EXCLUDED.volume
	`

	// Обрабатываем свечи батчами
	//	totalBatches := (len(candles) + batchSize - 1) / batchSize
	//	for i := 0; i < len(candles); i += batchSize {
	for _, candle := range candles {
		//		end := i + batchSize
		//		if end > len(candles) {
		//			end = len(candles)
		//		}
		//
		//		batch := candles[i:end]
		//		batchNum := (i / batchSize) + 1
		//
		//		logger.Debugf("Обрабатываем батч %d/%d (%d свечей)...", batchNum, totalBatches, len(batch))

		// Начинаем транзакцию для батча
		//		tx, err := dbpool.Begin(context.Background())
		//		if err != nil {
		//			return fmt.Errorf("ошибка начала транзакции для батча %d-%d: %w", i, end, err)
		//		}

		// Выполняем вставку батча
		//		for _, candle := range batch {
		//_, err := tx.Exec(context.Background(), query,
		_, err := dbpool.Exec(context.Background(), query,
			figi,
			candle.GetTime().AsTime(),
			money.ConvertMoneyValue(candle.GetOpen().GetUnits(), candle.GetOpen().GetNano()),
			money.ConvertMoneyValue(candle.GetHigh().GetUnits(), candle.GetHigh().GetNano()),
			money.ConvertMoneyValue(candle.GetLow().GetUnits(), candle.GetLow().GetNano()),
			money.ConvertMoneyValue(candle.GetClose().GetUnits(), candle.GetClose().GetNano()),
			candle.GetVolume(),
			intervalType,
		)

		if err != nil {
			// Проверяем, является ли ошибка связанной с отсутствием партиции
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				// Проверяем код ошибки
				switch {
				case pgErr.Code == "23514":
					logger.Debugf("Обнаружена ошибка отсутствия партиции (код 23514) для времени %s", candle.GetTime().AsTime().Format("2006-01-02"))
				case strings.Contains(pgErr.Message, "no partition of relation"):
					logger.Debugf("Обнаружена ошибка отсутствия партиции (английское сообщение) для времени %s", candle.GetTime().AsTime().Format("2006-01-02"))
				case strings.Contains(pgErr.Message, "для строки не найдена секция"):
					logger.Debugf("Обнаружена ошибка отсутствия партиции (русское сообщение) для времени %s", candle.GetTime().AsTime().Format("2006-01-02"))
				case strings.Contains(pgErr.Message, "partition"):
					logger.Debugf("Обнаружена ошибка партиции (общее сообщение) для времени %s", candle.GetTime().AsTime().Format("2006-01-02"))
				default:
					// Это не ошибка партиции - откатываем транзакцию и возвращаем ошибку
					//		if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
					//					logger.Errorf("Ошибка отката транзакции: %v", rollbackErr)
					//				}
					return fmt.Errorf("ошибка вставки свечи: %w", err)
				}

				// Если это ошибка партиции - обрабатываем её
				logger.Debugf("Создаем партицию для времени %s...", candle.GetTime().AsTime().Format("2006-01-02"))

				// Подтверждаем текущую транзакцию перед созданием партиции
				//			if commitErr := tx.Commit(context.Background()); commitErr != nil {
				//
				//				return fmt.Errorf("ошибка подтверждения транзакции перед созданием партиции: %w", commitErr)
				//			}

				// Создаем партицию
				if createErr := CreatePartition(dbpool, candle.GetTime().AsTime()); createErr != nil {
					return fmt.Errorf("ошибка создания партиции: %w", createErr)
				}

				// Начинаем новую транзакцию для повторной вставки
				//			tx, err = dbpool.Begin(context.Background())
				//			if err != nil {
				//				return fmt.Errorf("ошибка начала новой транзакции после создания партиции: %w", err)
				//			}

				// Повторяем вставку этой свечи
				//		_, retryErr := tx.Exec(context.Background(), query,
				_, retryErr := dbpool.Exec(context.Background(), query,
					figi,
					candle.GetTime().AsTime(),
					money.ConvertMoneyValue(candle.GetOpen().GetUnits(), candle.GetOpen().GetNano()),
					money.ConvertMoneyValue(candle.GetHigh().GetUnits(), candle.GetHigh().GetNano()),
					money.ConvertMoneyValue(candle.GetLow().GetUnits(), candle.GetLow().GetNano()),
					money.ConvertMoneyValue(candle.GetClose().GetUnits(), candle.GetClose().GetNano()),
					candle.GetVolume(),
					intervalType,
				)
				if retryErr != nil {
					//			if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
					//				logger.Errorf("Ошибка отката транзакции после создания партиции: %v", rollbackErr)
					//			}
					return fmt.Errorf("ошибка вставки свечи после создания партиции: %w", retryErr)
				}

				continue
			}

			// Если это не PostgreSQL ошибка - откатываем транзакцию и возвращаем ошибку
			//		if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
			//			logger.Errorf("Ошибка отката транзакции: %v", rollbackErr)
			//		}
			return fmt.Errorf("ошибка вставки свечи: %w", err)
		}
		//		}

		// Подтверждаем транзакцию батча
		//	if err := tx.Commit(context.Background()); err != nil {
		//		return fmt.Errorf("ошибка подтверждения транзакции для батча %d-%d: %w", i, end, err)
		//	}
	}

	return nil
}
