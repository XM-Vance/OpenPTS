-- 99 回滚：删除统一时序表（仅在一次性库验证可逆性，生产/ptis_dev 禁止 down）。

DROP TABLE IF EXISTS md_time_series;
