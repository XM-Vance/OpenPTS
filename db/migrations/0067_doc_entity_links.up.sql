-- 67: 文档关联业务实体
ALTER TABLE documents ADD COLUMN IF NOT EXISTS customer_id UUID;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS contract_id UUID;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS intent_customer_id UUID;

ALTER TABLE documents ADD CONSTRAINT fk_doc_customer
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE SET NULL;
ALTER TABLE documents ADD CONSTRAINT fk_doc_contract
    FOREIGN KEY (contract_id) REFERENCES retail_contracts(id) ON DELETE SET NULL;
ALTER TABLE documents ADD CONSTRAINT fk_doc_intent
    FOREIGN KEY (intent_customer_id) REFERENCES intent_customers(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_doc_customer ON documents(customer_id);
CREATE INDEX IF NOT EXISTS idx_doc_contract ON documents(contract_id);
CREATE INDEX IF NOT EXISTS idx_doc_intent ON documents(intent_customer_id);

COMMENT ON COLUMN documents.customer_id IS '关联客户（合同/账单等文档归属的客户）';
COMMENT ON COLUMN documents.contract_id IS '关联零售合同';
COMMENT ON COLUMN documents.intent_customer_id IS '关联意向客户';
