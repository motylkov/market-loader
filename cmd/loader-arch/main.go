// Package main содержит загрузчик минутных свечей
// Используется загрузка сжатых архивных данных, работает быстрее
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
	"market-loader/internal/app"
	"market-loader/internal/arch"
	"market-loader/internal/storage"
	"market-loader/pkg/config"
	"market-loader/pkg/logs"
	"os"
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

	logger.Info("Запуск загрузчика минутных данных через архивы")

	// Логируем настройки лимитов
	if cfg.Loading.RateLimitPause > 0 {
		logger.Debugf("Установлена пауза между запросами: %d секунд (API limit)", cfg.Loading.RateLimitPause)
	} else {
		logger.Debug("Пауза между запросами не установлена (API limit)")
	}

	// Определяем год начала загрузки
	startDate := cfg.GetStartDate()
	var startYear int
	if cfg.Loading.StartDate != "" {
		startYear = startDate.Year()
		logger.WithField("startYear", startYear).Debug("Год начала загрузки данных")
	} else {
		startYear = time.Now().Year() - config.DefaultYearsBack // По умолчанию 5 лет назад
		logger.WithField("startYear", startYear).Debug("Используем год начала загрузки данных по умолчанию (now - 5)")
	}

	currentYear := time.Now().Year()
	logger.Infof("Загрузка данных с %d по %d год (всего %d лет)", startYear, currentYear, currentYear-startYear+1)

	// Создаем контекст
	ctx := context.Background()

	// Подключение и получение исходных данных
	instance, err := app.Initialize(ctx, cfg, startDate, logger, "instruments")
	if err != nil {
		logger.Fatalf("Ошибка инициализации: %v", err)
	}
	defer instance.DBPool.Close()

	logger.WithField("count", len(instance.Instruments)).Debug("Количество активных (enabled=true) инструментов в БД")

	// Определяем временную директорию для архивов
	var tempDir string
	if cfg.Archive.TempDir != "" {
		// Используем настроенную директорию
		tempDir = cfg.Archive.TempDir
		// Создаем директорию, если она не существует
		if err := os.MkdirAll(tempDir, config.DefaultDirPerm); err != nil {
			logger.Fatalf("Ошибка создания временной директории %s: %v", tempDir, err)
		}
	} else {
		// Используем системную временную директорию
		var err error
		tempDir, err = os.MkdirTemp("", "tinvest_archives")
		if err != nil {
			logger.Fatalf("Ошибка создания временной директории: %v", err)
		}
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				logger.Errorf("Ошибка удаления временной директории: %v", err)
			}
		}()
	}

	// Загружаем данные по каждому инструменту
	totalCandles := 0
	requestCount := 0

	for _, instrument := range instance.Instruments {
		logger.Infof("Загрузка данных для %s (%s)", instrument.Ticker, instrument.Figi)

		instrumentCandles := 0
		for year := startYear; year <= currentYear; year++ {
			// Создаем партиции для года заранее
			logger.Infof("Создание партиций для %d года...", year)
			if err := storage.CreateYearPartitions(instance.DBPool, year); err != nil {
				logger.Warnf("Ошибка создания партиций за %d год для %s: %v", year, instrument.Ticker, err)
				continue
			}

			// Проверяем лимиты API
			if cfg.Loading.RateLimitPause > 0 {
				logger.Infof("Пауза %d секунд для соблюдения лимитов API...", cfg.Loading.RateLimitPause)
				time.Sleep(time.Duration(cfg.Loading.RateLimitPause) * time.Second)
			}

			candles, err := arch.DownloadYearArchive(ctx, cfg.Tinvest.Token, instrument.Figi, year, tempDir, instance.DBPool, logger)
			if err != nil {
				logger.Warnf("Ошибка загрузки архива за %d год для %s: %v", year, instrument.Ticker, err)
				continue
			}

			requestCount++

			instrumentCandles += len(candles)
			logger.Infof("Загружено %d свечей за %d год для %s (запросов: %d)", len(candles), year, instrument.Ticker, requestCount)
		}

		totalCandles += instrumentCandles
		logger.Infof("Всего загружено %d свечей для %s", instrumentCandles, instrument.Ticker)
	}

	logger.Infof("Загрузка завершена. Всего загружено %d свечей", totalCandles)
}
