-- 迁移 #80: 菜单页面级权限 — 支持后台可视化配置每个角色可见的菜单

-- ═══ 页面注册表（系统内置） ═══
CREATE TABLE IF NOT EXISTS menu_pages (
    id          SERIAL PRIMARY KEY,
    code        VARCHAR(128) UNIQUE NOT NULL,
    label       VARCHAR(64)  NOT NULL,
    href        VARCHAR(255) NOT NULL,
    icon        VARCHAR(64)  DEFAULT 'FileText',
    sort_order  INT          DEFAULT 0,
    group_name  VARCHAR(64)  NOT NULL,
    is_required BOOLEAN      DEFAULT FALSE
);

-- ═══ 角色-页面关联表 ═══
CREATE TABLE IF NOT EXISTS role_menu_pages (
    role_code  VARCHAR(64)  NOT NULL REFERENCES auth_roles(code) ON DELETE CASCADE,
    page_code  VARCHAR(128) NOT NULL REFERENCES menu_pages(code) ON DELETE CASCADE,
    PRIMARY KEY (role_code, page_code)
);

-- ═══ 插入 69 个页面（从 menu.ts 同步） ═══
INSERT INTO menu_pages (code, label, href, icon, sort_order, group_name, is_required) VALUES
-- 主页
('page:dashboard',       '仪表盘',     '/dashboard',            'LayoutDashboard',  1,  '主页',     TRUE),
-- 客户管理
('page:customers',       '客户档案',   '/customers',            'Building2',       10, '客户管理', FALSE),
('page:intent-customers','意向客户',   '/intent-customers',     'UserPlus',        11, '客户管理', FALSE),
('page:analytics-features','客户分析', '/analytics/features',  'BarChart3',       12, '客户管理', FALSE),
('page:analytics-cluster','聚类分析',  '/analytics/cluster',    'BarChart3',       13, '客户管理', FALSE),
('page:analytics-load',  '客户负荷',   '/analytics/load',       'Gauge',           14, '客户管理', FALSE),
('page:analytics-profit','客户利润',   '/analytics/profit',     'PieChart',        15, '客户管理', FALSE),
('page:agents',          '代理商管理', '/agents',               'Handshake',       16, '客户管理', FALSE),
('page:bonds',           '保函管理',   '/bonds',                'FileKey',         17, '客户管理', FALSE),
-- 零售管理
('page:retail-contracts','零售管理',   '/retail/contracts',     'FileText',        20, '零售管理', FALSE),
('page:retail-progress', '合同进度',   '/retail/contract-progress','CalendarCheck', 21, '零售管理', FALSE),
('page:green-power',     '绿电交易',   '/trade/green-power',    'Battery',         22, '零售管理', FALSE),
-- 文档中心
('page:documents',       '文档解析',   '/documents',            'FileScan',        30, '文档中心', FALSE),
-- 运营管理
('page:trade-rules',     '交易规则',   '/trade-rules',          'Gavel',           40, '运营管理', FALSE),
('page:solar-monitor',   '光伏监控',   '/solar/monitor',        'Sun',             42, '运营管理', FALSE),
('page:solar-settlement','光伏结算',   '/solar/settlement',     'Sun',             43, '运营管理', FALSE),
('page:market-data',     '市场行情',   '/market-data',          'CandlestickChart',44, '运营管理', FALSE),
-- 负荷与预测
('page:load-forecast',   '负荷管理',   '/load/forecast',        'Activity',        50, '负荷与预测', FALSE),
('page:load-total',      '系统总负荷', '/load/total',           'TrendingDown',    51, '负荷与预测', FALSE),
('page:load-diagnosis',  '负荷诊断',   '/load/diagnosis',       'Stethoscope',     52, '负荷与预测', FALSE),
('page:load-characteristics','负荷特性','/load/characteristics','SearchCheck',     53, '负荷与预测', FALSE),
('page:forecast-accuracy','预测管理',  '/forecast/accuracy',    'Target',          54, '负荷与预测', FALSE),
('page:forecast-base-data','基础数据', '/forecast/base-data',   'Target',          55, '负荷与预测', FALSE),
('page:weather',         '气象数据',   '/weather',              'Cloud',           56, '负荷与预测', FALSE),
-- 价格管理
('page:price-spot',      '价格管理',   '/price/spot',           'TrendingUp',      60, '价格管理', FALSE),
('page:price-long-term', '中长期预测', '/price/long-term',      'LineChart',       61, '价格管理', FALSE),
('page:price-trend',     '现货趋势',   '/price/trend',          'LineChart',       62, '价格管理', FALSE),
('page:price-tou',       'TOU 规则',   '/price/tou',            'Hourglass',       63, '价格管理', FALSE),
('page:price-grid-agency','代理购电价','/price/grid-agency',    'Zap',             64, '价格管理', FALSE),
-- 交易管理
('page:trade-da-review', '日前复盘',   '/trade/da-review',      'CalendarCheck',   70, '交易管理', FALSE),
('page:trade-da-simulation','日前模拟','/trade/da-simulation',  'FlaskConical',    71, '交易管理', FALSE),
('page:trade-monthly-review','月度复盘','/trade/monthly-review','CalendarRange',   72, '交易管理', FALSE),
('page:trade-match-quotes','撮合报价', '/trade/match-quotes',   'ArrowLeftRight',  73, '交易管理', FALSE),
('page:trade-rolling',   '滚动交易',   '/trade/rolling',        'ArrowDownUp',     74, '交易管理', FALSE),
('page:trade-bidding',   '竞价策略',   '/trade/bidding',        'Gavel',           75, '交易管理', FALSE),
('page:trade-strategies','交易策略',   '/trade/strategies',     'TrendingUp',      76, '交易管理', FALSE),
('page:vpp',             '虚拟电厂',   '/vpp',                  'Factory',         77, '交易管理', FALSE),
-- 结算管理
('page:settlement-daily','结算管理',   '/settlement/daily',     'Receipt',         80, '结算管理', FALSE),
('page:settlement-monthly','月度结算', '/settlement/monthly',   'Receipt',         81, '结算管理', FALSE),
('page:settlement-pre',  '预结算明细', '/settlement/pre',       'FileSpreadsheet', 82, '结算管理', FALSE),
('page:settlement-deviation','偏差管理','/settlement/deviation','TrendingDown',    83, '结算管理', FALSE),
('page:settlement-manual-data','手工数据','/settlement/manual-data','ClipboardEdit', 84, '结算管理', FALSE),
-- 储能与调频
('page:freq-clearing',   '调频管理',   '/freq/clearing',        'Radio',           90, '储能与调频', FALSE),
('page:storage-operation','储能管理',  '/storage/operation',    'Battery',         91, '储能与调频', FALSE),
('page:storage-declaration','储能申报','/storage/declaration',  'PackagePlus',     92, '储能与调频', FALSE),
-- 算法服务
('page:algo-carbon',     '碳排放分析', '/algo/carbon',          'Leaf',            100,'算法服务', FALSE),
('page:algo-pv-forecast','光伏预测(算法)','/algo/pv-forecast',  'Sun',             101,'算法服务', FALSE),
('page:algo-grid',       '电网潮流计算','/algo/grid',            'Cpu',             102,'算法服务', FALSE),
('page:algo-market',     '市场出清',   '/algo/market',          'DollarSign',      103,'算法服务', FALSE),
('page:algo-storage',    '储能优化',   '/algo/storage',         'BatteryCharging', 104,'算法服务', FALSE),
('page:algo-load-forecast','负荷预测', '/algo/load-forecast',   'Activity',        105,'算法服务', FALSE),
('page:algo-price-forecast','电价预测','/algo/price-forecast',  'TrendingUp',      106,'算法服务', FALSE),
('page:algo-grid-opf',   '电网优化(OPF)','/algo/grid-opf',      'GitBranch',       107,'算法服务', FALSE),
-- 系统设置
('page:users',           '用户管理',   '/users',                'Users',           110,'系统设置', FALSE),
('page:roles',           '角色管理',   '/roles',                'Settings',        111,'系统设置', FALSE),
('page:system-orgs',     '组织管理',   '/system/orgs',          'Building2',       112,'系统设置', FALSE),
('page:system-custom-fields','自定义字段','/system/custom-fields','Sliders',        113,'系统设置', FALSE),
('page:system-tags',     '标签管理',   '/system/tags',          'Tag',             114,'系统设置', FALSE),
('page:system-jobs',     '调度任务',   '/system/jobs',          'Clock',           115,'系统设置', FALSE),
('page:system-rpa',      'RPA 监控',   '/system/rpa',           'Workflow',        116,'系统设置', FALSE),
('page:system-audit',    '操作审计',   '/system/audit',         'ScrollText',      117,'系统设置', FALSE),
('page:approvals',       '审批中心',   '/approvals',            'CheckSquare',     118,'系统设置', FALSE),
('page:attachments',     '附件管理',   '/attachments',          'Paperclip',       119,'系统设置', FALSE),
('page:system-security', '安全大屏',   '/system/security',      'Shield',          120,'系统设置', FALSE),
('page:system-settings', '系统配置',   '/system/settings',      'Sliders',         121,'系统设置', FALSE),
('page:carbon',          '碳交易',     '/carbon',               'Leaf',            122,'系统设置', FALSE),
('page:settings',        '个人设置',   '/settings',             'UserCog',         123,'系统设置', TRUE)
ON CONFLICT (code) DO NOTHING;

