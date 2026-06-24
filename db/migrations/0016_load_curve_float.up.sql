-- 0016: user_load_data.curve_96 由 numeric(14,4)[] 改为 double precision[]
-- 原因：pgx 对 numeric[] 数组类型的读写转换不可靠；负荷曲线值用 float8 精度足够。
ALTER TABLE user_load_data
    ALTER COLUMN curve_96 TYPE double precision[] USING curve_96::double precision[];
