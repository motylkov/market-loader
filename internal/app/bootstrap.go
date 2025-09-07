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
	"time"

	"market-loader/internal/data"
	"market-loader/internal/storage"
	"market-loader/pkg/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	"github.com/sirupsen/logrus"
)

// Result — структура для загурзчиков
type Result struct {
	Ctx         context.Context
	DBPool      *pgxpool.Pool
	Client      *investgo.Client
	Instruments []storage.Instrument
	StartDate   time.Time
	Logger      *logrus.Entry
}

// Initialize — централизованная инициализация для загрузчиков
// loaderName используется как имя и интервал
func Initialize(
	ctx context.Context,
	cfg *config.Config,
	startDate time.Time,
	logger *logrus.Logger,
	loaderName string,
) (*Result, error) {
	log := logger.WithField("loader", loaderName)
	log.Debug("Начало инициализации компонентов")

	// Подключение к БД
	dbpool, err := storage.ConnectToDatabase(ctx, &cfg.Database)
	if err != nil {
		return nil, &InitializationError{Msg: "ошибка подключения к БД", Err: err}
	}

	// Клиент API
	client, err := data.CreateTinvestClient(ctx, cfg)
	if err != nil {
		dbpool.Close()
		return nil, &InitializationError{Msg: "ошибка создания клиента API", Err: err}
	}

	// Загрузка инструментов
	instruments, err := storage.LoadInstruments(ctx, dbpool, logger)
	if err != nil {
		dbpool.Close()
		return nil, &InitializationError{Msg: "ошибка загрузки инструментов", Err: err}
	}

	log.WithField("count", len(instruments)).Debug("Инструменты загружены")

	return &Result{
		Ctx:         ctx,
		DBPool:      dbpool,
		Client:      client,
		Instruments: instruments,
		StartDate:   startDate,
		Logger:      log,
	}, nil
}

// InitializationError — кастомная ошибка для диагностики
type InitializationError struct {
	Msg   string
	Field string
	Err   error
}

func (e *InitializationError) Error() string {
	msg := "bootstrap: " + e.Msg
	if e.Field != "" {
		msg += " (" + e.Field + ")"
	}
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}
