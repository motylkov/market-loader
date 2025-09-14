// Package data - Запросы в API и обработка данных
// Market Loader
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	"github.com/sirupsen/logrus"

	"market-loader/internal/storage"
	"market-loader/pkg/config"
)

// LoadCandleData универсальная функция для загрузки данных свечей
func LoadCandleData(
	ctx context.Context,
	client *investgo.Client,
	dbpool *pgxpool.Pool,
	instrument storage.Instrument,
	lastLoadedTime time.Time,
	intervalType string,
	cfg *config.Config,
	logger *logrus.Logger,
) error {
	// Проверяем, нужно ли обновлять данные
	if !lastLoadedTime.IsZero() && !config.ShouldUpdateData(lastLoadedTime, intervalType) {
		logger.WithFields(logrus.Fields{
			"figi":   instrument.Figi,
			"ticker": instrument.Ticker,
		}).Debug("Данные актуальны, пропускаем")
		return nil
	}

	// Определяем единицу времени и ключ конфигурации по типу интервала
	timeUnit, configKey := config.GetTimeUnitAndConfigKey(intervalType)

	// Рассчитываем размер чанка
	chunkSize := time.Duration(cfg.GetIntervalLimit(configKey)) * timeUnit

	// Определяем период загрузки
	var from time.Time
	if lastLoadedTime.IsZero() {
		// Новый инструмент - загружаем полную историю
		from = cfg.GetStartDate()
	} else {
		// Существующий инструмент - обновляем с последней свечи
		from = lastLoadedTime
	}
	to := time.Now()

	// Определяем формат даты для логирования
	dateFormat := config.GetDateFormat(intervalType)

	// Формируем дополнительные поля для логов в зависимости от типа интервала
	logFields := logrus.Fields{
		"figi":      instrument.Figi,
		"ticker":    instrument.Ticker,
		"isin":      instrument.Isin,
		"startTime": from.Format("2006-01-02"),
		"endTime":   to.Format("2006-01-02"),
		"apiLimit":  cfg.GetIntervalLimit(configKey),
		"chunkSize": chunkSize,
	}

	// Добавляем специфичные поля для разных типов интервалов
	switch intervalType {
	case config.CandleIntervalDay:
		logFields["chunkSizeDays"] = chunkSize.Hours() / config.HoursInDay
	case config.CandleIntervalWeek:
		logFields["chunkSizeWeeks"] = chunkSize.Hours() / (config.HoursInDay * config.DaysInWeek)
	case config.CandleIntervalMonth:
		logFields["chunkSizeMonths"] = chunkSize.Hours() / (config.HoursInDay * config.DaysInMonth)
	default:
		logFields["chunkSizeHours"] = chunkSize.Hours()
	}

	// Определяем тип операции для логирования
	operationType := "обновляем данные"
	if lastLoadedTime.IsZero() {
		operationType = "загружаем полную историю"
	}
	logFields["operation"] = operationType

	logger.WithFields(logFields).Info("Загружаем данные с разбивкой по лимитам API")

	// Загружаем данные чанками
	totalCandles := 0
	currentFrom := from

	for currentFrom.Before(to) {
		currentTo := currentFrom.Add(chunkSize)
		if currentTo.After(to) {
			currentTo = to
		}

		logger.WithFields(logrus.Fields{
			"figi":      instrument.Figi,
			"ticker":    instrument.Ticker,
			"isin":      instrument.Isin,
			"chunkFrom": currentFrom.Format(dateFormat),
			"chunkTo":   currentTo.Format(dateFormat),
		}).Info("Загружаем чанк")

		// Загружаем чанк данных
		candles, err := LoadCandleChunk(ctx, client, instrument.Figi, currentFrom, currentTo, config.GetCandleInterval(intervalType))
		if err != nil {
			return fmt.Errorf("ошибка загрузки чанка %s - %s: %w",
				currentFrom.Format("2006-01-02"), currentTo.Format("2006-01-02"), err)
		}

		// Сохраняем чанк в БД
		if len(candles) > 0 {
			if err := storage.SaveCandles(dbpool, instrument.Figi, candles, intervalType, logger); err != nil {
				return fmt.Errorf("ошибка сохранения чанка: %w", err)
			}

			totalCandles += len(candles)
			logger.WithFields(logrus.Fields{
				"figi":      instrument.Figi,
				"ticker":    instrument.Ticker,
				"isin":      instrument.Isin,
				"chunkSize": len(candles),
				"total":     totalCandles,
			}).Info("Чанк сохранен")
		}

		// Переходим к следующему чанку
		currentFrom = currentTo

		// Пауза между запросами согласно конфигурации
		time.Sleep(time.Duration(cfg.Loading.RateLimitPause) * time.Second)
	}

	// Определяем сообщение завершения
	completionMessage := "Данные обновлены"
	if lastLoadedTime.IsZero() {
		completionMessage = "Полная история загружена"
	}

	logger.WithFields(logrus.Fields{
		"figi":         instrument.Figi,
		"ticker":       instrument.Ticker,
		"isin":         instrument.Isin,
		"totalCandles": totalCandles,
	}).Info(completionMessage)

	return nil
}

// ProcessLoadResult обрабатывает результат загрузки данных
func ProcessLoadResult(
	ctx context.Context,
	dbpool *pgxpool.Pool,
	figi, intervalType string,
	loadError error,
	logger *logrus.Logger,
) error {
	// Получаем время последней загруженной свечи из БД
	lastCandleTime, err := storage.GetLastCandleTime(ctx, dbpool, figi, intervalType)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"figi":         figi,
			"intervalType": intervalType,
			"error":        err,
		}).Warn("Не удалось получить время последней свечи для обновления прогресса")
		return loadError // Возвращаем исходную ошибку
	}

	// Если есть свечи в БД, обновляем время последней загрузки
	if !lastCandleTime.IsZero() {
		if err := storage.UpdateLastLoadedTime(ctx, dbpool, figi, lastCandleTime); err != nil {
			logger.WithFields(logrus.Fields{
				"figi":           figi,
				"intervalType":   intervalType,
				"lastCandleTime": lastCandleTime,
				"error":          err,
			}).Warn("Не удалось обновить время последней загрузки")
		} else {
			logger.WithFields(logrus.Fields{
				"figi":           figi,
				"intervalType":   intervalType,
				"lastCandleTime": lastCandleTime,
			}).Info("Обновлено время последней загрузки на основе последней свечи")
		}
	}

	// Возвращаем исходную ошибку загрузки (если была)
	return loadError
}
