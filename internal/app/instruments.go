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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	"github.com/sirupsen/logrus"
)

// LoadAllInstruments загружает все типы инструментов
func LoadAllInstruments(
	ctx context.Context,
	client *investgo.Client,
	dbpool *pgxpool.Pool,
	logger *logrus.Logger,
) error {
	// Получаем или создаем источник данных T-Invest
	dataSourceID, err := data.GetOrCreateTInvestDataSource(ctx, dbpool)
	if err != nil {
		return fmt.Errorf("ошибка получения источника данных T-Invest: %w", err)
	}

	// Загружаем акции
	logger.Debug("Загружаем акции...")
	if err := data.LoadInstrumentsByType(ctx, client, dbpool, "share", dataSourceID, logger); err != nil {
		return fmt.Errorf("ошибка загрузки share: %w", err)
	}

	// Загружаем облигации
	logger.Debug("Загружаем облигации...")
	if err := data.LoadInstrumentsByType(ctx, client, dbpool, "bond", dataSourceID, logger); err != nil {
		return fmt.Errorf("ошибка загрузки bond: %w", err)
	}

	// Загружаем ETF
	logger.Debug("Загружаем ETF...")
	if err := data.LoadInstrumentsByType(ctx, client, dbpool, "etf", dataSourceID, logger); err != nil {
		return fmt.Errorf("ошибка загрузки etf: %w", err)
	}

	logger.Info("Все инструменты (share, bond, etf) загружены с расширенными данными")

	return nil
}
