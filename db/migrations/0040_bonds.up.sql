-- 0040: 保函管理表
CREATE TABLE IF NOT EXISTS bonds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    bond_type VARCHAR(50) NOT NULL DEFAULT '',
    amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    issuer VARCHAR(200) NOT NULL DEFAULT '',
    beneficiary VARCHAR(200) NOT NULL DEFAULT '',
    issue_date DATE,
    expire_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    description TEXT NOT NULL DEFAULT '',
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_bonds_status ON bonds (status);
CREATE INDEX IF NOT EXISTS idx_bonds_expire_date ON bonds (expire_date);
