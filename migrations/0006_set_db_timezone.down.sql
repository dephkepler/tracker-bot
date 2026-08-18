UPDATE users SET timezone = 'UTC' WHERE timezone = 'Europe/Warsaw';

DO $$
BEGIN
    EXECUTE format('ALTER DATABASE %I SET timezone TO DEFAULT', current_database());
END $$;
