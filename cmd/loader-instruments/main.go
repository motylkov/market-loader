// Package main содержит загрузчик инструментов из API
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

	logger.Info("Запуск загрузчика инструментов")

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

	logger.WithField("count", len(instance.Instruments)).Debug("Количество активных (enabled=true) инструментов в БД")

	// Загружаем все типы инструментов из API
	logger.Debug("Загружаем все инструменты из API и обновляем в БД")
	if err := app.LoadAllInstruments(ctx, instance.Client, instance.DBPool, logger); err != nil {
		logger.Fatalf("Ошибка загрузки инструментов из API: %v", err)
	}
}
