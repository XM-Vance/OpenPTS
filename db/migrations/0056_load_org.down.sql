DROP INDEX IF EXISTS user_load_data_org_uniq;
DROP INDEX IF EXISTS idx_user_load_data_org;
ALTER TABLE user_load_data DROP COLUMN IF EXISTS org_id;
ALTER TABLE user_load_data ADD CONSTRAINT user_load_data_customer_id_date_key UNIQUE (customer_id, date);
