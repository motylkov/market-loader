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
	"fmt"
	"market-loader/internal/money"
	"market-loader/internal/storage"
	"strconv"
	"time"

	"github.com/russianinvestments/invest-api-go-sdk/investgo"
)

// LoadDividends загружает дивиденды для инструмента
func LoadDividends(client *investgo.Client, figi string, from, to time.Time) ([]storage.Dividend, error) {
	instrumentsClient := client.NewInstrumentsServiceClient()

	// Загружаем дивиденды через API
	dividends, err := instrumentsClient.GetDividents(figi, from, to)

	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки дивидендов: %w", err)
	}

	result := make([]storage.Dividend, 0, len(dividends.Dividends))

	for _, dividend := range dividends.Dividends {
		// Конвертируем в нашу структуру
		dbDividend := storage.Dividend{
			Figi:        figi,
			PaymentDate: dividend.GetPaymentDate().AsTime(),
		}

		// Обрабатываем declared_date (может быть nil)
		if dividend.GetDeclaredDate() != nil {
			declaredDate := dividend.GetDeclaredDate().AsTime()
			dbDividend.DeclaredDate = &declaredDate
		}

		// Обрабатываем dividend_net (сумма дивиденда)
		if dividend.GetDividendNet() != nil {
			// Используем точное преобразование для избежания проблем с плавающей точкой
			amountStr := money.ConvertMoneyValue(dividend.GetDividendNet().GetUnits(), dividend.GetDividendNet().GetNano())
			if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
				dbDividend.Amount = amount
			}
			dbDividend.Currency = dividend.GetDividendNet().GetCurrency()
		}

		// Обрабатываем yield_value (доходность)
		if dividend.GetYieldValue() != nil {
			yieldStr := money.ConvertMoneyValue(dividend.GetYieldValue().GetUnits(), dividend.GetYieldValue().GetNano())
			if yieldPercent, err := strconv.ParseFloat(yieldStr, 64); err == nil {
				dbDividend.YieldPercent = &yieldPercent
			}
		}

		result = append(result, dbDividend)
	}

	return result, nil
}
