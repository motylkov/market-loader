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
	"github.com/sirupsen/logrus"
)

// Instrument структура инструмента
type Instrument struct {
	Figi              string
	Ticker            string
	Name              string
	InstrumentType    string
	Currency          string
	LotSize           int32
	MinPriceIncrement float64
	TradingStatus     string
	Enabled           bool
	Isin              string    // ISIN код инструмента
	ShortEnabledFlag  bool      // Флаг доступности для шорта
	IpoDate           time.Time // Дата IPO (для акций)
	IssueSize         int64     // Размер выпуска
	Sector            string    // Сектор экономики
	RealExchange      string    // Реальная биржа торговли
	// Даты первых свечей для оптимизации загрузки
	First1MinCandleDate time.Time // Дата первой 1-минутной свечи
	First1DayCandleDate time.Time // Дата первой дневной свечи
	// Метаданные
	DataSourceID   int32 // ID источника данных
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastLoadedTime time.Time

	ForQualInvestorFlag bool

	// Новые поля из AssetResponse
	AssetUID         string // Уникальный идентификатор актива
	AssetType        string // Тип актива
	AssetDescription string // Описание актива
	//	AssetCountryOfRisk string // Страна риска - нет
	//	AssetSector        string // Сектор (более детальный) - нет

	// Новые поля из AssetSecurity
	SecurityType          string  // Тип ценной бумаги
	InstrumentKind        string  // Тип инструмента
	FaceValue             float64 // Номинальная стоимость
	FaceUnit              string  // Валюта номинала
	IssueDate             string  // Дата начала торгов
	ListingLevel          int     // Уровень листинга
	RegistrarName         string  // Наименование регистратора
	CouponQuantityPerYear int     // Количество купонов в год

	// Для акций
	ShareType     string // Тип акции (обыкновенная, привилегированная)
	DivYieldFlag  bool   // Флаг дивидендной доходности
	IssueSizePlan int64  // Плановый объем выпуска

	// Для облигаций
	StateRegDate   string  // Дата гос. регистрации
	PlacementDate  string  // Дата размещения
	PlacementPrice float64 // Цена размещения
}

// SaveInstrument сохраняет информацию об инструменте
func SaveInstrument(ctx context.Context, dbpool *pgxpool.Pool, instrument Instrument) error {
	query := `
		INSERT INTO instruments (
			figi, ticker, name, instrument_type, currency, lot_size, min_price_increment, 
			trading_status, enabled, isin, short_enabled_flag, ipo_date, issue_size, 
			sector, real_exchange, first_1min_candle_date, first_1day_candle_date, 
			data_source_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		ON CONFLICT (figi) DO UPDATE SET
			ticker = EXCLUDED.ticker,
			name = EXCLUDED.name,
			instrument_type = EXCLUDED.instrument_type,
			currency = EXCLUDED.currency,
			lot_size = EXCLUDED.lot_size,
			min_price_increment = EXCLUDED.min_price_increment,
			trading_status = EXCLUDED.trading_status,
			isin = EXCLUDED.isin,
			short_enabled_flag = EXCLUDED.short_enabled_flag,
			ipo_date = EXCLUDED.ipo_date,
			issue_size = EXCLUDED.issue_size,
			sector = EXCLUDED.sector,
			real_exchange = EXCLUDED.real_exchange,
			first_1min_candle_date = EXCLUDED.first_1min_candle_date,
			first_1day_candle_date = EXCLUDED.first_1day_candle_date,
			data_source_id = EXCLUDED.data_source_id,
			-- Не изменяем флаг enabled при обновлении существующих записей
			updated_at = NOW()
	`

	_, err := dbpool.Exec(ctx, query,
		instrument.Figi, instrument.Ticker, instrument.Name, instrument.InstrumentType,
		instrument.Currency, instrument.LotSize, instrument.MinPriceIncrement, instrument.TradingStatus, instrument.Enabled,
		instrument.Isin, instrument.ShortEnabledFlag, instrument.IpoDate, instrument.IssueSize,
		instrument.Sector, instrument.RealExchange, instrument.First1MinCandleDate, instrument.First1DayCandleDate,
		instrument.DataSourceID, instrument.CreatedAt, instrument.UpdatedAt)

	if err != nil {
		return fmt.Errorf("ошибка сохранения инструмента: %w", err)
	}
	return nil
}

