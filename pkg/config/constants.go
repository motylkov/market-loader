// Package config содержит общие функции и константы для загрузчиков
// Market Loader
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package config

import "time"

const (
	// Константы для временных интервалов

	// DefaultYearsBack  количество лет для загрузки данных
	DefaultYearsBack = 5
	// DefaultRetryDelay задержка между повторными попытками
	DefaultRetryDelay = 5 * time.Second
	// DefaultHTTPTimeout таймаут HTTP-запросов по умолчанию
	DefaultHTTPTimeout = 30 * time.Second
	// DefaultUpdateThreshold минимальный порог времени для решения, что данные устарели
	DefaultUpdateThreshold = 1 * time.Minute
	// MinutesInHour количество минут в часе
	MinutesInHour = 60
	// HoursInDay количество часов в сутках
	HoursInDay = 24
	// DaysInWeek количество дней в неделе
	DaysInWeek = 7
	// DaysInMonth количество дней в месяце (условное значение для расчётов)
	DaysInMonth = 30
	// MinutesInDay количество минут в сутках
	MinutesInDay = HoursInDay * MinutesInHour
	// Interval1Min интервал 1 минута
	Interval1Min = 1
	// Interval2Min интервал 2 минуты
	Interval2Min = 2
	// Interval3Min интервал 3 минуты
	Interval3Min = 3
	// Interval5Min интервал 5 минут
	Interval5Min = 5
	// Interval10Min интервал 10 минут
	Interval10Min = 10
	// Interval15Min интервал 15 минут
	Interval15Min = 15
	// Interval30Min интервал 30 минут
	Interval30Min = 30
	// Interval1Hour интервал 1 час
	Interval1Hour = 1
	// Interval2Hour интервал 2 часа
	Interval2Hour = 2
	// Interval4Hour интервал 4 часа
	Interval4Hour = 4

	// Константы интервалов свечей

	// CandleInterval1Min интервал свечей 1 минута
	CandleInterval1Min = "CANDLE_INTERVAL_1_MIN"
	// CandleInterval2Min интервал свечей 2 минуты
	CandleInterval2Min = "CANDLE_INTERVAL_2_MIN"
	// CandleInterval3Min интервал свечей 3 минуты
	CandleInterval3Min = "CANDLE_INTERVAL_3_MIN"
	// CandleInterval5Min интервал свечей 5 минут
	CandleInterval5Min = "CANDLE_INTERVAL_5_MIN"
	// CandleInterval10Min интервал свечей 10 минут
	CandleInterval10Min = "CANDLE_INTERVAL_10_MIN"
	// CandleInterval15Min интервал свечей 15 минут
	CandleInterval15Min = "CANDLE_INTERVAL_15_MIN"
	// CandleInterval30Min интервал свечей 30 минут
	CandleInterval30Min = "CANDLE_INTERVAL_30_MIN"

	// Часовые интервалы

	// CandleIntervalHour интервал свечей 1 час
	CandleIntervalHour = "CANDLE_INTERVAL_HOUR"
	// CandleInterval2Hour интервал свечей 2 часа
	CandleInterval2Hour = "CANDLE_INTERVAL_2_HOUR"
	// CandleInterval4Hour интервал свечей 4 часа
	CandleInterval4Hour = "CANDLE_INTERVAL_4_HOUR"

	// Дневные и более длинные интервалы

	// CandleIntervalDay интервал свечей 1 день
	CandleIntervalDay = "CANDLE_INTERVAL_DAY"
	// CandleIntervalWeek интервал свечей 1 неделя
	CandleIntervalWeek = "CANDLE_INTERVAL_WEEK"
	// CandleIntervalMonth интервал свечей 1 месяц
	CandleIntervalMonth = "CANDLE_INTERVAL_MONTH"

	// Интервалы текстом

	// CandleIntervalText1Min текстовый интервал 1 минута
	CandleIntervalText1Min = "1min"
	// CandleIntervalText2Min текстовый интервал 2 минуты
	CandleIntervalText2Min = "2min"
	// CandleIntervalText3Min текстовый интервал 3 минуты
	CandleIntervalText3Min = "3min"
	// CandleIntervalText5Min текстовый интервал 5 минут
	CandleIntervalText5Min = "5min"
	// CandleIntervalText10Min текстовый интервал 10 минут
	CandleIntervalText10Min = "10min"
	// CandleIntervalText15Min текстовый интервал 15 минут
	CandleIntervalText15Min = "15min"
	// CandleIntervalText30Min текстовый интервал 30 минут
	CandleIntervalText30Min = "30min"
	// CandleIntervalTextHour текстовый интервал 1 час
	CandleIntervalTextHour = "1hour"
	// CandleIntervalText2Hour текстовый интервал2 часа
	CandleIntervalText2Hour = "2hour"
	// CandleIntervalText4Hour текстовый интервал 4 часа
	CandleIntervalText4Hour = "4hour"
	// CandleIntervalTextDay текстовый интервал 1 день
	CandleIntervalTextDay = "1day"
	// CandleIntervalTextWeek текстовый интервал 1 неделя
	CandleIntervalTextWeek = "1week"
	// CandleIntervalTextMonth текстовый интервал 1 месяц
	CandleIntervalTextMonth = "1month"

	// Shares обозначает тип инструмента «акции»
	Shares = "share"

	// MinCSVFields минимально число полей в CSV-строке
	MinCSVFields = 7
	// MaxFractionDigits максимальное число знаков после запятой
	MaxFractionDigits = 9
	// MaxNanoDigits максимальное число знаков для наносекунд
	MaxNanoDigits = 9
	// DefaultDirPerm права доступа создаваемых директорий
	DefaultDirPerm = 0750
)
