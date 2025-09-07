// Package arch содержит функции для работы с архивом свечей
// Market Loader
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package arch

import (
	"market-loader/pkg/config"
	"strconv"
	"strings"

	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
)

// parsePriceString точно парсит строку цены в pb.Quotation
func parsePriceString(priceStr string) *pb.Quotation {
	// Убираем пробелы
	priceStr = strings.TrimSpace(priceStr)

	// Ищем точку
	dotIndex := strings.Index(priceStr, ".")
	if dotIndex == -1 {
		// Нет дробной части
		if units, err := strconv.ParseInt(priceStr, 10, 64); err == nil {
			return &pb.Quotation{
				Units: units,
				Nano:  0,
			}
		}
		return &pb.Quotation{Units: 0, Nano: 0}
	}

	// Есть дробная часть
	unitsStr := priceStr[:dotIndex]
	fractionStr := priceStr[dotIndex+1:]

	// Парсим целую часть
	units, err := strconv.ParseInt(unitsStr, 10, 64)
	if err != nil {
		return &pb.Quotation{Units: 0, Nano: 0}
	}

	// Обрабатываем дробную часть
	if len(fractionStr) == 0 {
		return &pb.Quotation{Units: units, Nano: 0}
	}

	// Дополняем дробную часть до 9 цифр
	for len(fractionStr) < 9 {
		fractionStr += "0"
	}

	// Обрезаем до 9 цифр
	if len(fractionStr) > config.MaxNanoDigits {
		fractionStr = fractionStr[:config.MaxNanoDigits]
	}

	// Парсим nano
	nano, err := strconv.ParseInt(fractionStr, 10, 32)
	if err != nil {
		return &pb.Quotation{Units: units, Nano: 0}
	}

	return &pb.Quotation{
		Units: units,
		Nano:  int32(nano),
	}
}
