-- 0050: 多租户 · 初始化一个默认组织，作为开发/演示环境的根租户。
-- 多租户体系支持多个组织（如按地区/省份划分），这里仅插入一个示例默认组织，
-- 二次部署时可按需添加你自己的组织记录（organizations 表 code 唯一、幂等）。
INSERT INTO organizations (code, name) VALUES
  ('default', '默认组织')
ON CONFLICT (code) DO NOTHING;
