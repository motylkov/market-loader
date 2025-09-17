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

	"market-loader/internal/money"
	"market-loader/internal/storage"
	"market-loader/pkg/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
	"github.com/sirupsen/logrus"
)

// CreateInstrumentFromProto создает структуру Instrument из protobuf данных
func CreateInstrumentFromProto(
	protoInstrument interface{},
	dataSourceID int32,
) (*storage.Instrument, error) {
	now := time.Now()
	var inst storage.Instrument

	// Устанавливаем базовые метаданные
	inst.CreatedAt = now
	inst.UpdatedAt = now
	inst.DataSourceID = dataSourceID

	switch v := protoInstrument.(type) {
	case *pb.Share:
		inst.Figi = orEmpty(&v.Figi)
		inst.Ticker = orEmpty(&v.Ticker)
		inst.Name = escapeTabs(v.GetName())
		inst.InstrumentType = "share"
		inst.Currency = orEmpty(&v.Currency)
		inst.LotSize = v.Lot
		inst.MinPriceIncrement = money.ConvertQuotationToFloat(v.MinPriceIncrement)
		inst.TradingStatus = tradingStatusToString(v.TradingStatus)
		inst.Enabled = v.ApiTradeAvailableFlag
		inst.ShortEnabledFlag = v.ShortEnabledFlag
		inst.Isin = orEmpty(&v.Isin)
		if ts := v.IpoDate; ts != nil {
			t := ts.AsTime()
			inst.IpoDate = t
		}
		if v.IssueSize > 0 {
			inst.IssueSize = v.IssueSize
		}
		inst.RealExchange = v.RealExchange.String()
		if v.ForQualInvestorFlag {
			flag := true
			inst.ForQualInvestorFlag = flag

		}

		// Специфичные поля акций
		shareType := shareTypeToString(v.ShareType)
		if shareType != "" {
			inst.ShareType = shareType
		}

		if v.DivYieldFlag {
			if v.DivYieldFlag {
				flag := true
				inst.DivYieldFlag = flag
			}
		}
		if v.IssueSizePlan > 0 {
			plan := v.IssueSizePlan
			inst.IssueSizePlan = plan
		}

	case *pb.Bond:
		inst.Figi = orEmpty(&v.Figi)
		inst.Ticker = orEmpty(&v.Ticker)
		inst.Name = escapeTabs(v.GetName())
		inst.InstrumentType = "bond"
		inst.Currency = orEmpty(&v.Currency)
		inst.LotSize = v.Lot
		inst.MinPriceIncrement = money.ConvertQuotationToFloat(v.MinPriceIncrement)
		inst.TradingStatus = tradingStatusToString(v.TradingStatus)
		inst.Enabled = v.ApiTradeAvailableFlag
		inst.ShortEnabledFlag = v.ShortEnabledFlag
		inst.Isin = orEmpty(&v.Isin)
		if v.IssueSize > 0 {
			inst.IssueSize = v.IssueSize
		}
		inst.RealExchange = v.RealExchange.String()
		if v.ForQualInvestorFlag {
			flag := true
			inst.ForQualInvestorFlag = flag

		}

		// Поля облигаций
		if ts := v.StateRegDate; ts != nil {
			s := ts.AsTime().Format("2006-01-02")
			inst.StateRegDate = s
		}
		if ts := v.PlacementDate; ts != nil {
			s := ts.AsTime().Format("2006-01-02")
			inst.PlacementDate = s
		}
		inst.PlacementPrice = money.ConvertMoneyValueToFloat(v.PlacementPrice)

	case *pb.Etf:
		inst.Figi = orEmpty(&v.Figi)
		inst.Ticker = orEmpty(&v.Ticker)
		inst.Name = escapeTabs(v.GetName())
		inst.InstrumentType = "etf"
		inst.Currency = orEmpty(&v.Currency)
		inst.LotSize = v.Lot
		inst.MinPriceIncrement = money.ConvertQuotationToFloat(v.MinPriceIncrement)
		inst.TradingStatus = tradingStatusToString(v.TradingStatus)
		inst.Enabled = v.ApiTradeAvailableFlag
		inst.ShortEnabledFlag = v.ShortEnabledFlag
		inst.Isin = orEmpty(&v.Isin)
		inst.RealExchange = v.RealExchange.String()
		if v.ForQualInvestorFlag {
			flag := true
			inst.ForQualInvestorFlag = flag

		}
	default:
		return nil, fmt.Errorf("unknown instrument type: %T", protoInstrument)
	}

	return &inst, nil
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
	client *investgo.Client,
	instruments []T,
	instrumentType string,
	dataSourceID *int32,
	dbpool *pgxpool.Pool,
	logger *logrus.Logger,
) error {
	count := 0

	for _, protoInstrument := range instruments {
		if config.IsNormalTrading(protoInstrument.GetTradingStatus()) {

			// Создаём инструмент с расширенными данными
			instrument, err := CreateInstrumentFromProto(protoInstrument, *dataSourceID)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"figi":   protoInstrument.GetFigi(),
					"ticker": protoInstrument.GetTicker(),
					"type":   instrumentType,
					"error":  err,
				}).Error("Ошибка создания инструмента")
			}

			if err := storage.SaveInstrument(ctx, dbpool, *instrument); err != nil {
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
		return processInstruments(ctx, client, response.Instruments, instrumentType, dataSourceID, dbpool, logger)
	case "bond":
		response, err := instrumentsClient.Bonds(pb.InstrumentStatus_INSTRUMENT_STATUS_ALL)
		if err != nil {
			return fmt.Errorf("ошибка загрузки облигаций: %w", err)
		}
		return processInstruments(ctx, client, response.Instruments, instrumentType, dataSourceID, dbpool, logger)
	case "etf":
		response, err := instrumentsClient.Etfs(pb.InstrumentStatus_INSTRUMENT_STATUS_ALL)
		if err != nil {
			return fmt.Errorf("ошибка загрузки ETF: %w", err)
		}
		return processInstruments(ctx, client, response.Instruments, instrumentType, dataSourceID, dbpool, logger)
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
