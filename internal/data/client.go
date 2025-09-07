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
	"market-loader/pkg/config"

	// "market-loader/pkg/mainlib"

	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	"github.com/sirupsen/logrus"
)

// CreateTinvestClient создает клиент для работы с T-Invest API
func CreateTinvestClient(ctx context.Context, cfg *config.Config) (*investgo.Client, error) {
	config := investgo.Config{
		EndPoint: cfg.Tinvest.Endpoint,
		Token:    cfg.Tinvest.Token,
		AppName:  cfg.Tinvest.AppName,
	}

	// Создаем простой логгер для SDK
	sdkLogger := logrus.New()
	sdkLogger.SetLevel(logrus.WarnLevel) // Минимальное логирование от SDK

	client, err := investgo.NewClient(ctx, config, sdkLogger)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента: %w", err)
	}

	return client, nil
}
