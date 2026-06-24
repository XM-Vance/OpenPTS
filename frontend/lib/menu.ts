import {
  FileScan,
  LayoutDashboard,
  LayoutGrid,
  Monitor,
  Users,
  Settings,
  Building2,
  FileText,
  Activity,
  TrendingUp,
  Receipt,
  Radio,
  Battery,
  BarChart3,
  Clock,
  ScrollText,
  Cloud,
  Workflow,
  CalendarCheck,
  LineChart,
  UserPlus,
  Gauge,
  Stethoscope,
  SearchCheck,
  Hourglass,
  Zap,
  PackagePlus,
  PieChart,
  CalendarRange,
  ArrowLeftRight,
  ClipboardEdit,
  CheckSquare,
  Paperclip,
  UserCog,
  Shield,
  Sliders,
  Target,
  TrendingDown,
  FileSpreadsheet,
  Factory,
  ArrowDownUp,
  Gavel,
  Handshake,
  FileKey,
  Sun,
  FlaskConical,
  Leaf,
  Cpu,
  DollarSign,
  BatteryCharging,
  GitBranch,
  CandlestickChart,
  Tag,
  type LucideIcon,
} from 'lucide-react';

export interface MenuItem {
  label: string;
  href: string;
  icon: LucideIcon;
  /** 需要的权限码；缺省则总是可见 */
  permission?: string;
}

export interface MenuGroup {
  group: string;
  items: MenuItem[];
}

// 前端固定菜单。每项绑定权限码，无权限的项自动隐藏。
// 按模块关联性重组：客户管理 → 零售管理 → 文档中心 → 福建管理 → 负荷/价格 → 交易 → 结算 → 算法 → 系统
export const MENU: MenuGroup[] = [
  {
    group: '主页',
    items: [
      { label: '仪表盘', href: '/dashboard', icon: LayoutDashboard },
      { label: '展示大屏', href: '/display-screen', icon: Monitor },
    ],
  },
  {
    group: '客户管理',
    items: [
      { label: '客户档案', href: '/customers', icon: Building2, permission: 'customer_management:read' },
      { label: '意向客户', href: '/intent-customers', icon: UserPlus, permission: 'customer_management:read' },
      { label: '客户分析', href: '/analytics/features', icon: BarChart3, permission: 'analytics:read' },
      { label: '聚类分析', href: '/analytics/cluster', icon: BarChart3, permission: 'analytics:read' },
      { label: '客户负荷', href: '/analytics/load', icon: Gauge, permission: 'analytics:read' },
      { label: '客户利润', href: '/analytics/profit', icon: PieChart, permission: 'analytics:read' },
      { label: '客户电量', href: '/customer-energy', icon: Zap, permission: 'customer_management:read' },
      { label: '代理商管理', href: '/agents', icon: Handshake, permission: 'customer_management:read' },
      { label: '保函管理', href: '/bonds', icon: FileKey, permission: 'customer_management:read' },
    ],
  },
  {
    group: '零售管理',
    items: [
      { label: '零售管理', href: '/retail/contracts', icon: FileText, permission: 'retail_management:read' },
      { label: '合同进度', href: '/retail/contract-progress', icon: CalendarCheck, permission: 'retail_management:read' },
      { label: '绿电交易', href: '/trade/green-power', icon: Battery, permission: 'retail_management:read' },
    ],
  },
  {
    group: '文档中心',
    items: [
      { label: '文档解析', href: '/documents', icon: FileScan, permission: 'document_management:read' },
      { label: '政策文件', href: '/policies', icon: ScrollText, permission: 'document_management:read' },
    ],
  },
  {
    group: '运营管理',
    items: [
      { label: '交易规则', href: '/trade-rules', icon: Gavel, permission: 'system:read' },
      { label: '光伏监控', href: '/solar/monitor', icon: Sun, permission: 'storage:read' },
      { label: '光伏结算', href: '/solar/settlement', icon: Sun, permission: 'storage:read' },
      { label: '市场行情', href: '/market-data', icon: CandlestickChart, permission: 'price_management:read' },
    ],
  },
  {
    group: '负荷管理',
    items: [
      { label: '系统总负荷', href: '/load/total', icon: TrendingDown, permission: 'load_management:read' },
      { label: '气象数据', href: '/weather', icon: Cloud, permission: 'load_management:read' },
    ],
  },
  {
    group: '价格管理',
    items: [
      { label: '价格管理', href: '/price/spot', icon: TrendingUp, permission: 'price_management:read' },
      { label: 'TOU 规则', href: '/price/tou', icon: Hourglass, permission: 'price_management:read' },
      { label: '代理购电价', href: '/price/grid-agency', icon: Zap, permission: 'price_management:read' },
    ],
  },
  {
    group: '结算管理',
    items: [
      { label: '结算管理', href: '/settlement/daily', icon: Receipt, permission: 'settlement_management:read' },
      { label: '月度结算', href: '/settlement/monthly', icon: Receipt, permission: 'settlement_management:read' },
      { label: '预结算明细', href: '/settlement/pre', icon: FileSpreadsheet, permission: 'settlement_management:read' },
      { label: '偏差管理', href: '/settlement/deviation', icon: TrendingDown, permission: 'settlement_management:read' },
      { label: '手工数据', href: '/settlement/manual-data', icon: ClipboardEdit, permission: 'settlement_management:read' },
    ],
  },
  {
    group: '储能与调频',
    items: [
      { label: '调频管理', href: '/freq/clearing', icon: Radio, permission: 'freq_regulation:read' },
      { label: '储能管理', href: '/storage/operation', icon: Battery, permission: 'storage:read' },
      { label: '储能申报', href: '/storage/declaration', icon: PackagePlus, permission: 'storage:read' },
    ],
  },
  {
    group: '系统设置',
    items: [
      { label: '用户管理', href: '/users', icon: Users, permission: 'user_management:read' },
      { label: '角色管理', href: '/roles', icon: Settings, permission: 'user_management:read' },
      { label: '组织管理', href: '/system/orgs', icon: Building2, permission: 'user_management:read' },
      { label: '自定义字段', href: '/system/custom-fields', icon: Sliders, permission: 'system:write' },
      { label: '标签管理', href: '/system/tags', icon: Tag, permission: 'system:read' },
      { label: '调度任务', href: '/system/jobs', icon: Clock, permission: 'task_scheduler:read' },
      { label: 'RPA 监控', href: '/system/rpa', icon: Workflow, permission: 'task_scheduler:read' },
      { label: '操作审计', href: '/system/audit', icon: ScrollText, permission: 'system:read' },
      { label: '审批中心', href: '/approvals', icon: CheckSquare },
      { label: '附件管理', href: '/attachments', icon: Paperclip },
      { label: '安全大屏', href: '/system/security', icon: Shield, permission: 'system:read' },
      { label: '菜单管理', href: '/system/menu-management', icon: LayoutGrid, permission: 'system:write' },
      { label: '系统配置', href: '/system/settings', icon: Sliders, permission: 'system:read' },
      { label: '碳交易', href: '/carbon', icon: Leaf, permission: 'price_management:read' },
      { label: '个人设置', href: '/settings', icon: UserCog },
    ],
  },
];
