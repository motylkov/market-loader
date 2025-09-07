// Package app - основные функции загрузчиков
// Market Loader
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package app

import (
	"context"
	"fmt"
	"market-loader/internal/data"
	"market-loader/internal/storage"
	"market-loader/pkg/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	"github.com/sirupsen/logrus"
)

// ProcessInstrumentDividends обрабатывает дивиденды одного инструмента
func ProcessInstrumentDividends(ctx context.Context, client *investgo.Client, dbpool *pgxpool.Pool, instrument storage.Instrument, cfg *config.Config, logger *logrus.Logger) error {
	// Проверяем последнюю дату выплаты дивидендов
	lastDividendDate, _ := storage.GetLastDividendDate(ctx, dbpool, instrument.Figi)

	// Определяем период загрузки
	endTime := time.Now()
	startTime := cfg.GetStartDate()

	// Если есть последняя выплата, начинаем с неё
	if !lastDividendDate.IsZero() {
		startTime = lastDividendDate.AddDate(0, 0, 1) // Следующий день после последней выплаты
	}

	// Проверяем, нужно ли загружать данные
	if startTime.After(endTime) {
		logger.WithFields(logrus.Fields{
			"figi":   instrument.Figi,
			"ticker": instrument.Ticker,
		}).Debug("Дивиденды актуальны, пропускаем")
		return nil
	}

	logger.WithFields(logrus.Fields{
		"figi":      instrument.Figi,
		"ticker":    instrument.Ticker,
		"startTime": startTime.Format("2006-01-02"),
		"endTime":   endTime.Format("2006-01-02"),
	}).Info("Загружаем дивиденды")

	// Загружаем дивиденды
	dividends, err := data.LoadDividends(client, instrument.Figi, startTime, endTime)
	if err != nil {
		return fmt.Errorf("ошибка загрузки дивидендов: %w", err)
	}

	// Сохраняем дивиденды
	if len(dividends) > 0 {
		for _, dividend := range dividends {
			if err := storage.SaveDividend(ctx, dbpool, dividend); err != nil {
				return fmt.Errorf("ошибка сохранения дивиденда: %w", err)
			}
		}

		logger.WithFields(logrus.Fields{
			"figi":   instrument.Figi,
			"ticker": instrument.Ticker,
			"count":  len(dividends),
		}).Info("Дивиденды сохранены")
	} else {
		logger.WithFields(logrus.Fields{
			"figi":   instrument.Figi,
			"ticker": instrument.Ticker,
		}).Debug("Новых дивидендов нет")
	}

	return nil
}
