-- примеры SQL запросов

-- По умолчанию все инструменты имеют enabled = false
-- Примеры включения инструментов:

-- 1. Включить все акции определенного типа (например, только российские)
-- UPDATE instruments 
-- SET enabled = true 
-- WHERE instrument_type = 'share' 
--   AND currency = 'rub'
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING';

-- 2. Включить конкретные тикеры
-- UPDATE instruments 
-- SET enabled = true 
-- WHERE ticker IN ('SBER', 'GAZP', 'LKOH', 'YNDX', 'TCSG')
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING';

-- 3. Включить инструменты по FIGI
-- UPDATE instruments 
-- SET enabled = true 
-- WHERE figi IN ('BBG000B9XRY4', 'BBG000B9XRY5', 'BBG000B9XRY6')
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING';

-- 4. Включить все ETF
-- UPDATE instruments 
-- SET enabled = true 
-- WHERE instrument_type = 'etf'
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING';

-- 5. Включить все облигации федерального займа (ОФЗ)
-- UPDATE instruments 
-- SET enabled = true 
-- WHERE instrument_type = 'bond'
--   AND name LIKE '%ОФЗ%'
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING';

-- 6. Включить топ-20 акций по объему торгов (пример)
-- UPDATE instruments 
-- SET enabled = true 
-- WHERE figi IN (
--   SELECT figi FROM (
--     SELECT figi, SUM(volume) as total_volume
--     FROM candles 
--     WHERE time >= NOW() - INTERVAL '30 days'
--     GROUP BY figi
--     ORDER BY total_volume DESC
--     LIMIT 20
--   ) top_instruments
-- );

-- 7. Включить все инструменты определенной валюты
-- UPDATE instruments 
-- SET enabled = true 
-- WHERE currency = 'usd'
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING';

-- 8. Включить инструменты по диапазону тикеров
-- UPDATE instruments 
-- SET enabled = true 
-- WHERE ticker >= 'A' AND ticker <= 'M'
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING';

-- 9. Включить инструменты по ISIN коду
-- UPDATE instruments 
-- SET enabled = true 
-- WHERE isin IN ('RU0009029540', 'RU0009029541', 'RU0009029542')
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING';

-- 10. Включить инструменты с возможностью шорта
-- UPDATE instruments 
-- SET enabled = true 
-- WHERE short_enabled_flag = true
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING';

-- 11. Включить инструменты с IPO после определенной даты
-- UPDATE instruments 
-- SET enabled = true 
-- WHERE ipo_date >= '2020-01-01'
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING';
--------------------------------------------------------------------

-- Проверка текущего состояния:

-- SELECT 
--   instrument_type,
--   currency,
--   COUNT(*) as total,
--   COUNT(CASE WHEN enabled = true THEN 1 END) as enabled_count,
--   COUNT(CASE WHEN enabled = false THEN 1 END) as disabled_count
-- FROM instruments 
-- WHERE trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING'
-- GROUP BY instrument_type, currency
-- ORDER BY instrument_type, currency;

-- Показать включенные инструменты:

-- SELECT ticker, name, instrument_type, currency, enabled
-- FROM instruments 
-- WHERE enabled = true 
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING'
-- ORDER BY instrument_type, ticker;

-- Показать отключенные инструменты:

-- SELECT ticker, name, instrument_type, currency, enabled
-- FROM instruments 
-- WHERE enabled = false 
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING'
-- ORDER BY instrument_type, ticker
-- LIMIT 100;

-- Статистика по секторам:

-- SELECT 
--   sector,
--   COUNT(*) as total_instruments,
--   COUNT(CASE WHEN enabled = true THEN 1 END) as enabled_instruments,
--   AVG(issue_size) as avg_issue_size
-- FROM instruments 
-- WHERE trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING'
--   AND sector IS NOT NULL
-- GROUP BY sector
-- ORDER BY total_instruments DESC;

-- Инструменты с возможностью шорта:

-- SELECT 
--   ticker, 
--   name, 
--   instrument_type, 
--   sector,
--   short_enabled_flag,
--   enabled
-- FROM instruments 
-- WHERE short_enabled_flag = true
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING'
-- ORDER BY instrument_type, ticker;

-- Поиск инструментов по ISIN:

-- SELECT 
--   ticker, 
--   name, 
--   isin,
--   instrument_type,
--   sector,
--   enabled
-- FROM instruments 
-- WHERE isin LIKE '%RU0009029540%'
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING';

-- Инструменты по размеру выпуска (топ-20):
-- SELECT 
--   ticker, 
--   name, 
--   issue_size,
--   instrument_type,
--   enabled
-- FROM instruments 
-- WHERE issue_size IS NOT NULL
--   AND trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING'
-- ORDER BY issue_size DESC
-- LIMIT 20;
