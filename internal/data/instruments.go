// Package data - Запросы в API и обработка данных
// Market Loader
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	"market-loader/internal/money"
	"market-loader/internal/storage"
	"market-loader/pkg/config"
)

// CreateInstrumentFromProto создает структуру Instrument из protobuf данных
func CreateInstrumentFromProto(
	figi, ticker, name, instrumentType, currency string,
	lotSize int32, minPriceIncrement *pb.Quotation,
	tradingStatus pb.SecurityTradingStatus,
	isin *string,
	shortEnabledFlag bool,
	ipoDate *time.Time,
	issueSize *int64,
	sector, realExchange *string,
	first1MinCandleDate, first1DayCandleDate *time.Time,
	dataSourceID *int32,
) storage.Instrument {
	now := time.Now()
	return storage.Instrument{
		Figi:                figi,
		Ticker:              ticker,
		Name:                name,
		InstrumentType:      instrumentType,
		Currency:            currency,
		LotSize:             lotSize,
		MinPriceIncrement:   money.ConvertMinPriceIncrement(minPriceIncrement),
		TradingStatus:       tradingStatus.String(),
		Enabled:             false,
		Isin:                isin,
		ShortEnabledFlag:    shortEnabledFlag,
		IpoDate:             ipoDate,
		IssueSize:           issueSize,
		Sector:              sector,
		RealExchange:        realExchange,
		First1MinCandleDate: first1MinCandleDate,
		First1DayCandleDate: first1DayCandleDate,
		DataSourceID:        dataSourceID,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// processInstruments обрабатывает и сохраняет инструменты
func processInstruments[T interface {
	GetFigi() string
	GetTicker() string
	GetName() string
	GetCurrency() string
	GetLot() int32
	GetMinPriceIncrement() *pb.Quotation
	GetTradingStatus() pb.SecurityTradingStatus
}](
	ctx context.Context,
	instruments []T,
	instrumentType string,
	dataSourceID *int32,
	dbpool *pgxpool.Pool,
	logger *logrus.Logger,
) error {
	count := 0

	for _, protoInstrument := range instruments {
		if config.IsNormalTrading(protoInstrument.GetTradingStatus()) {
			var (
				// isin                string
				isin             *string
				shortEnabledFlag bool
				ipoDate          *time.Time
				issueSize        int64 = 0
				// sector              string
				sector *string
				// realExchange        string
				realExchange        *string
				first1MinCandleDate *time.Time
				first1DayCandleDate *time.Time
			)

			// Извлечение общих полей из инструментов
			extractCommonFields := func(v interface{}) {
				setIfNotNil := func(dst **string, getFn func() string) {
					if val := getFn(); val != "" {
						*dst = &val
					}
				}

				// Общие строки
				setIfNotNil(&isin, v.(interface{ GetIsin() string }).GetIsin)
				shortEnabledFlag = v.(interface{ GetShortEnabledFlag() bool }).GetShortEnabledFlag()
				setIfNotNil(&sector, v.(interface{ GetSector() string }).GetSector)

				// RealExchange — enum -> string
				if reGetter, ok := v.(interface{ GetRealExchange() string }); ok {
					if re := reGetter.GetRealExchange(); re != "" {
						realExchange = &re
					}
				}

				// Первые свечи
				if tsGetter, ok := v.(interface{ GetFirst_1MinCandleDate() *timestamppb.Timestamp }); ok {
					if ts := tsGetter.GetFirst_1MinCandleDate(); ts != nil {
						t := ts.AsTime()
						first1MinCandleDate = &t
					}
				}
				if tsGetter, ok := v.(interface{ GetFirst_1DayCandleDate() *timestamppb.Timestamp }); ok {
					if ts := tsGetter.GetFirst_1DayCandleDate(); ts != nil {
						t := ts.AsTime()
						first1DayCandleDate = &t
					}
				}
			}

			// Обработка по типам
			switch v := any(protoInstrument).(type) {
			case *pb.Share:
				extractCommonFields(v)
				// Специфичные поля
				issueSize = v.GetIssueSize()
				if ts := v.GetIpoDate(); ts != nil {
					t := ts.AsTime()
					ipoDate = &t
				}
			case *pb.Bond:
				extractCommonFields(v)
				issueSize = v.GetIssueSize()
				// IPO не устанавливаем
			case *pb.Etf:
				extractCommonFields(v)
				// Ни issue_size, ни IPO не устанавливаем
			default:
				logger.WithFields(logrus.Fields{
					"figi":   protoInstrument.GetFigi(),
					"ticker": protoInstrument.GetTicker(),
					"type":   fmt.Sprintf("%T", protoInstrument),
				}).Warn("Неизвестный тип инструмента, пропуск")
				continue
			}

			// Создаём инструмент с расширенными данными
			instrument := CreateInstrumentFromProto(
				protoInstrument.GetFigi(),
				protoInstrument.GetTicker(),
				protoInstrument.GetName(),
				instrumentType,
				protoInstrument.GetCurrency(),
				protoInstrument.GetLot(),
				protoInstrument.GetMinPriceIncrement(),
				protoInstrument.GetTradingStatus(),
				isin,
				shortEnabledFlag,
				ipoDate,
				&issueSize,
				sector,
				realExchange,
				first1MinCandleDate,
				first1DayCandleDate,
				dataSourceID,
			)

			if err := storage.SaveInstrument(ctx, dbpool, instrument); err != nil {
				logger.WithFields(logrus.Fields{
					"figi":   protoInstrument.GetFigi(),
					"ticker": protoInstrument.GetTicker(),
					"type":   instrumentType,
					"error":  err,
				}).Error("Ошибка сохранения инструмента")
				continue
			}
			count++
		}
	}

	logger.WithFields(logrus.Fields{
		"type":  instrumentType,
		"count": count,
	}).Info("Инструменты загружены с расширенными данными")
	return nil
}

// LoadInstrumentsByType загружает инструменты определенного типа из API и сохраняет в БД
func LoadInstrumentsByType(
	ctx context.Context,
	client *investgo.Client,
	dbpool *pgxpool.Pool,
	instrumentType string,
	dataSourceID *int32,
	logger *logrus.Logger,
) error {
	instrumentsClient := client.NewInstrumentsServiceClient()

	// Получаем инструменты в зависимости от типа
	switch instrumentType {
	case "share":
		response, err := instrumentsClient.Shares(pb.InstrumentStatus_INSTRUMENT_STATUS_ALL)
		if err != nil {
			return fmt.Errorf("ошибка загрузки акций: %w", err)
		}
		return processInstruments(ctx, response.Instruments, instrumentType, dataSourceID, dbpool, logger)
	case "bond":
		response, err := instrumentsClient.Bonds(pb.InstrumentStatus_INSTRUMENT_STATUS_ALL)
		if err != nil {
			return fmt.Errorf("ошибка загрузки облигаций: %w", err)
		}
		return processInstruments(ctx, response.Instruments, instrumentType, dataSourceID, dbpool, logger)
	case "etf":
		response, err := instrumentsClient.Etfs(pb.InstrumentStatus_INSTRUMENT_STATUS_ALL)
		if err != nil {
			return fmt.Errorf("ошибка загрузки ETF: %w", err)
		}
		return processInstruments(ctx, response.Instruments, instrumentType, dataSourceID, dbpool, logger)
	default:
		return fmt.Errorf("неподдерживаемый тип инструмента: %s", instrumentType)
	}
}

// GetOrCreateTInvestDataSource получает или создает запись источника данных T-Invest
func GetOrCreateTInvestDataSource(ctx context.Context, dbpool *pgxpool.Pool) (*int32, error) {
	// Сначала пытаемся найти существующую запись
	var dataSourceID int32
	query := `SELECT id FROM data_sources WHERE name = 'T-Invest API'`
	err := dbpool.QueryRow(ctx, query).Scan(&dataSourceID)
	if err == nil {
		return &dataSourceID, nil
	}

	// Если не найдена, создаем новую
	insertQuery := `
		INSERT INTO data_sources (name, description, base_url, created_at, updated_at)
		VALUES ('T-Invest API', 'T-Invest API - API для получения рыночных данных', 'https://invest-public-api.tinkoff.ru', NOW(), NOW())
		RETURNING id
	`
	err = dbpool.QueryRow(ctx, insertQuery).Scan(&dataSourceID)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания источника данных T-Invest: %w", err)
	}

	return &dataSourceID, nil
}
