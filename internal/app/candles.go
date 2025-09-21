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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	"github.com/sirupsen/logrus"
)

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

// ProcessInstrument обрабатывает один инструмент
//
//nolint:wrapcheck
func ProcessInstrument(
	ctx context.Context,
	client *investgo.Client,
	dbpool *pgxpool.Pool,
	interval string,
	instrument storage.Instrument,
	cfg *config.Config,
	logger *logrus.Logger,
) error {
	// Проверяем статус загрузки по реально загруженным данным
	lastLoadedTime, err := storage.GetLastLoadedTime(ctx, dbpool, instrument.Figi, interval)
	if err != nil {
		return fmt.Errorf("ошибка получения времени последней загрузки: %w", err)
	}

	// Загружаем данные с помощью универсальной функции
	loadError := data.LoadCandleData(ctx, client, dbpool, instrument, lastLoadedTime, interval, cfg, logger)

	// Обрабатываем результат загрузки и обновляем прогресс
	return data.ProcessLoadResult(ctx, dbpool, instrument.Figi, interval, loadError, logger)
}
