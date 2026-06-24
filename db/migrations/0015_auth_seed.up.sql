-- 0015: 权限系统种子数据
-- 插入：14 个系统模块 → 42 个默认权限点（read/write/delete）→ 4 个系统角色 → 角色权限分配

-- ─── 系统模块 ──────────────────────────────────────
INSERT INTO auth_modules (code, name, menu_group, route_paths, sort_order) VALUES
    ('dashboard',             '仪表盘',         '主页', ARRAY['/dashboard'],                                                10),
    ('user_management',       '用户管理',       '系统', ARRAY['/users', '/roles'],                                          20),
    ('customer_management',   '客户档案管理',   '客户', ARRAY['/customers', '/intent-customers'],                            30),
    ('retail_management',     '零售管理',       '零售', ARRAY['/retail/contracts', '/retail/packages', '/retail/settlement'], 40),
    ('load_management',       '负荷管理',       '负荷', ARRAY['/load/raw', '/load/forecast', '/load/diagnosis'],             50),
    ('price_management',      '价格管理',       '价格', ARRAY['/price/spot', '/price/forecast', '/price/sgcc'],              60),
    ('settlement_management', '结算管理',       '结算', ARRAY['/settlement/daily', '/settlement/monthly'],                   70),
    ('freq_regulation',       '调频管理',       '调频', ARRAY['/freq/clearing', '/freq/demand'],                             80),
    ('storage',               '储能管理',       '储能', ARRAY['/storage/declaration', '/storage/operation'],                 90),
    ('analytics',             '客户分析',       '分析', ARRAY['/analytics/features', '/analytics/alerts'],                  100),
    ('trade_strategy',        '交易策略',       '交易', ARRAY['/trade/strategy', '/trade/review'],                          110),
    ('weather_data',          '气象数据',       '基础', ARRAY['/weather/locations', '/weather/forecast'],                   120),
    ('task_scheduler',        '任务调度',       '系统', ARRAY['/tasks/commands', '/tasks/logs'],                            130),
    ('system',                '系统管理',       '系统', ARRAY['/system/logs', '/system/alerts'],                            140)
ON CONFLICT (code) DO NOTHING;

-- ─── 默认权限点 ────────────────────────────────────
-- 每个模块自动派生 read / write / delete 三个权限点
-- delete 标记为 critical（对应 v1 docs/spec/权限与鉴权规则.md 中的 data:critical:delete）
INSERT INTO auth_permissions (code, name, module_code, action, permission_type)
SELECT
    m.code || ':' || a.action,
    m.name || ' - ' || a.label,
    m.code,
    a.action,
    CASE WHEN a.action = 'delete' THEN 'critical' ELSE 'normal' END
FROM auth_modules m
CROSS JOIN (VALUES
    ('read',   '查看'),
    ('write',  '编辑'),
    ('delete', '删除')
) AS a(action, label)
ON CONFLICT (code) DO NOTHING;

-- ─── 系统角色 ──────────────────────────────────────
INSERT INTO auth_roles (code, name, description, is_system) VALUES
    ('super_admin', '超级管理员', '拥有全部权限（含删除）',                          TRUE),
    ('admin',       '管理员',     '业务管理权限（read + write，不含 delete）',       TRUE),
    ('analyst',     '分析师',     '全模块只读 + 客户/负荷/价格/分析模块可编辑',     TRUE),
    ('viewer',      '只读用户',   '所有模块只读',                                   TRUE)
ON CONFLICT (code) DO NOTHING;

-- ─── 角色权限分配 ──────────────────────────────────

-- super_admin：所有权限
INSERT INTO role_permissions (role_code, permission_code)
SELECT 'super_admin', code FROM auth_permissions
ON CONFLICT DO NOTHING;

-- admin：所有 read + write
INSERT INTO role_permissions (role_code, permission_code)
SELECT 'admin', code FROM auth_permissions
WHERE action IN ('read', 'write')
ON CONFLICT DO NOTHING;

-- analyst：所有 read + 客户/负荷/价格/分析模块的 write
INSERT INTO role_permissions (role_code, permission_code)
SELECT 'analyst', code FROM auth_permissions WHERE action = 'read'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_code, permission_code)
SELECT 'analyst', code FROM auth_permissions
WHERE action = 'write'
  AND module_code IN ('customer_management', 'load_management', 'price_management', 'analytics')
ON CONFLICT DO NOTHING;

-- viewer：所有 read
INSERT INTO role_permissions (role_code, permission_code)
SELECT 'viewer', code FROM auth_permissions WHERE action = 'read'
ON CONFLICT DO NOTHING;
