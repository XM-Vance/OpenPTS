-- Drop all market data tables
DO $$ DECLARE r RECORD; BEGIN
  FOR r IN SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename LIKE 'md_%' LOOP
    EXECUTE 'DROP TABLE IF EXISTS ' || r.tablename || ' CASCADE';
  END LOOP;
END $$;
