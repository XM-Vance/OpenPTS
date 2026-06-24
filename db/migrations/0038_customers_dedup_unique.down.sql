-- 0038 回滚：移除唯一约束
ALTER TABLE customers
    DROP CONSTRAINT IF EXISTS customers_user_name_source_unique;
