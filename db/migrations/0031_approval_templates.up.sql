-- 审批模板：预设 资源/字段/默认值，提交人选模板即可填充。

CREATE TABLE IF NOT EXISTS approval_templates (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL UNIQUE,
    resource    text NOT NULL,            -- 适用资源类型
    title_tpl   text NOT NULL,            -- 标题模板，可含 {field} {new}
    field       text NOT NULL,            -- 默认变更字段
    description text,
    enabled     boolean NOT NULL DEFAULT true,
    created_at  timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_approval_templates_resource
    ON approval_templates(resource);

-- 预置 4 个模板
INSERT INTO approval_templates (name, resource, title_tpl, field, description) VALUES
    ('合同电量变更',     'retail_contracts', '合同变更：购电量 → {new}',  'purchasing_energy_mwh', '调整合同购电量'),
    ('合同绿电占比变更', 'retail_contracts', '合同变更：绿电占比 → {new}%', 'green_power_ratio',   '调整合同绿电占比'),
    ('合同到期月变更',   'retail_contracts', '合同变更：结束月 → {new}',  'purchase_end_month',    '延期或提前结束购电'),
    ('合同状态变更',     'retail_contracts', '合同变更：状态 → {new}',    'status',                '激活 / 终止 / 到期')
ON CONFLICT (name) DO NOTHING;
