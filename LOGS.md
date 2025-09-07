# Логирование

Система использует структурированное логирование с помощью logrus. Логи включают:

- Время выполнения операций
- Уровень важности (debug, info, warn, error, fatal)
- Контекстную информацию
- Ошибки и их детали

## Настройки логирования

В файле `config/config.yaml` можно настроить:

```yaml
logging:
  level: "info"        # debug, info, warn, error, fatal
  format: "text"       # text или json
  output: "stdout"   # stdout, file, both
  file_path: "./logs/market-loader.log"
  max_file_size: 100 # МБ
  max_files: 5       # Количество файлов для ротации
```

## Уровни логирования

- **`debug`** - Подробная отладочная информация (для разработки)
- **`info`** - Общая информация о работе (рекомендуется для продакшена)
- **`warn`** - Только предупреждения и ошибки
- **`error`** - Только ошибки
- **`fatal`** - Только критические ошибки

## Форматы вывода

- **`text`** - Человекочитаемый текст (для разработки)
- **`json`** - Структурированный JSON (для систем мониторинга: ELK, Grafana)

## Примеры конфигурации

### Разработка
```yaml
logging:
  level: "debug"
  format: "text"
```

### Продакшен
```yaml
logging:
  level: "warn"
  format: "json"
  output: "file"
  file_path: "/var/log/market-loader/app.log"
  max_file_size: 100
  max_files: 10
```

### Мониторинг
```yaml
logging:
  level: "info"
  format: "json"
  output: "both"
  file_path: "/var/log/market-loader/app.log"
```