// getInstrumentsInternal внутренняя функция для получения инструментов
func getInstrumentsInternal(ctx context.Context, dbpool *pgxpool.Pool, instrumentType string, enabledOnly bool) ([]Instrument, error) {
	var query string
	var args []interface{}

	baseQuery := `SELECT figi, ticker, name, instrument_type, data_source_id
				FROM instruments 
				WHERE trading_status = 'normal_trading'`
	// baseQuery := `SELECT figi, ticker, name, instrument_type, currency, lot_size, min_price_increment,
	// 			trading_status, enabled, isin, short_enabled_flag, ipo_date, issue_size,
	// 			sector, real_exchange, first_1min_candle_date, first_1day_candle_date,
	// 			data_source_id, created_at, updated_at, last_loaded_time
	// 			FROM instruments
	// 			WHERE trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING'`

	if enabledOnly {
		baseQuery += ` AND enabled = true`
	}

	if instrumentType == "" {
		query = baseQuery + ` ORDER BY instrument_type, ticker`
	} else {
		query = baseQuery + ` AND instrument_type = $1 ORDER BY ticker`
		args = append(args, instrumentType)
	}

	rows, err := dbpool.Query(ctx, query, args...)
	if err != nil {
		errorMsg := "ошибка запроса инструментов"
		if enabledOnly {
			errorMsg = "ошибка запроса включенных инструментов"
		}
		return nil, fmt.Errorf("%s: %w", errorMsg, err)
	}
	defer rows.Close()

	var instruments []Instrument
	for rows.Next() {
		var instrument Instrument
		err := rows.Scan(
			&instrument.Figi,
			&instrument.Ticker,
			&instrument.Name,
			&instrument.InstrumentType,
			// &instrument.Currency,
			// &instrument.LotSize,
			// &instrument.MinPriceIncrement,
			// &instrument.TradingStatus,
			// &instrument.Enabled,
			// &instrument.Isin,
			// &instrument.ShortEnabledFlag,
			// &instrument.IpoDate,
			// &instrument.IssueSize,
			// &instrument.Sector,
			// &instrument.RealExchange,
			// &instrument.First1MinCandleDate,
			// &instrument.First1DayCandleDate,
			&instrument.DataSourceID,
			// &instrument.CreatedAt,
			// &instrument.UpdatedAt,
			// &instrument.LastLoadedTime,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования инструмента: %w", err)
		}
		instruments = append(instruments, instrument)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по инструментам: %w", err)
	}

	return instruments, nil
}

// LoadInstruments загружает список ИЗ БД, только включённые (enabled = true) с логированием
func LoadInstruments(ctx context.Context, dbpool *pgxpool.Pool, logger *logrus.Logger) ([]Instrument, error) {
	logger.Debug("Загружаем инструменты из БД")

	// Загружаем инструменты из базы данных
	instruments, err := GetEnabledInstruments(ctx, dbpool, "")
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	logger.WithField("count", len(instruments)).Debug("Получены включенные (enabled=true) инструменты")
	return instruments, nil
}

// GetInstruments получает список инструментов из базы данных
func GetInstruments(ctx context.Context, dbpool *pgxpool.Pool, instrumentType string) ([]Instrument, error) {
	return getInstrumentsInternal(ctx, dbpool, instrumentType, false)
}

// GetEnabledInstruments получает только включенные инструменты для загрузки свечей
func GetEnabledInstruments(ctx context.Context, dbpool *pgxpool.Pool, instrumentType string) ([]Instrument, error) {
	return getInstrumentsInternal(ctx, dbpool, instrumentType, true)
}

// UpdateLastLoadedTime обновляет время последней загрузки для инструмента
// поле для информации
func UpdateLastLoadedTime(ctx context.Context, dbpool *pgxpool.Pool, figi string, lastLoadedTime time.Time) error {
	query := `
		UPDATE instruments 
		SET last_loaded_time = $1 
		WHERE figi = $2
	`

	_, err := dbpool.Exec(ctx, query, lastLoadedTime, figi)
	if err != nil {
		return fmt.Errorf("ошибка обновления времени последней загрузки: %w", err)
	}

	return nil
}
