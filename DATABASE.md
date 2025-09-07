# Структура базы данных Market Loader

## Обзор

База данных Market Loader построена на PostgreSQL с использованием следующих возможностей:

- **Партиционирование** для оптимизации больших объемов данных
- **Индексы** для ускорения запросов
- **Внешние ключи** для обеспечения целостности данных
- **Транзакционность** для надежности операций

## Схема базы данных

### Основные таблицы

#### 1. Таблица `instruments`

Справочник всех доступных инструментов (акции, облигации, ETF).

```sql
CREATE TABLE instruments (
			figi varchar(50) NOT NULL,
			ticker varchar(30) NOT NULL,
			name text NOT NULL,
			instrument_type varchar(20) NOT NULL,
			currency varchar(3) NOT NULL,
			lot_size int4 NOT NULL,
			min_price_increment numeric(20, 9) NOT NULL,
			trading_status varchar(40) NOT NULL,
			enabled bool DEFAULT false NOT NULL,
			created_at timestamp DEFAULT now() NOT NULL,
			updated_at timestamp DEFAULT now() NOT NULL,
			last_loaded_time timestamp NULL, -- только для информации
			CONSTRAINT instruments_pkey PRIMARY KEY (figi)
);
```

**Поля:**
- `figi` - уникальный идентификатор инструмента (первичный ключ)
- `ticker` - тикер инструмента (например, "SBER")
- `name` - полное название инструмента
- `instrument_type` - тип инструмента (share, bond, etf)
- `currency` - валюта инструмента (RUB, USD, EUR)
- `lot_size` - размер лота
- `min_price_increment` - минимальный шаг цены
- `trading_status` - статус торговли
- `created_at` - дата создания записи
- `updated_at` - дата последнего обновления
- `last_loaded_time` - дата последней загрузки свечей (только для информации)

**Индексы:**
```sql
CREATE INDEX idx_instruments_ticker ON instruments(ticker);
CREATE INDEX idx_instruments_type ON instruments(instrument_type);
```

#### 2. Таблица `candles` (партиционированная)

Исторические данные свечей по всем временным интервалам.

```sql
CREATE TABLE candles (
			id BIGSERIAL,
			figi VARCHAR(50) NOT NULL,
			time TIMESTAMP NOT NULL,
			open_price DECIMAL(20, 9) NOT NULL,
			high_price DECIMAL(20, 9) NOT NULL,
			low_price DECIMAL(20, 9) NOT NULL,
			close_price DECIMAL(20, 9) NOT NULL,
			volume BIGINT NOT NULL,
			interval_type VARCHAR(30) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (figi, time, interval_type)
) PARTITION BY RANGE ("time");
```

**Поля:**
- `figi` - идентификатор инструмента (внешний ключ)
- `time` - время свечи
- `open_price` - цена открытия
- `high_price` - максимальная цена
- `low_price` - минимальная цена
- `close_price` - цена закрытия
- `volume` - объем торгов
- `interval_type` - тип интервала (1min, 5min, 1hour, 1day, etc.)
- `created_at` - дата создания записи

**Партиционирование:**
- Партиции создаются по месяцам
- Название партиции: `candles_YYYY_MM`
- Диапазон: с первого дня месяца до последнего

**Индексы:**
```sql
CREATE INDEX idx_candles_figi_interval ON candles(figi, interval_type);
CREATE INDEX idx_candles_time ON candles(time);
```

#### 3. Таблица `dividends`

Данные о дивидендных выплатах по акциям.

```sql
			id BIGSERIAL,
			figi VARCHAR(50) NOT NULL,
			payment_date TIMESTAMPTZ NOT NULL,
			declared_date TIMESTAMPTZ NULL,
			amount NUMERIC(20, 10) NOT NULL,
			currency VARCHAR(3) NULL,
			yield_percent NUMERIC(5, 2) NULL,
			created_at TIMESTAMPTZ DEFAULT NOW() NULL,
			PRIMARY KEY (id),
			UNIQUE (figi, payment_date)
);
```

**Поля:**
- `id` - автоинкрементный идентификатор
- `figi` - идентификатор инструмента (внешний ключ)
- `payment_date` - дата выплаты дивидендов
- `declared_date` - дата объявления дивидендов
- `amount` - сумма дивидендов на акцию
- `currency` - валюта дивидендов
- `yield_percent` - доходность в процентах
- `created_at` - дата создания записи

