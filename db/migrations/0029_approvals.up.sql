-- 通用审批流：任意资源（合同 / 套餐 / 客户变更等）发起审批 → pending → approved/rejected。
-- 通过 resource + resource_id + status 索引快速查询「我待审批的」。

CREATE TABLE IF NOT EXISTS approval_requests (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    resource      text NOT NULL,                          -- retail_contracts / customers / ...
    resource_id   text NOT NULL,
    title         text NOT NULL,                          -- 摘要（如 "合同 123 变更"）
    payload       jsonb NOT NULL DEFAULT '{}'::jsonb,     -- 变更详情（前后值对比 / 关键字段）
    status        text NOT NULL DEFAULT 'pending'         -- draft / pending / approved / rejected / withdrawn
                   CHECK (status IN ('draft', 'pending', 'approved', 'rejected', 'withdrawn')),
    submitted_by  text NOT NULL,
    reviewed_by   text,
    review_note   text,
    reviewed_at   timestamp,
    created_at    timestamp NOT NULL DEFAULT now(),
    updated_at    timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_approval_requests_status
    ON approval_requests(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_approval_requests_resource
    ON approval_requests(resource, resource_id);
CREATE INDEX IF NOT EXISTS idx_approval_requests_submitter
    ON approval_requests(submitted_by, created_at DESC);
