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
	"time"
)

// GetIntervalLimit получает лимит для конкретного интервала
func (c *Config) GetIntervalLimit(interval string) int {
	if limit, exists := c.Loading.Limits[interval]; exists {
		return limit
	}
	// Значение по умолчанию
	return MinutesInDay
}

// GetStartDate получает дату начала загрузки данных
func (c *Config) GetStartDate() time.Time {
	if c.Loading.StartDate == "" {
		// По умолчанию 5 лет назад
		return time.Now().AddDate(-5, 0, 0)
	}

	// Парсим дату из конфигурации
	startDate, err := time.Parse("2006-01-02", c.Loading.StartDate)
	if err != nil {
		// В случае ошибки парсинга возвращаем 5 лет назад
		return time.Now().AddDate(-5, 0, 0)
	}

	return startDate
}
