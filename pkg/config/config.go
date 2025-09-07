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
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DatabaseConfig структура конфигурации базы данных
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// Config структура конфигурации
type Config struct {
	Database DatabaseConfig `yaml:"database"`

	Tinvest struct {
		Token    string `yaml:"token"`
		Endpoint string `yaml:"endpoint"`
		AppName  string `yaml:"app_name"`
	} `yaml:"tinvest"`

	Loading struct {
		StartDate      string         `yaml:"start_date"`
		Limits         map[string]int `yaml:"limits"`
		RateLimitPause int            `yaml:"rate_limit_pause"`
	} `yaml:"loading"`

	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`

	// Настройки для архивного загрузчика
	Archive struct {
		TempDir string `yaml:"temp_dir"`
	} `yaml:"archive"`
}

// LoadConfig загружает конфигурацию из YAML файла
func LoadConfig(path string) (*Config, error) {
	// Читаем файл
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл конфигурации %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга YAML: %w", err)
	}

	return &cfg, nil
}

// GetConfigPath определяет путь к файлу конфигурации
func GetConfigPath() string {
	// Получаем путь к исполняемому файлу
	execPath, err := os.Executable()
	if err != nil {
		// Если не удалось получить путь к исполняемому файлу, используем относительный путь
		return "config/config.yaml"
	}

	// Определяем директорию исполняемого файла
	execDir := filepath.Dir(execPath)

	// Если исполняемый файл в папке bin, то конфиг на один уровень выше
	if filepath.Base(execDir) == "bin" {
		return filepath.Join(filepath.Dir(execDir), "config", "config.yaml")
	}

	// Иначе используем относительный путь (для go run)
	return "config/config.yaml"
}
