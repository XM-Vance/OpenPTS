'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { MENU } from '@/lib/menu';
import { menuApi, type MenuPage } from '@/lib/api/menu';
import { usePermission } from '@/lib/auth/use-permission';
import { useI18n } from '@/lib/i18n/context';
import { cn } from '@/lib/utils';
import * as LucideIcons from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

// 菜单 label / group → i18n key 映射；找不到则回退到原值。
const LABEL_TO_KEY: Record<string, string> = {
  // 分组
  '主页': 'nav.home',
  '系统': 'nav.system',
  '业务': 'nav.business',
  // 主页
  '仪表盘': 'nav.dashboard',
  '展示大屏': 'nav.displayScreen',
  // 系统
  '用户管理': 'nav.users',
  '角色管理': 'nav.roles',
  '调度任务': 'nav.jobs',
  'RPA 监控': 'nav.rpa',
  '操作审计': 'nav.audit',
  '审批中心': 'nav.approvals',
  '附件管理': 'nav.attachments',
  '安全大屏': 'nav.security',
  '系统配置': 'nav.settings',
  '个人设置': 'nav.personalSettings',
  // 业务
  '客户档案': 'nav.customers',
  '意向客户': 'nav.intentCustomers',
  '零售管理': 'nav.retail',
  '合同进度': 'nav.contractProgress',
  '负荷管理': 'nav.load',
  '系统总负荷': 'nav.systemLoad',
  '预测管理': 'nav.forecastMgmt',
  '基础数据': 'nav.baseData',
  '负荷诊断': 'nav.loadDiagnosis',
  '负荷特性': 'nav.loadCharacteristics',
  '价格管理': 'nav.price',
  '中长期预测': 'nav.longTermForecast',
  '现货趋势': 'nav.spotTrend',
  'TOU 规则': 'nav.touRules',
  '代理购电价': 'nav.gridAgencyPrice',
  '日前复盘': 'nav.daReview',
  '日前模拟': 'nav.daSimulation',
  '月度复盘': 'nav.monthlyReview',
  '撮合报价': 'nav.matchQuotes',
  '绿电交易': 'nav.greenPower',
  '滚动交易': 'nav.rollingTrade',
  '竞价策略': 'nav.bidding',
  '交易策略': 'nav.strategies',
  '虚拟电厂': 'nav.vpp',
  '结算管理': 'nav.dailySettlement',
  '月度结算': 'nav.monthlySettlement',
  '预结算明细': 'nav.preSettlement',
  '偏差管理': 'nav.deviation',
  '手工数据': 'nav.manualData',
  '调频管理': 'nav.freqClearing',
  '储能管理': 'nav.storageOperation',
  '储能申报': 'nav.storageDeclaration',
  '客户分析': 'nav.features',
  '聚类分析': 'nav.cluster',
  '客户负荷': 'nav.customerLoad',
  '客户利润': 'nav.profit',
  '气象数据': 'nav.weather',
  '代理商管理': 'nav.agents',
  '光伏预测': 'nav.solarForecast',
  '光伏监控': 'nav.solarMonitor',
  '光伏结算': 'nav.solarSettlement',
  // 算法
  '碳排放分析': 'nav.carbonAnalysis',
  '光伏预测(算法)': 'nav.algoPVForecast',
  '电网潮流计算': 'nav.gridCalc',
  '市场出清': 'nav.marketClearing',
  '储能优化': 'nav.storageOptimize',
  '负荷预测': 'nav.loadForecast',
  '电价预测': 'nav.priceForecast',
  '电网优化(OPF)': 'nav.gridOPF',
};

// 图标名 → 组件 映射
function getIcon(name: string): LucideIcon {
  const icons = LucideIcons as unknown as Record<string, LucideIcon>;
  return icons[name] || LucideIcons.FileText;
}

/**
 * 导航主体（不含 <aside> 外壳）：供桌面常驻侧边栏与移动 Sheet 抽屉复用。
 * 调用方负责包外壳与滚动容器；本组件只渲染标题 + 分组链接列表。
 */
export function SidebarNav({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname();
  const { has } = usePermission();
  const { t } = useI18n();

  const { data: dynamicPages } = useQuery({
    queryKey: ['menu-visible'],
    queryFn: menuApi.getVisiblePages,
    staleTime: 5 * 60 * 1000,
  });

  const tr = (label: string): string => {
    const key = LABEL_TO_KEY[label];
    return key ? t(key) : label;
  };

  const renderItem = (href: string, label: string, Icon: LucideIcon, key: string) => {
    const active = pathname === href || pathname.startsWith(href + '/');
    return (
      <li key={key}>
        <Link
          href={href}
          onClick={onNavigate}
          className={cn(
            'flex items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors',
            active ? 'bg-primary text-primary-foreground' : 'text-foreground hover:bg-accent',
          )}
        >
          <Icon className="h-4 w-4" />
          {tr(label)}
        </Link>
      </li>
    );
  };

  return (
    <>
      <div className="flex h-14 items-center border-b px-4">
        <span className="text-sm font-bold">{t('app.title')}</span>
      </div>
      <nav className="flex-1 space-y-4 overflow-y-auto p-3">
        {/* 动态菜单优先 */}
        {dynamicPages && dynamicPages.length > 0
          ? (() => {
              const groups = new Map<string, MenuPage[]>();
              for (const page of dynamicPages) {
                if (!groups.has(page.group_name)) groups.set(page.group_name, []);
                groups.get(page.group_name)!.push(page);
              }
              return Array.from(groups.entries()).map(([groupName, pages]) => (
                <div key={groupName}>
                  <p className="mb-1 px-2 text-xs font-medium text-muted-foreground">
                    {tr(groupName)}
                  </p>
                  <ul className="space-y-0.5">
                    {pages.map((page) => {
                      const Icon = getIcon(page.icon);
                      return renderItem(page.href, page.label, Icon, page.code);
                    })}
                  </ul>
                </div>
              ));
            })()
          : // 回退：静态菜单
            MENU.map((group) => {
              const visible = group.items.filter((it) => !it.permission || has(it.permission));
              if (visible.length === 0) return null;
              return (
                <div key={group.group}>
                  <p className="mb-1 px-2 text-xs font-medium text-muted-foreground">
                    {tr(group.group)}
                  </p>
                  <ul className="space-y-0.5">
                    {visible.map((it) => renderItem(it.href, it.label, it.icon, it.href))}
                  </ul>
                </div>
              );
            })}
      </nav>
    </>
  );
}

/**
 * 桌面侧边栏：<768px 隐藏（移动端由 AppShell 用 Sheet 承载同一 SidebarNav）。
 * hidden md:flex 保证桌面端布局与原 w-60 常驻行为完全一致。
 */
export function Sidebar() {
  return (
    <aside className="hidden h-full w-60 shrink-0 flex-col border-r bg-card md:flex">
      <SidebarNav />
    </aside>
  );
}
