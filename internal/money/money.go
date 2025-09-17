// Package money содержит функции для корректного преобразования денежных форматов
// Market Loader
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package money

import (
	"fmt"

	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
)

// ConvertMoneyValue точно конвертирует денежное значение из API
// избегая проблем с плавающей точкой
func ConvertMoneyValue(units int64, nano int32) string {
	if nano == 0 {
		return fmt.Sprintf("%d", units)
	}

	// Преобразуем nano в строку с ведущими нулями
	nanoStr := fmt.Sprintf("%09d", nano)

	// Убираем trailing zeros
	for len(nanoStr) > 0 && nanoStr[len(nanoStr)-1] == '0' {
		nanoStr = nanoStr[:len(nanoStr)-1]
	}

	if len(nanoStr) == 0 {
		return fmt.Sprintf("%d", units)
	}

	return fmt.Sprintf("%d.%s", units, nanoStr)
}

// ConvertMinPriceIncrement конвертирует Quotation в float64 для MinPriceIncrement
func ConvertMinPriceIncrement(quotation *pb.Quotation) float64 {
	return float64(quotation.Units) + float64(quotation.Nano)/1e9
}

func ConvertQuotationToFloat(q *pb.Quotation) float64 {
	if q == nil {
		return 0
	}
	return float64(q.Units) + float64(q.Nano)/1e9
}

func ConvertMoneyValueToFloat(m *pb.MoneyValue) float64 {
	if m == nil {
		return 0.0
	}
	return float64(m.Units) + float64(m.Nano)/1e9
}
