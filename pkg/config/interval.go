// Package config содержит общие функции и константы для загрузчиков
// Market Loader
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package config

import (
	"fmt"
	"time"

	pb "github.com/russianinvestments/invest-api-go-sdk/proto"
)

// ParseInterval 1min->CANDLE_INTERVAL_1_MIN
func ParseInterval(intervalStr string) (string, error) {
	// Маппинг интервалов
	intervalMap := map[string]string{
		CandleIntervalText1Min:  CandleInterval1Min,
		CandleIntervalText2Min:  CandleInterval2Min,
		CandleIntervalText3Min:  CandleInterval3Min,
		CandleIntervalText5Min:  CandleInterval5Min,
		CandleIntervalText10Min: CandleInterval10Min,
		CandleIntervalText15Min: CandleInterval15Min,
		CandleIntervalText30Min: CandleInterval30Min,
		CandleIntervalTextHour:  CandleIntervalHour,
		CandleIntervalText2Hour: CandleInterval2Hour,
		CandleIntervalText4Hour: CandleInterval4Hour,
		CandleIntervalTextDay:   CandleIntervalDay,
		CandleIntervalTextWeek:  CandleIntervalWeek,
		CandleIntervalTextMonth: CandleIntervalMonth,
	}

	if intervalType, exists := intervalMap[intervalStr]; exists {
		return intervalType, nil
	}

	return "", fmt.Errorf("неподдерживаемый интервал: %s", intervalStr)
}

// Interval2text CANDLE_INTERVAL_1_MIN->1min
func Interval2text(interval string) string {
	// Маппинг интервалов
	var candleIntervalToText = map[string]string{
		CandleInterval1Min:  CandleIntervalText1Min,
		CandleInterval2Min:  CandleIntervalText2Min,
		CandleInterval3Min:  CandleIntervalText3Min,
		CandleInterval5Min:  CandleIntervalText5Min,
		CandleInterval10Min: CandleIntervalText10Min,
		CandleInterval15Min: CandleIntervalText15Min,
		CandleInterval30Min: CandleIntervalText30Min,
		CandleIntervalHour:  CandleIntervalTextHour,
		CandleInterval2Hour: CandleIntervalText2Hour,
		CandleInterval4Hour: CandleIntervalText4Hour,
		CandleIntervalDay:   CandleIntervalTextDay,
		CandleIntervalWeek:  CandleIntervalTextWeek,
		CandleIntervalMonth: CandleIntervalTextMonth,
	}

	text := candleIntervalToText[interval]
	return text
}

// GetCandleInterval конвертирует строковый интервал в protobuf тип
func GetCandleInterval(intervalType string) pb.CandleInterval {
	switch intervalType {
	case CandleInterval1Min:
		return pb.CandleInterval_CANDLE_INTERVAL_1_MIN
	case CandleInterval2Min:
		return pb.CandleInterval_CANDLE_INTERVAL_2_MIN
	case CandleInterval3Min:
		return pb.CandleInterval_CANDLE_INTERVAL_3_MIN
	case CandleInterval5Min:
		return pb.CandleInterval_CANDLE_INTERVAL_5_MIN
	case CandleInterval10Min:
		return pb.CandleInterval_CANDLE_INTERVAL_10_MIN
	case CandleInterval15Min:
		return pb.CandleInterval_CANDLE_INTERVAL_15_MIN
	case CandleInterval30Min:
		return pb.CandleInterval_CANDLE_INTERVAL_30_MIN
	case CandleIntervalHour:
		return pb.CandleInterval_CANDLE_INTERVAL_HOUR
	case CandleInterval2Hour:
		return pb.CandleInterval_CANDLE_INTERVAL_2_HOUR
	case CandleInterval4Hour:
		return pb.CandleInterval_CANDLE_INTERVAL_4_HOUR
	case CandleIntervalDay:
		return pb.CandleInterval_CANDLE_INTERVAL_DAY
	case CandleIntervalWeek:
		return pb.CandleInterval_CANDLE_INTERVAL_WEEK
	case CandleIntervalMonth:
		return pb.CandleInterval_CANDLE_INTERVAL_MONTH
	case "":
		return pb.CandleInterval_CANDLE_INTERVAL_1_MIN
	default:
		return pb.CandleInterval_CANDLE_INTERVAL_1_MIN
	}
}