-- ═══ 初始化角色-页面关联（根据现有菜单权限映射） ═══
-- super_admin 看到所有页面
INSERT INTO role_menu_pages (role_code, page_code)
SELECT 'super_admin', code FROM menu_pages
ON CONFLICT DO NOTHING;

-- admin: 看到除用户管理/角色管理/组织管理之外的所有页面
INSERT INTO role_menu_pages (role_code, page_code)
SELECT 'admin', code FROM menu_pages
WHERE code NOT IN ('page:users', 'page:roles', 'page:system-orgs', 'page:system-rpa', 'page:system-audit')
ON CONFLICT DO NOTHING;

-- analyst: 业务相关页面
INSERT INTO role_menu_pages (role_code, page_code)
SELECT 'analyst', code FROM menu_pages
WHERE code IN (
    'page:dashboard', 'page:settings',
    'page:customers', 'page:intent-customers',
    'page:analytics-features', 'page:analytics-cluster', 'page:analytics-load', 'page:analytics-profit',
    'page:documents',
    'page:load-forecast', 'page:load-total', 'page:load-diagnosis', 'page:load-characteristics',
    'page:forecast-accuracy', 'page:forecast-base-data', 'page:weather',
    'page:price-spot', 'page:price-long-term', 'page:price-trend', 'page:price-tou', 'page:price-grid-agency',
    'page:trade-da-review', 'page:trade-da-simulation', 'page:trade-monthly-review',
    'page:trade-match-quotes', 'page:trade-rolling', 'page:trade-bidding', 'page:trade-strategies', 'page:vpp',
    'page:settlement-daily', 'page:settlement-monthly', 'page:settlement-pre',
    'page:settlement-deviation', 'page:settlement-manual-data',
    'page:freq-clearing', 'page:storage-operation', 'page:storage-declaration',
    'page:algo-carbon', 'page:algo-pv-forecast', 'page:algo-grid', 'page:algo-market',
    'page:algo-storage', 'page:algo-load-forecast', 'page:algo-price-forecast', 'page:algo-grid-opf',
    'page:retail-contracts', 'page:retail-progress', 'page:green-power',
    'page:trade-rules', 'page:solar-monitor', 'page:solar-settlement', 'page:market-data',
    'page:agents', 'page:bonds', 'page:carbon'
)
ON CONFLICT DO NOTHING;

-- viewer: 只看基础页面
INSERT INTO role_menu_pages (role_code, page_code)
SELECT 'viewer', code FROM menu_pages
WHERE code IN (
    'page:dashboard', 'page:settings',
    'page:customers', 'page:intent-customers',
    'page:documents',
    'page:load-forecast', 'page:load-total', 'page:weather',
    'page:price-spot', 'page:price-trend',
    'page:settlement-daily', 'page:settlement-monthly',
    'page:approvals', 'page:attachments'
)
ON CONFLICT DO NOTHING;
