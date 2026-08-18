-- Все временные метки хранятся как TIMESTAMPTZ (абсолютный момент времени),
-- но SQL-функции вроде now()/date_trunc('day', ...)/CURRENT_DATE интерпретируют
-- "какой сейчас календарный день" в таймзоне текущей сессии. По умолчанию это
-- UTC, из-за чего "сегодня" в отчётах считалось по Гринвичу, а не по Варшаве
-- (пользователь на 1-2 часа впереди UTC получал неверные границы дня).
-- current_database() используется, чтобы не хардкодить имя БД.
DO $$
BEGIN
    EXECUTE format('ALTER DATABASE %I SET timezone TO %L', current_database(), 'Europe/Warsaw');
END $$;

-- Existing users were stored with the old hardcoded default ('UTC') even
-- though the app never actually applied it — align the displayed value with
-- the zone the app now actually uses everywhere (there is still no working
-- per-user timezone picker, see internal/service/entry.go).
UPDATE users SET timezone = 'Europe/Warsaw' WHERE timezone = 'UTC';
