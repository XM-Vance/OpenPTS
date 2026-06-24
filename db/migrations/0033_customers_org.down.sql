DROP INDEX IF EXISTS idx_customers_org;
ALTER TABLE customers DROP COLUMN IF EXISTS org_id;
