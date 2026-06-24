import { apiClient } from './client';

/** 单个 widget 配置 */
export interface WidgetConfig {
  id: string;
  visible: boolean;
}

/** 用户仪表盘布局配置 */
export interface DashboardConfig {
  widgets: WidgetConfig[];
}

/** 所有可用 widget 定义 */
export interface WidgetDef {
  id: string;
  label: string;
  description: string;
  defaultVisible: boolean;
  /** 分组：kpi(数据卡片) | spark(趋势) | panel(主面板) */
  group: 'kpi' | 'spark' | 'panel';
}

/** 可用 widget 注册表（顺序即默认展示顺序） */
export const WIDGET_REGISTRY: WidgetDef[] = [
  {
    id: 'kpi_cards',
    label: 'KPI 数据卡片',
    description: '客户数、合同、套餐、告警、电站、结算额',
    defaultVisible: true,
    group: 'kpi',
  },
  {
    id: 'spark_settlement',
    label: '近14日结算额趋势',
    description: '迷你折线图：每日结算收入',
    defaultVisible: true,
    group: 'spark',
  },
  {
    id: 'spark_freq',
    label: '近14日调频收益趋势',
    description: '迷你折线图：每日调频收入',
    defaultVisible: true,
    group: 'spark',
  },
  {
    id: 'market_overview',
    label: '市场概览',
    description: '日前/实时/中长期交易量与均价',
    defaultVisible: true,
    group: 'panel',
  },
  {
    id: 'timeline',
    label: '待办事项 & 预警通知',
    description: '待处理事项和系统告警时间线',
    defaultVisible: true,
    group: 'panel',
  },
  {
    id: 'settlement_panel',
    label: '结算详情',
    description: '近期结算明细和统计',
    defaultVisible: true,
    group: 'panel',
  },
  {
    id: 'trade_review',
    label: '交易复盘',
    description: '交易策略执行回顾与分析',
    defaultVisible: true,
    group: 'panel',
  },
  {
    id: 'customer_overview',
    label: '客户概览',
    description: '客户分布、合同到期、用电量排行',
    defaultVisible: true,
    group: 'panel',
  },
  {
    id: 'market_price',
    label: '市场价格',
    description: '实时节点电价和价格走势',
    defaultVisible: true,
    group: 'panel',
  },
  {
    id: 'alerts',
    label: '告警面板',
    description: '系统告警和偏差预警',
    defaultVisible: true,
    group: 'panel',
  },
];

/** 默认配置（所有 widget 按注册顺序可见） */
export function getDefaultConfig(): DashboardConfig {
  return {
    widgets: WIDGET_REGISTRY.map((w) => ({
      id: w.id,
      visible: w.defaultVisible,
    })),
  };
}

/** 将后端返回的配置合并为完整列表（补全缺失的 widget） */
export function normalizeConfig(saved: DashboardConfig | null): DashboardConfig {
  if (!saved?.widgets) return getDefaultConfig();
  const savedMap = new Map(saved.widgets.map((w) => [w.id, w.visible]));
  return {
    widgets: WIDGET_REGISTRY.map((w) => ({
      id: w.id,
      visible: savedMap.has(w.id) ? (savedMap.get(w.id) as boolean) : w.defaultVisible,
    })),
  };
}

const LS_KEY = 'ptis_dashboard_config';

/** 加载配置：API 优先，localStorage 回退 */
export async function loadConfig(): Promise<DashboardConfig> {
  try {
    const { data } = await apiClient.get('/api/v1/dashboard/config');
    return normalizeConfig(data);
  } catch {
    // localStorage 回退
    try {
      const raw = localStorage.getItem(LS_KEY);
      if (raw) return normalizeConfig(JSON.parse(raw));
    } catch {}
    return getDefaultConfig();
  }
}

/** 保存配置：API + localStorage 双写 */
export async function saveConfig(config: DashboardConfig): Promise<void> {
  // localStorage 即时写入（用户无需等待网络）
  try {
    localStorage.setItem(LS_KEY, JSON.stringify(config));
  } catch {}
  // 后端异步写入
  try {
    await apiClient.put('/api/v1/dashboard/config', config);
  } catch {
    // 后端失败不影响前端体验（已存 localStorage）
  }
}
