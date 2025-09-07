// Package database для подключения к базе данных
// Market Loader
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package database

import (
	"context"
	"fmt"

	"market-loader/pkg/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect подключается к базе данных
func Connect(ctx context.Context, dbConfig *config.DatabaseConfig) (*pgxpool.Pool, error) {
	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName, dbConfig.SSLMode)

	dbpool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пула подключений: %w", err)
	}

	return dbpool, nil
}
