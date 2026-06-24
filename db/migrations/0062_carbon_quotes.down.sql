-- 回滚碳交易行情表。
-- md_carbon_eua 原表未被删除，EUA 数据仍在原表中，故此处直接删除新表即可。
DROP TABLE IF EXISTS carbon_quotes;
