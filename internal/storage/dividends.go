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
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Dividend структура дивиденда
type Dividend struct {
	Figi         string
	PaymentDate  time.Time
	DeclaredDate *time.Time
	Amount       float64
	Currency     string
	YieldPercent *float64
}

// SaveDividend сохраняет информацию о дивиденде
func SaveDividend(ctx context.Context, dbpool *pgxpool.Pool, dividend Dividend) error {
	query := `
		INSERT INTO dividends (figi, payment_date, declared_date, amount, currency, yield_percent)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (figi, payment_date) DO UPDATE SET
			declared_date = EXCLUDED.declared_date,
			amount = EXCLUDED.amount,
			currency = EXCLUDED.currency,
			yield_percent = EXCLUDED.yield_percent
	`

	_, err := dbpool.Exec(ctx, query,
		dividend.Figi, dividend.PaymentDate, dividend.DeclaredDate,
		dividend.Amount, dividend.Currency, dividend.YieldPercent)

	return fmt.Errorf("ошибка сохранения дивиденда: %w", err)
}

// GetLastDividendDate получает дату последней выплаты дивидендов
func GetLastDividendDate(ctx context.Context, dbpool *pgxpool.Pool, figi string) (time.Time, error) {
	query := `SELECT MAX(payment_date) FROM dividends WHERE figi = $1`

	var lastDividendDate sql.NullTime
	err := dbpool.QueryRow(ctx, query, figi).Scan(&lastDividendDate)

	if err == pgx.ErrNoRows || !lastDividendDate.Valid {
		return time.Time{}, nil // Нет записей - новый инструмент
	}

	return lastDividendDate.Time, fmt.Errorf("ошибка сканирования даты последнего дивиденда: %w", err)
}
