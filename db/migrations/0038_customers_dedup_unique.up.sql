-- 0038: customers 去重 + UNIQUE(user_name, source) 约束
-- 确保 (user_name, source) 组合唯一，支持迁移脚本幂等写入

-- 1) 删除重复行，保留每组 (user_name, source) 中 id 最小的一行
DELETE FROM customers a
USING customers b
WHERE a.user_name = b.user_name
  AND a.source   = b.source
  AND a.id > b.id;

-- 2) 添加唯一约束
ALTER TABLE customers
    ADD CONSTRAINT customers_user_name_source_unique
    UNIQUE (user_name, source);