// GetCandleIntervalString конвертирует protobuf тип в строковый интервал
//
//nolint:exhaustive
func GetCandleIntervalString(interval pb.CandleInterval) string {
	switch interval {
	case pb.CandleInterval_CANDLE_INTERVAL_1_MIN:
		return CandleInterval1Min
	case pb.CandleInterval_CANDLE_INTERVAL_2_MIN:
		return CandleInterval2Min
	case pb.CandleInterval_CANDLE_INTERVAL_3_MIN:
		return CandleInterval3Min
	case pb.CandleInterval_CANDLE_INTERVAL_5_MIN:
		return CandleInterval5Min
	case pb.CandleInterval_CANDLE_INTERVAL_10_MIN:
		return CandleInterval10Min
	case pb.CandleInterval_CANDLE_INTERVAL_15_MIN:
		return CandleInterval15Min
	case pb.CandleInterval_CANDLE_INTERVAL_30_MIN:
		return CandleInterval30Min
	case pb.CandleInterval_CANDLE_INTERVAL_HOUR:
		return CandleIntervalHour
	case pb.CandleInterval_CANDLE_INTERVAL_2_HOUR:
		return CandleInterval2Hour
	case pb.CandleInterval_CANDLE_INTERVAL_4_HOUR:
		return CandleInterval4Hour
	case pb.CandleInterval_CANDLE_INTERVAL_DAY:
		return CandleIntervalDay
	case pb.CandleInterval_CANDLE_INTERVAL_WEEK:
		return CandleIntervalWeek
	case pb.CandleInterval_CANDLE_INTERVAL_MONTH:
		return CandleIntervalMonth
	default:
		return CandleInterval1Min
	}
}

// CalculateChunkSize вычисляет размер чанка
func CalculateChunkSize(intervalType string, apiLimit int) time.Duration {
	return GetThreshold(intervalType) * time.Duration(apiLimit)
}

// ShouldUpdateData проверяет, нужно ли обновлять данные для заданного интервала
func ShouldUpdateData(lastLoadedTime time.Time, intervalType string) bool {
	// Определяем порог обновления в зависимости от интервала
	return time.Since(lastLoadedTime) > GetThreshold(intervalType)
}

// GetDateFormat определяет формат даты для логирования в зависимости от интервала
func GetDateFormat(intervalType string) string {
	switch intervalType {
	case CandleInterval1Min, CandleInterval2Min, CandleInterval3Min, CandleInterval5Min,
		CandleInterval10Min, CandleInterval15Min, CandleInterval30Min,
		CandleIntervalHour, CandleInterval2Hour, CandleInterval4Hour:
		// Для минутных и часовых интервалов показываем время
		return "2006-01-02 15:04"
	case CandleIntervalDay, CandleIntervalWeek, CandleIntervalMonth:
		// Для дневных и более длинных интервалов показываем только дату
		return "2006-01-02"
	default:
		// По умолчанию показываем время
		return "2006-01-02 15:04"
	}
}

// IsNormalTrading проверяет, что инструмент находится в нормальном торговом статусе
func IsNormalTrading(status pb.SecurityTradingStatus) bool {
	return status == pb.SecurityTradingStatus_SECURITY_TRADING_STATUS_NORMAL_TRADING
}

// ConvertMinPriceIncrement конвертирует Quotation в float64 для MinPriceIncrement
func ConvertMinPriceIncrement(quotation *pb.Quotation) float64 {
	return float64(quotation.Units) + float64(quotation.Nano)/1e9
}

// GetTimeUnitAndConfigKey определяет единицу времени и ключ конфигурации по типу интервала
func GetTimeUnitAndConfigKey(intervalType string) (time.Duration, string) {
	switch intervalType {
	case CandleInterval1Min:
		return time.Minute, CandleIntervalText1Min
	case CandleInterval2Min:
		return time.Minute, CandleIntervalText1Min
	case CandleInterval3Min:
		return time.Minute, CandleIntervalText1Min
	case CandleInterval5Min:
		return time.Minute, CandleIntervalText1Min
	case CandleInterval10Min:
		return time.Minute, CandleIntervalText1Min
	case CandleInterval15Min:
		return time.Minute, CandleIntervalText1Min
	case CandleInterval30Min:
		return time.Minute, CandleIntervalText1Min
	case CandleIntervalHour:
		return time.Minute, CandleIntervalText1Min
	case CandleInterval2Hour:
		return time.Hour, CandleIntervalTextHour
	case CandleInterval4Hour:
		return time.Hour, CandleIntervalTextHour
	case CandleIntervalDay:
		return time.Duration(HoursInDay) * time.Hour, CandleIntervalTextDay
	case CandleIntervalWeek:
		return time.Duration(DaysInWeek*HoursInDay) * time.Hour, CandleIntervalTextWeek
	case CandleIntervalMonth:
		return time.Duration(DaysInMonth*HoursInDay) * time.Hour, CandleIntervalTextMonth
	default:
		// По умолчанию используем минуту
		return DefaultUpdateThreshold, CandleIntervalText1Min
	}
}

// GetThreshold получает порог обновления для конкретного интервала
func GetThreshold(intervalType string) time.Duration {
	duration, _ := GetTimeUnitAndConfigKey(intervalType)
	return duration
}
