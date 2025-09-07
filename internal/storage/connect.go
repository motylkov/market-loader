// Package storage содержит функции для работы с базой данных свечей
// Market Loader
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package storage

import (
	"context"
	"fmt"

	"market-loader/pkg/config"
	"market-loader/pkg/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectToDatabase подключается к базе данных и инициализирует её
func ConnectToDatabase(ctx context.Context, dbConfig *config.DatabaseConfig) (*pgxpool.Pool, error) {
	// Подключаемся к БД
	dbpool, err := database.Connect(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	// Сначала выполняем миграции для существующих таблиц
	if err := MigrateDatabase(dbpool); err != nil {
		dbpool.Close()
		return nil, fmt.Errorf("ошибка миграции БД: %w", err)
	}

	// Затем создаем базовые таблицы
	if err := InitDatabase(dbpool); err != nil {
		dbpool.Close()
		return nil, fmt.Errorf("ошибка инициализации БД: %w", err)
	}

	// После миграций создаем индексы и ограничения
	if err := CreateIndexesAndConstraints(dbpool); err != nil {
		dbpool.Close()
		return nil, fmt.Errorf("ошибка создания индексов и ограничений: %w", err)
	}

	// Создаем начальную партицию для текущего месяца
	if err := CreateInitialPartition(dbpool); err != nil {
		dbpool.Close()
		return nil, fmt.Errorf("ошибка создания начальной партиции: %w", err)
	}

	return dbpool, nil
}
