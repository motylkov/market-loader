// Package main содержит CLI загрузчик свечей с возможностью переопределения параметров
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
	"fmt"
	"log"
	"market-loader/internal/app"
	"market-loader/internal/storage"
	"market-loader/pkg/config"
	"market-loader/pkg/logs"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Флаги командной строки
	interval   string
	figi       string
	startDate  string
	configPath string

	// Корневая команда
	rootCmd = &cobra.Command{
		Use:   "t-loader_cli",
		Short: "CLI загрузчик свечей",
		Long: `CLI загрузчик свечей с возможностью переопределения параметров конфигурации.

Примеры использования:
  t-loader_cli --figi BBG000B9XRY4 --interval 1min
  t-loader_cli --figi BBG000B9XRY4 --interval 1hour --start-date 2024-01-01
  t-loader_cli --figi BBG000B9XRY4 --interval 1day --start-date 2024-01-01 --debug`,
		RunE: runLoader,
	}
)

func runLoader(cmd *cobra.Command, _ []string) error {
	// Определяем путь к конфигурации
	if !cmd.Flags().Changed("config") {
		configPath = config.GetConfigPath()
	}

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Настраиваем логирование
	logger := logs.SetupLogger(cfg)

	logger.Info("Запуск CLI загрузчика свечей")

	// Определяем интервал
	// Выходим если не задан
	intervalType, err := config.ParseInterval(interval)
	if err != nil {
		logger.Fatalf("Ошибка парсинга интервала: %v", err)
	}

	// Читаем дату из конфига если нет параметра
	if !cmd.Flags().Changed("start-date") {
		startDate = cfg.Loading.StartDate
	}
	// Проверяем валидность даты начала загрузки
	parsedTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		logger.Fatalf("Ошибка парсинга интервала: %v", err)
	}
	if parsedTime.After(time.Now()) {
		logger.Fatalf("Дата начала загрузки (%s) не может быть в будущем", startDate)
	} else {
		cfg.Loading.StartDate = parsedTime.Format("2006-01-02")
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
	instance, err := app.Initialize(ctx, cfg, parsedTime, logger, config.Interval2text(intervalType))
	if err != nil {
		logger.Fatalf("Ошибка инициализации: %v", err)
	}
	defer instance.DBPool.Close()

	logger.WithField("count", len(instance.Instruments)).Debug("Количество инструментов в БД")

	var instruments []storage.Instrument
	if cmd.Flags().Changed("figi") {
		// Получаем инструмент из базы данных или API
		instr, err := getInstrument(ctx, instance, figi, logger)
		if err != nil {
			logger.Fatalf("Ошибка получения инструмента: %v", err)
		}
		instruments = append(instruments, *instr)
	} else {
		instruments = instance.Instruments
	}

	logger.Infof("Запуск загрузчика данных на интервал %s", config.Interval2text(intervalType))

	// Логируем настройки загрузки
	logger.WithFields(logrus.Fields{
		"startDate":      cfg.GetStartDate().Format("2006-01-02"),
		"rateLimitPause": cfg.Loading.RateLimitPause,
		"apiLimit":       cfg.GetIntervalLimit(config.Interval2text(intervalType)),
	}).Info("Настройки загрузки")

	// Обрабатываем инструменты
	for _, instrument := range instruments {
		if err := app.ProcessInstrument(ctx, instance.Client, instance.DBPool, intervalType, instrument, cfg, logger); err != nil {
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

	return nil
}

func getInstrument(ctx context.Context, instance *app.Result, figi string, logger *logrus.Logger) (*storage.Instrument, error) {
	// Ищем инструмент по FIGI
	for _, instrument := range instance.Instruments {
		if instrument.Figi == figi {
			logger.Infof("Инструмент найден в базе данных: %s (%s)", instrument.Name, instrument.Figi)
			return &instrument, nil
		}
	}

	// Если не найден в базе, получаем из API
	logger.Infof("Инструмент не найден в базе данных, получаем из API: %s", figi)
	if err := app.LoadAllInstruments(ctx, instance.Client, instance.DBPool, logger); err != nil {
		logger.Fatalf("Ошибка загрузки инструментов из API: %v", err)
	}
	newInstruments, err := storage.GetInstruments(ctx, instance.DBPool, "")
	if err != nil {
		logger.Errorf("Ошибка загрузки инструментов из API: %v", err)
	} else {
		for _, instrument := range newInstruments {
			if instrument.Figi == figi {
				logger.Infof("Инструмент найден в базе данных: %s (%s)", instrument.Name, instrument.Figi)
				return &instrument, nil
			}
		}
	}

	return nil, fmt.Errorf("инструмент с FIGI %s не найден", figi)
}

func main() {
	// Добавляем флаги
	rootCmd.Flags().StringVarP(&interval, "interval", "i", "1min", "Интервал свечей (1min, 2min, 3min, 5min, 10min, 15min, 30min, 1hour, 2hour, 4hour, 1day, 1week, 1month)")
	rootCmd.Flags().StringVarP(&figi, "figi", "f", "", "FIGI инструмента (по умолчанию enabled=true из БД)")
	rootCmd.Flags().StringVarP(&startDate, "start-date", "s", "", "Дата начала загрузки в формате YYYY-MM-DD (по умолчанию из конфига)")
	rootCmd.Flags().StringVarP(&configPath, "conf", "c", "config/config.yaml", "Путь к файлу конфигурации (опционально)")

	// Делаем --interval обязательным
	if err := rootCmd.MarkFlagRequired("interval"); err != nil {
		log.Fatalf("%v", err)
	}

	// Выполняем команду
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка выполнения команды: %v\n", err)
		os.Exit(1)
	}
}
