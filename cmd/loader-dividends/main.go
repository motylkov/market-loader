// Package main содержит загрузчик дивидендов
// Market Loader
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
	"market-loader/internal/app"
	"market-loader/pkg/config"
	"market-loader/pkg/logs"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	// Определяем путь к конфигурации
	configPath := config.GetConfigPath()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Настраиваем логирование
	logger := logs.SetupLogger(cfg)

	logger.Info("Запуск загрузчика дивидендов")

	// Проверяем валидность даты начала загрузки
	startDate := cfg.GetStartDate()
	if startDate.After(time.Now()) {
		logger.Fatalf("Дата начала загрузки (%s) не может быть в будущем", startDate.Format("2006-01-02"))
	}

	// Логируем настройки лимитов
	if cfg.Loading.RateLimitPause > 0 {
		logger.Debugf("Установлена пауза между запросами: %d секунд (API limit)", cfg.Loading.RateLimitPause)
	} else {
		logger.Debug("Пауза между запросами не установлена (API limit)")
	}

	// Создаем контекст
	ctx := context.Background()

	// Подключение и получение исходных данных
	instance, err := app.Initialize(ctx, cfg, startDate, logger, "instruments")
	if err != nil {
		logger.Fatalf("Ошибка инициализации: %v", err)
	}
	defer instance.DBPool.Close()

	logger.WithField("count", len(instance.Instruments)).Debug("Количество инструментов в БД")

	var shareCount = 0
	// Обрабатываем каждый инструмент
	for _, instrument := range instance.Instruments {
		// Обрабатываем только активные (enabled=true) акции
		if instrument.InstrumentType == config.Shares && instrument.Enabled {
			logger.WithFields(logrus.Fields{
				"figi":   instrument.Figi,
				"ticker": instrument.Ticker,
				"name":   instrument.Name,
			}).Debug("Обработка дивидендов инструмента")
			if err := app.ProcessInstrumentDividends(ctx, instance.Client, instance.DBPool, instrument, cfg, logger); err != nil {
				logger.WithFields(logrus.Fields{
					"figi":   instrument.Figi,
					"ticker": instrument.Ticker,
					"name":   instrument.Name,
					"error":  err,
				}).Error("Ошибка обработки дивидендов инструмента")
				continue
			}

			// Пауза между запросами
			time.Sleep(time.Duration(cfg.Loading.RateLimitPause) * time.Second)

			shareCount++
		}
	}
	logger.Debugf("Обработано акций %d", shareCount)

	logger.Info("Загрузка дивидендов завершена")
}
