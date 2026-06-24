ALTER TABLE documents DROP CONSTRAINT IF EXISTS fk_doc_customer;
ALTER TABLE documents DROP CONSTRAINT IF EXISTS fk_doc_contract;
ALTER TABLE documents DROP CONSTRAINT IF EXISTS fk_doc_intent;
ALTER TABLE documents DROP COLUMN IF EXISTS customer_id;
ALTER TABLE documents DROP COLUMN IF EXISTS contract_id;
ALTER TABLE documents DROP COLUMN IF EXISTS intent_customer_id;
