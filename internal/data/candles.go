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

	"github.com/russianinvestments/invest-api-go-sdk/investgo"
	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
)

// LoadCandleChunk загружает один чанк свечей согласно лимитам API
func LoadCandleChunk(_ context.Context, client *investgo.Client, figi string, from, to time.Time, interval pb.CandleInterval) ([]*pb.HistoricCandle, error) {
	marketDataClient := client.NewMarketDataServiceClient()

	// Загружаем чанк данных
	candles, err := marketDataClient.GetHistoricCandles(&investgo.GetHistoricCandlesRequest{
		Instrument: figi,
		Interval:   interval,
		From:       from,
		To:         to,
		File:       false,
		FileName:   "",
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки свечей: %w", err)
	}

	return candles, nil
}
