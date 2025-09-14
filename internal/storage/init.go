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

const newView = 1

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
	// Создаем таблицу data_sources
	dataSourcesTable := `
		CREATE TABLE IF NOT EXISTS data_sources (
			id serial4 NOT NULL,
			"name" varchar(50) NOT NULL,
			description text NULL,
			base_url varchar(200) NULL,
			created_at timestamp DEFAULT now() NULL,
			updated_at timestamp DEFAULT now() NULL,
			CONSTRAINT data_sources_name_key UNIQUE (name),
			CONSTRAINT data_sources_pkey PRIMARY KEY (id)
		);
	`

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
			isin varchar(12) NULL,
			short_enabled_flag boolean DEFAULT false NOT NULL,
			ipo_date date NULL,
			issue_size bigint NULL,
			sector varchar(100) NULL,
			real_exchange varchar(50) NULL,
			first_1min_candle_date timestamp NULL,
			first_1day_candle_date timestamp NULL,
			data_source_id int4 NULL,
			created_at timestamp DEFAULT now() NOT NULL,
			updated_at timestamp DEFAULT now() NOT NULL,
			last_loaded_time timestamp NULL,
			enabled bool DEFAULT false NOT NULL,
			CONSTRAINT instruments_pkey PRIMARY KEY (figi),
			CONSTRAINT instruments_data_source_id_fkey FOREIGN KEY (data_source_id) REFERENCES data_sources(id)
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
	// data_sources должна быть создана первой
	queries := []string{dataSourcesTable, instrumentsTable, candlesTable, dividendsTable}
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
		// Индексы для candles
		`CREATE INDEX IF NOT EXISTS idx_candles_figi_interval ON candles(figi, interval_type);`,
		`CREATE INDEX IF NOT EXISTS idx_candles_time ON candles(time);`,

		// Индексы для instruments
		`CREATE INDEX IF NOT EXISTS idx_instruments_ticker ON instruments(ticker);`,
		`CREATE INDEX IF NOT EXISTS idx_instruments_type ON instruments(instrument_type);`,
		`CREATE INDEX IF NOT EXISTS idx_instruments_enabled ON instruments(enabled);`,
		`CREATE INDEX IF NOT EXISTS idx_instruments_isin ON instruments(isin);`,
		`CREATE INDEX IF NOT EXISTS idx_instruments_sector ON instruments(sector);`,
		`CREATE INDEX IF NOT EXISTS idx_instruments_real_exchange ON instruments(real_exchange);`,
		`CREATE INDEX IF NOT EXISTS idx_instruments_ipo_date ON instruments(ipo_date);`,
		`CREATE INDEX IF NOT EXISTS idx_instruments_first_1min_candle_date ON instruments(first_1min_candle_date);`,
		`CREATE INDEX IF NOT EXISTS idx_instruments_first_1day_candle_date ON instruments(first_1day_candle_date);`,
		`CREATE INDEX IF NOT EXISTS idx_instruments_data_source_id ON instruments(data_source_id);`,

		// Индексы для dividends
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

	// Создаем представление instrument_view
	createView := `
		CREATE OR REPLACE VIEW instrument_view
		AS SELECT 
			i.ticker,
			i.figi,
			i.name,
			i.instrument_type,
			i.currency,
			i.lot_size,
			i.isin,
			i.short_enabled_flag,
			i.ipo_date,
			i.issue_size,
			i.sector,
			i.real_exchange,
			i.first_1min_candle_date,
			i.first_1day_candle_date,
			ds.name AS data_source_name,
			i.enabled,
			i.last_loaded_time,
			i.created_at,
			i.updated_at
		FROM instruments i
		LEFT JOIN data_sources ds ON i.data_source_id = ds.id;
	`

	// Выполняем создание индексов, ограничений и представления
	queries := make([]string, 0, len(indexes)+len(foreignKeys)+newView)
	queries = append(queries, indexes...)
	queries = append(queries, foreignKeys...)
	queries = append(queries, createView)

	for _, query := range queries {
		_, err := dbpool.Exec(context.Background(), query)
		if err != nil {
			return fmt.Errorf("ошибка создания индекса/ограничения/представления: %w", err)
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

	// Создаем таблицу data_sources если её нет
	createDataSourcesTable := `
		CREATE TABLE IF NOT EXISTS data_sources (
			id serial4 NOT NULL,
			"name" varchar(50) NOT NULL,
			description text NULL,
			base_url varchar(200) NULL,
			created_at timestamp DEFAULT now() NULL,
			updated_at timestamp DEFAULT now() NULL,
			CONSTRAINT data_sources_name_key UNIQUE (name),
			CONSTRAINT data_sources_pkey PRIMARY KEY (id)
		);
	`

	// Добавляем новые поля в таблицу instruments
	addInstrumentFields := `
		DO $$ 
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'instruments') THEN
				-- Добавляем новые поля если их нет
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
					WHERE table_name = 'instruments' AND column_name = 'isin') THEN
					ALTER TABLE instruments ADD COLUMN isin varchar(12) NULL;
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
					WHERE table_name = 'instruments' AND column_name = 'short_enabled_flag') THEN
					ALTER TABLE instruments ADD COLUMN short_enabled_flag boolean DEFAULT false NOT NULL;
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
					WHERE table_name = 'instruments' AND column_name = 'ipo_date') THEN
					ALTER TABLE instruments ADD COLUMN ipo_date date NULL;
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
					WHERE table_name = 'instruments' AND column_name = 'issue_size') THEN
					ALTER TABLE instruments ADD COLUMN issue_size bigint NULL;
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
					WHERE table_name = 'instruments' AND column_name = 'sector') THEN
					ALTER TABLE instruments ADD COLUMN sector varchar(100) NULL;
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
					WHERE table_name = 'instruments' AND column_name = 'real_exchange') THEN
					ALTER TABLE instruments ADD COLUMN real_exchange varchar(50) NULL;
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
					WHERE table_name = 'instruments' AND column_name = 'first_1min_candle_date') THEN
					ALTER TABLE instruments ADD COLUMN first_1min_candle_date timestamp NULL;
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
					WHERE table_name = 'instruments' AND column_name = 'first_1day_candle_date') THEN
					ALTER TABLE instruments ADD COLUMN first_1day_candle_date timestamp NULL;
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
					WHERE table_name = 'instruments' AND column_name = 'data_source_id') THEN
					ALTER TABLE instruments ADD COLUMN data_source_id int4 NULL;
				END IF;
			END IF;
		END $$;
	`

	// Добавляем индексы для новых полей
	addNewIndexes := `
		DO $$ 
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'instruments') THEN
				-- Создаем индексы для новых полей если их нет
				IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_instruments_isin') THEN
					CREATE INDEX idx_instruments_isin ON instruments USING btree (isin);
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_instruments_sector') THEN
					CREATE INDEX idx_instruments_sector ON instruments USING btree (sector);
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_instruments_real_exchange') THEN
					CREATE INDEX idx_instruments_real_exchange ON instruments USING btree (real_exchange);
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_instruments_ipo_date') THEN
					CREATE INDEX idx_instruments_ipo_date ON instruments USING btree (ipo_date);
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_instruments_first_1min_candle_date') THEN
					CREATE INDEX idx_instruments_first_1min_candle_date ON instruments USING btree (first_1min_candle_date);
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_instruments_first_1day_candle_date') THEN
					CREATE INDEX idx_instruments_first_1day_candle_date ON instruments USING btree (first_1day_candle_date);
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_instruments_data_source_id') THEN
					CREATE INDEX idx_instruments_data_source_id ON instruments USING btree (data_source_id);
				END IF;
			END IF;
		END $$;
	`

	// Добавляем внешний ключ для data_source_id
	addDataSourceForeignKey := `
		DO $$ 
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'instruments') 
			   AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'data_sources') THEN
				IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
					WHERE table_name = 'instruments' AND constraint_name = 'instruments_data_source_id_fkey') THEN
					ALTER TABLE instruments ADD CONSTRAINT instruments_data_source_id_fkey 
						FOREIGN KEY (data_source_id) REFERENCES data_sources(id);
				END IF;
			END IF;
		END $$;
	`

	// Обновляем представление instrument_view
	updateInstrumentView := `
		DROP VIEW IF EXISTS instrument_view;
		CREATE OR REPLACE VIEW instrument_view
		AS SELECT 
			i.ticker,
			i.figi,
			i.name,
			i.instrument_type,
			i.currency,
			i.lot_size,
			i.isin,
			i.short_enabled_flag,
			i.ipo_date,
			i.issue_size,
			i.sector,
			i.real_exchange,
			i.first_1min_candle_date,
			i.first_1day_candle_date,
			ds.name AS data_source_name,
			i.enabled,
			i.last_loaded_time,
			i.created_at,
			i.updated_at
		FROM instruments i
		LEFT JOIN data_sources ds ON i.data_source_id = ds.id;
	`

	queries := []string{
		addEnabledColumn,
		addDividendsUniqueConstraint,
		createDataSourcesTable,
		addInstrumentFields,
		addNewIndexes,
		addDataSourceForeignKey,
		updateInstrumentView,
	}

	for _, query := range queries {
		_, err := dbpool.Exec(context.Background(), query)
		if err != nil {
			return fmt.Errorf("ошибка выполнения миграции: %w", err)
		}
	}

	return nil
}
