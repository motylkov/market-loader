// Package main содержит загрузчик свечей из API
// из данного файла мы компилируем все интервальные загрузчики
// подставляя значение интервала MAININTERVAL при сборке
//
// # Market Loader
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"context"
	"log"
	"time"

	"market-loader/internal/app"
	"market-loader/pkg/config"
	"market-loader/pkg/logs"

	"github.com/sirupsen/logrus"
)

var MAININTERVAL string

func main() {
	if MAININTERVAL == "" {
		log.Println("MAININTERVAL не задан при сборке (или произошла ошибка)")
		log.Println("Используйте Makefile для корректной сборки")
		log.Println("По умолчанию используется интервал 1 минута")
		MAININTERVAL = config.CandleInterval1Min
	}

	// Определяем путь к конфигурации
	configPath := config.GetConfigPath()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Проверяем валидность даты начала загрузки
	startDate := cfg.GetStartDate()
	if startDate.After(time.Now()) {
		log.Fatalf("Дата начала загрузки (%s) не может быть в будущем", startDate)
	}

	// Настраиваем логирование
	logger := logs.SetupLogger(cfg)

	logger.Infof("Запуск загрузчика данных на интервал %s", config.Interval2text(MAININTERVAL))

	// Логируем настройки загрузки
	logger.WithFields(logrus.Fields{
		"startDate":      cfg.GetStartDate().Format("2006-01-02"),
		"rateLimitPause": cfg.Loading.RateLimitPause,
		"apiLimit":       cfg.GetIntervalLimit(config.Interval2text(MAININTERVAL)),
	}).Info("Настройки загрузки")

	// Создаем контекст
	ctx := context.Background()

	// Подключение и получение исходных данных
	instance, err := app.Initialize(ctx, cfg, startDate, logger, config.Interval2text(MAININTERVAL))
	if err != nil {
		logger.Fatalf("Ошибка инициализации: %v", err)
	}
	defer instance.DBPool.Close()

	logger.WithField("count", len(instance.Instruments)).Debug("Количество инструментов в БД")

	// Обрабатываем каждый инструмент
	for _, instrument := range instance.Instruments {
		if err := app.ProcessInstrument(ctx, instance.Client, instance.DBPool, MAININTERVAL, instrument, cfg, logger); err != nil {
			logger.WithFields(logrus.Fields{
				"figi":   instrument.Figi,
				"ticker": instrument.Ticker,
				"error":  err,
			}).Error("Ошибка обработки инструмента")
			continue
		}

		// Пауза между запросами
		time.Sleep(time.Duration(cfg.Loading.RateLimitPause) * time.Second)
	}

	logger.Info("Загрузка завершена")
}
