-- 74: 电站关联客户
ALTER TABLE solar_stations ADD COLUMN IF NOT EXISTS customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_solar_customer ON solar_stations(customer_id);


