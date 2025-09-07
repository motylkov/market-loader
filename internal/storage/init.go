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
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreatePartition создает партицию
func CreatePartition(dbpool *pgxpool.Pool, t time.Time) error {
	// Начало месяца
	monthStart := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	// Конец месяца (начало следующего месяца минус 1 секунда)
	monthEnd := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0).Add(-time.Second)
	// Название партиции
	partitionName := fmt.Sprintf("candles_%d_%02d", t.Year(), t.Month())

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s PARTITION OF candles
			FOR VALUES FROM ('%s') TO ('%s')
		`, partitionName,
		monthStart.Format("2006-01-02 15:04:05"),
		monthEnd.Format("2006-01-02 15:04:05"))

	_, err := dbpool.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("ошибка создания партиции: %w", err)
	}
	return nil
}

// CreateInitialPartition создает начальную партицию для текущего месяца
func CreateInitialPartition(dbpool *pgxpool.Pool) error {
	// Создаем партицию для текущего месяца
	if err := CreatePartition(dbpool, time.Now()); err != nil {
		return fmt.Errorf("ошибка создания партиции для текущего месяца: %w", err)
	}
	return nil
}

// CreateYearPartitions создает все партиции для указанного года
func CreateYearPartitions(dbpool *pgxpool.Pool, year int) error {
	for month := 1; month <= 12; month++ {
		t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		if err := CreatePartition(dbpool, t); err != nil {
			return fmt.Errorf("ошибка создания партиции для %d-%02d: %w", year, month, err)
		}
	}
	return nil
}

// InitDatabase инициализирует базу данных, создавая необходимые таблицы
func InitDatabase(dbpool *pgxpool.Pool) error {
	// Создаем таблицу instruments
	instrumentsTable := `
		CREATE TABLE IF NOT EXISTS instruments (
			figi varchar(50) NOT NULL,
			ticker varchar(30) NOT NULL,
			name text NOT NULL,
			instrument_type varchar(20) NOT NULL,
			currency varchar(3) NOT NULL,
			lot_size int4 NOT NULL,
			min_price_increment numeric(20, 9) NOT NULL,
			trading_status varchar(40) NOT NULL,
			enabled bool DEFAULT false NOT NULL,
			created_at timestamp DEFAULT now() NOT NULL,
			updated_at timestamp DEFAULT now() NOT NULL,
			last_loaded_time timestamp NULL,
			CONSTRAINT instruments_pkey PRIMARY KEY (figi)
		);
	`

	// Создаем таблицу candles
	candlesTable := `
		CREATE TABLE IF NOT EXISTS candles (
			id BIGSERIAL,
			figi VARCHAR(50) NOT NULL,
			time TIMESTAMP NOT NULL,
			open_price DECIMAL(20, 9) NOT NULL,
			high_price DECIMAL(20, 9) NOT NULL,
			low_price DECIMAL(20, 9) NOT NULL,
			close_price DECIMAL(20, 9) NOT NULL,
			volume BIGINT NOT NULL,
			interval_type VARCHAR(30) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (figi, time, interval_type)
		) PARTITION BY RANGE ("time");
	`

	// Создаем таблицу dividends
	dividendsTable := `
		CREATE TABLE IF NOT EXISTS dividends (
			id BIGSERIAL,
			figi VARCHAR(50) NOT NULL,
			payment_date TIMESTAMPTZ NOT NULL,
			declared_date TIMESTAMPTZ NULL,
			amount NUMERIC(20, 10) NOT NULL,
			currency VARCHAR(3) NULL,
			yield_percent NUMERIC(5, 2) NULL,
			created_at TIMESTAMPTZ DEFAULT NOW() NULL,
			PRIMARY KEY (id),
			UNIQUE (figi, payment_date)
		);
	`

	// Выполняем создание таблиц
	queries := []string{instrumentsTable, candlesTable, dividendsTable}
	for _, query := range queries {
		_, err := dbpool.Exec(context.Background(), query)
		if err != nil {
			return fmt.Errorf("ошибка создания таблицы: %w", err)
		}
	}

	return nil
}

// CreateIndexesAndConstraints создает индексы и ограничения для таблиц
func CreateIndexesAndConstraints(dbpool *pgxpool.Pool) error {
	// Создаем индексы для оптимизации запросов
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_candles_figi_interval ON candles(figi, interval_type);`,
		`CREATE INDEX IF NOT EXISTS idx_candles_time ON candles(time);`,
		`CREATE INDEX IF NOT EXISTS idx_instruments_ticker ON instruments(ticker);`,
		`CREATE INDEX IF NOT EXISTS idx_instruments_type ON instruments(instrument_type);`,
		`CREATE INDEX IF NOT EXISTS idx_instruments_enabled ON instruments(enabled);`,
		`CREATE INDEX IF NOT EXISTS idx_dividends_figi ON dividends(figi);`,
		`CREATE INDEX IF NOT EXISTS idx_dividends_payment_date ON dividends(payment_date);`,
	}

	// Создаем внешние ключи для обеспечения целостности данных
	foreignKeys := []string{
		`DO $$ 
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'candles_figi_fkey') THEN
				ALTER TABLE candles ADD CONSTRAINT candles_figi_fkey 
					FOREIGN KEY (figi) REFERENCES instruments(figi) ON UPDATE CASCADE ON DELETE CASCADE;
			END IF;
		END $$;`,
		`DO $$ 
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'dividends_figi_fkey') THEN
				ALTER TABLE dividends ADD CONSTRAINT dividends_figi_fkey 
					FOREIGN KEY (figi) REFERENCES instruments(figi) ON UPDATE CASCADE ON DELETE CASCADE;
			END IF;
		END $$;`,
	}

	// Выполняем создание индексов и ограничений
	queries := make([]string, 0, len(indexes)+len(foreignKeys))
	queries = append(queries, indexes...)
	queries = append(queries, foreignKeys...)
	for _, query := range queries {
		_, err := dbpool.Exec(context.Background(), query)
		if err != nil {
			return fmt.Errorf("ошибка создания индекса/ограничения: %w", err)
		}
	}

	return nil
}

