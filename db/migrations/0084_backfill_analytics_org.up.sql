-- 多租户隔离收口：customer_anomaly_alerts / customer_characteristics 此前写入未带 org_id，
-- 存量行 org_id 为 NULL（加省过滤后会"消失"）。按 customer_id 回填其所属省。
-- 客户单省，故 customer_id → customers.org_id 唯一确定。
UPDATE customer_anomaly_alerts a
   SET org_id = c.org_id
  FROM customers c
 WHERE a.customer_id = c.id AND a.org_id IS NULL;

UPDATE customer_characteristics cc
   SET org_id = c.org_id
  FROM customers c
 WHERE cc.customer_id = c.id AND cc.org_id IS NULL;
