// Package arch содержит функции для работы с архивом свечей
// Market Loader
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package arch

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"market-loader/internal/storage"
	"market-loader/pkg/config"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
)

// processArchive обрабатывает ZIP архив и извлекает данные свечей
func processArchive(archivePath, figi string, dbpool *pgxpool.Pool, logger *logrus.Logger) ([]*pb.HistoricCandle, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия архива: %w", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Errorf("Ошибка закрытия архива: %v", err)
		}
	}()

	var candles []*pb.HistoricCandle
	logger.Debugf("Открыт архив: %s, файлов: %d", archivePath, len(reader.File))

	// Ищем CSV файлы в архиве
	csvFileCount := 0
	for _, file := range reader.File {
		logger.Debugf("Файл в архиве: %s, размер: %d", file.Name, file.UncompressedSize64)

		if !strings.HasSuffix(file.Name, ".csv") {
			logger.Debugf("Пропускаем файл (не CSV): %s", file.Name)
			continue
		}

		csvFileCount++
		logger.Debugf("Обрабатываем CSV файл %d: %s", csvFileCount, file.Name)

		// Открываем CSV файл
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("ошибка открытия файла в архиве: %w", err)
		}

		// Парсим CSV
		csvReader := csv.NewReader(rc)
		csvReader.Comma = ';' // T-Invest использует точку с запятой как разделитель

		// Заголовка нет, сразу читаем данные
		rowCount := 0
		var firstTime, lastTime time.Time
		var fileCandles []*pb.HistoricCandle

		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger.Warnf("Ошибка чтения строки %d: %v", rowCount+1, err)
				continue
			}

			rowCount++

			// Парсим строку: UID, UTC, open, close, high, low, volume
			if len(record) < config.MinCSVFields {
				logger.Debugf("Строка %d: недостаточно полей (%d), пропускаем", rowCount, len(record))
				continue
			}

			// Парсим время (формат ISO 8601: 2024-12-19T04:00:00Z)
			timestamp, err := time.Parse("2006-01-02T15:04:05Z", record[1])
			if err != nil {
				logger.Debugf("Строка %d: ошибка парсинга времени '%s': %v", rowCount, record[1], err)
				continue
			}

			// Запоминаем первое и последнее время
			if rowCount == 1 {
				firstTime = timestamp
			}
			lastTime = timestamp

			// Парсим цены как строки для точного преобразования
			openStr := strings.TrimSpace(record[2])
			closeStr := strings.TrimSpace(record[3])
			highStr := strings.TrimSpace(record[4])
			lowStr := strings.TrimSpace(record[5])

			volume, err := strconv.ParseInt(record[6], 10, 64)
			if err != nil {
				logger.Debugf("Строка %d: ошибка парсинга volume '%s': %v", rowCount, record[6], err)
				continue
			}

			// Создаем protobuf структуру с точным парсингом цен
			candle := &pb.HistoricCandle{
				Time:   timestamppb.New(timestamp),
				Open:   parsePriceString(openStr),
				High:   parsePriceString(highStr),
				Low:    parsePriceString(lowStr),
				Close:  parsePriceString(closeStr),
				Volume: volume,
			}

			fileCandles = append(fileCandles, candle)
		}

		logger.Debugf("Обработано строк: %d, создано свечей: %d", rowCount, len(fileCandles))
		if rowCount > 0 {
			logger.Debugf("Временной диапазон: %s - %s (длительность: %v)",
				firstTime.Format("2006-01-02 15:04:05"),
				lastTime.Format("2006-01-02 15:04:05"),
				lastTime.Sub(firstTime))
		}
		if err := rc.Close(); err != nil {
			logger.Errorf("Ошибка закрытия файла в архиве: %v", err)
		}

		// Сохраняем свечи из этого файла сразу
		if len(fileCandles) > 0 {
			logger.Debugf("Сохраняем %d свечей из файла %s...", len(fileCandles), file.Name)
			if err := storage.SaveCandles(dbpool, figi, fileCandles, config.CandleInterval1Min, logger); err != nil {
				logger.Warnf("Ошибка сохранения свечей из файла %s: %v", file.Name, err)
				continue
			}
			logger.Debugf("Успешно сохранено %d свечей из файла %s", len(fileCandles), file.Name)
		}

		// Добавляем свечи из файла к общему результату
		candles = append(candles, fileCandles...)
		// Продолжаем обработку всех CSV файлов в архиве
	}

	logger.Debugf("Всего обработано CSV файлов: %d, создано свечей: %d", csvFileCount, len(candles))
	return candles, nil
}