// MigrateDatabase выполняет миграции для существующих таблиц
func MigrateDatabase(dbpool *pgxpool.Pool) error {
	// Добавляем колонку enabled в таблицу instruments если её нет
	addEnabledColumn := `
		DO $$ 
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'instruments') THEN
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
					WHERE table_name = 'instruments' AND column_name = 'enabled') THEN
					ALTER TABLE instruments ADD COLUMN enabled BOOLEAN DEFAULT FALSE;
				END IF;
			END IF;
		END $$;
	`

	// Добавляем уникальное ограничение для dividends если его нет
	addDividendsUniqueConstraint := `
		DO $$ 
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'dividends') THEN
				-- Проверяем, есть ли дублирующиеся записи перед добавлением ограничения
				IF EXISTS (
					SELECT figi, payment_date, COUNT(*) 
					FROM dividends 
					GROUP BY figi, payment_date 
					HAVING COUNT(*) > 1
				) THEN
					-- Если есть дубликаты, удаляем их, оставляя только одну запись
					DELETE FROM dividends 
					WHERE id NOT IN (
						SELECT MIN(id) 
						FROM dividends 
						GROUP BY figi, payment_date
					);
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
					WHERE table_name = 'dividends' AND constraint_type = 'UNIQUE' 
					AND constraint_name LIKE '%figi%payment_date%') THEN
					ALTER TABLE dividends ADD CONSTRAINT dividends_figi_payment_date_unique 
						UNIQUE (figi, payment_date);
				END IF;
			END IF;
		END $$;
	`

	queries := []string{addEnabledColumn, addDividendsUniqueConstraint}
	for _, query := range queries {
		_, err := dbpool.Exec(context.Background(), query)
		if err != nil {
			return fmt.Errorf("ошибка выполнения миграции: %w", err)
		}
	}

	return nil
}