**Индексы:**
```sql
CREATE INDEX idx_dividends_figi ON dividends(figi);
CREATE INDEX idx_dividends_payment_date ON dividends(payment_date);
```

## Связи между таблицами

### Внешние ключи

```sql
-- Связь candles -> instruments
ALTER TABLE candles ADD CONSTRAINT candles_figi_fkey 
    FOREIGN KEY (figi) REFERENCES instruments(figi) 
    ON UPDATE CASCADE ON DELETE CASCADE;

-- Связь dividends -> instruments
ALTER TABLE dividends ADD CONSTRAINT dividends_figi_fkey 
    FOREIGN KEY (figi) REFERENCES instruments(figi) 
    ON UPDATE CASCADE ON DELETE CASCADE;
```

## Партиционирование

### Принципы партиционирования

1. **По времени** - партиции создаются по месяцам
2. **Автоматическое создание** - новые партиции создаются автоматически
3. **Оптимизация запросов** - запросы по времени работают быстрее

### Создание партиций

```sql
-- Пример создания партиции для января 2025
CREATE TABLE candles_2025_01 PARTITION OF candles
    FOR VALUES FROM ('2025-01-01 00:00:00') TO ('2025-01-31 23:59:59');
```

### Управление партициями

- **Создание**: Автоматически при первом обращении к месяцу
- **Удаление**: Старые партиции можно удалять для экономии места
- **Архивирование**: Партиции можно архивировать в отдельные таблицы

## Индексы и оптимизация

### Рекомендации по индексам

```sql
-- Для запросов по инструменту и интервалу
CREATE INDEX idx_candles_figi_interval_time ON candles(figi, interval_type, time);

-- Для запросов по временному диапазону
CREATE INDEX idx_candles_time_interval ON candles(time, interval_type);

-- Для запросов дивидендов по дате
CREATE INDEX idx_dividends_payment_date_figi ON dividends(payment_date, figi);
```

## Типы данных

### Денежные значения
- `DECIMAL(20, 9)` - для цен (высокая точность)
- `NUMERIC(20, 10)` - для дивидендов
- `NUMERIC(5, 2)` - для процентов

### Временные метки
- `TIMESTAMP` - для времени свечей
- `TIMESTAMPTZ` - для дивидендов (с часовым поясом)

### Идентификаторы
- `VARCHAR(50)` - для FIGI инструментов
- `VARCHAR(30)` - для тикеров

## Производительность

### Мониторинг производительности

```sql
-- Анализ использования индексов
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Размер таблиц и партиций
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename))
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## Безопасность

### Права доступа

```sql
-- Создание пользователя с минимальными правами
CREATE USER t_invest_user WITH PASSWORD 'secure_password';

-- Предоставление необходимых прав
GRANT CONNECT ON DATABASE t_invest_data TO t_invest_user;
GRANT ALL PRIVILEGES ON DATABASE t_invest_data TO t_invest_user;
GRANT ALL PRIVILEGES ON SCHEMA public TO t_invest_user;
```

### Резервное копирование

1. **Полные бэкапы** - еженедельно
2. **Инкрементальные** - ежедневно
3. **WAL архивирование** - для point-in-time recovery

## Обслуживание

### Регулярные задачи

1. **VACUUM** - очистка мертвых записей
2. **ANALYZE** - обновление статистики
3. **REINDEX** - пересоздание индексов
4. **Удаление старых партиций** - для экономии места

### Автоматизация

```sql
-- Создание функции для автоматического создания партиций
CREATE OR REPLACE FUNCTION create_candles_partition(partition_date DATE)
RETURNS VOID AS $$
DECLARE
    partition_name TEXT;
    start_date TIMESTAMP;
    end_date TIMESTAMP;
BEGIN
    partition_name := 'candles_' || TO_CHAR(partition_date, 'YYYY_MM');
    start_date := DATE_TRUNC('month', partition_date);
    end_date := start_date + INTERVAL '1 month' - INTERVAL '1 second';
    
    EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF candles
                    FOR VALUES FROM (%L) TO (%L)',
                    partition_name, start_date, end_date);
END;
$$ LANGUAGE plpgsql;
```
