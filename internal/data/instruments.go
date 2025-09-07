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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
	"github.com/sirupsen/logrus"

	"market-loader/internal/money"
	"market-loader/internal/storage"
	"market-loader/pkg/config"
)

// CreateInstrumentFromProto создает структуру Instrument из protobuf данных
func CreateInstrumentFromProto(
	figi, ticker, name, instrumentType, currency string,
	lotSize int32, minPriceIncrement *pb.Quotation,
	tradingStatus pb.SecurityTradingStatus,
) storage.Instrument {
	return storage.Instrument{
		Figi:              figi,
		Ticker:            ticker,
		Name:              name,
		InstrumentType:    instrumentType,
		Currency:          currency,
		LotSize:           lotSize,
		MinPriceIncrement: money.ConvertMinPriceIncrement(minPriceIncrement),
		TradingStatus:     tradingStatus.String(),
		Enabled:           false,
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
	dbpool *pgxpool.Pool,
	logger *logrus.Logger,
) error {
	count := 0
	for _, protoInstrument := range instruments {
		if config.IsNormalTrading(protoInstrument.GetTradingStatus()) {
			instrument := CreateInstrumentFromProto(
				protoInstrument.GetFigi(),
				protoInstrument.GetTicker(),
				protoInstrument.GetName(),
				instrumentType,
				protoInstrument.GetCurrency(),
				protoInstrument.GetLot(),
				protoInstrument.GetMinPriceIncrement(),
				protoInstrument.GetTradingStatus(),
			)

			if err := storage.SaveInstrument(ctx, dbpool, instrument); err != nil {
				logger.WithFields(logrus.Fields{
					"figi":   protoInstrument.GetFigi(),
					"ticker": protoInstrument.GetTicker(),
					"error":  err,
				}).Errorf("Ошибка сохранения %s", instrumentType)
				continue
			}
			count++
		}
	}

	logger.WithField("count", count).Infof("%s загружены", instrumentType)
	return nil
}

// LoadInstrumentsByType загружает инструменты определенного типа из API и сохраняет в БД
func LoadInstrumentsByType(
	ctx context.Context,
	client *investgo.Client,
	dbpool *pgxpool.Pool,
	instrumentType string,
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
		return processInstruments(ctx, response.Instruments, instrumentType, dbpool, logger)
	case "bond":
		response, err := instrumentsClient.Bonds(pb.InstrumentStatus_INSTRUMENT_STATUS_ALL)
		if err != nil {
			return fmt.Errorf("ошибка загрузки облигаций: %w", err)
		}
		return processInstruments(ctx, response.Instruments, instrumentType, dbpool, logger)
	case "etf":
		response, err := instrumentsClient.Etfs(pb.InstrumentStatus_INSTRUMENT_STATUS_ALL)
		if err != nil {
			return fmt.Errorf("ошибка загрузки ETF: %w", err)
		}
		return processInstruments(ctx, response.Instruments, instrumentType, dbpool, logger)
	default:
		return fmt.Errorf("неподдерживаемый тип инструмента: %s", instrumentType)
	}
}
