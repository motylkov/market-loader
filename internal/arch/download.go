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
	"context"
	"fmt"
	"io"
	"market-loader/pkg/config"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
)

// DownloadYearArchive загружает архив за указанный год
func DownloadYearArchive(ctx context.Context, token, figi string, year int, tempDir string, dbpool *pgxpool.Pool, logger *logrus.Logger) ([]*pb.HistoricCandle, error) {
	// Формируем URL для запроса архива
	url := fmt.Sprintf("https://invest-public-api.tbank.ru/history-data?figi=%s&year=%d", figi, year)

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Выполняем запрос с повторными попытками
	var resp *http.Response
	maxRetries := 3
	retryDelay := config.DefaultRetryDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		client := &http.Client{Timeout: config.DefaultHTTPTimeout}
		resp, err = client.Do(req)

		if err == nil && resp.StatusCode == http.StatusOK {
			logger.Infof("Успешный ответ от API: статус %d, размер: %d байт", resp.StatusCode, resp.ContentLength)
			break
		}

		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				logger.Errorf("Ошибка закрытия тела ответа: %v", closeErr)
			}
		}

		if attempt < maxRetries {
			logger.Debugf("Попытка %d/%d не удалась, повтор через %v...", attempt, maxRetries, retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Экспоненциальная задержка
		} else {
			if err != nil {
				return nil, fmt.Errorf("ошибка выполнения запроса после %d попыток: %w", maxRetries, err)
			}
			return nil, fmt.Errorf("ошибка HTTP %d после %d попыток", resp.StatusCode, maxRetries)
		}
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Errorf("Ошибка закрытия тела ответа: %v", err)
		}
	}()

	// Сохраняем архив во временный файл
	archivePath := filepath.Join(tempDir, fmt.Sprintf("%s_%d.zip", figi, year))

	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания файла архива: %w", err)
	}
	defer func() {
		if err := archiveFile.Close(); err != nil {
			logger.Errorf("Ошибка закрытия файла архива: %v", err)
		}
	}()

	if _, err := io.Copy(archiveFile, resp.Body); err != nil {
		return nil, fmt.Errorf("ошибка сохранения архива: %w", err)
	}

	// Обрабатываем ZIP архив
	return processArchive(archivePath, figi, dbpool, logger)
}
