ALTER TABLE user_load_data
    ALTER COLUMN curve_96 TYPE numeric(14,4)[] USING curve_96::numeric(14,4)[];
