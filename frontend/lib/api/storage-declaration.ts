import { apiClient } from './client';

export type StationStatus = '启用' | '停用';
export type StrategyStatus = '启用' | '停用';
export type DeclareStatus = '已申报' | '未申报';

export interface StorageStation {
  station_id: string;
  station_name: string;
  control_unit_name: string;
  node_name: string;
  voltage_level: string;
  rated_power_mw: number;
  rated_capacity_mwh: number;
  is_hybrid: boolean;
  fm_power_mw: number;
  fm_capacity_mwh: number;
  charge_efficiency: number;
  discharge_efficiency: number;
  efficiency?: number;
  discharge_depth: number;
  fm_k_value: number;
  default_mileage_beta: number;
  default_soc: number;
  degradation_cost_per_mwh: number;
  status: StationStatus;
  created_at?: string;
  updated_at?: string;
}

export interface StationPayload {
  station_name: string;
  control_unit_name: string;
  node_name: string;
  voltage_level: string;
  rated_power_mw: number;
  rated_capacity_mwh: number;
  is_hybrid: boolean;
  fm_power_mw: number;
  fm_capacity_mwh: number;
  charge_efficiency: number;
  discharge_efficiency: number;
  discharge_depth: number;
  fm_k_value: number;
  default_mileage_beta: number;
  default_soc: number;
  degradation_cost_per_mwh: number;
  status: StationStatus;
}

export interface StrategyParam {
  param_key: string;
  param_name: string;
  param_value: string;
  unit: string;
  description: string;
}

export interface StorageStrategy {
  strategy_id: string;
  station_id: string;
  strategy_name: string;
  strategy_type: string;
  strategy_status: StrategyStatus;
  fm_price_threshold: number;
  description: string;
  strategy_params: StrategyParam[];
  next_day_declare_status?: DeclareStatus;
  created_at?: string;
  updated_at?: string;
}

export interface StrategyPayload {
  station_id: string;
  strategy_name: string;
  strategy_type: string;
  strategy_status: StrategyStatus;
  fm_price_threshold: number;
  description: string;
  strategy_params: StrategyParam[];
}

export interface EnergySlot {
  time_point: string;
  power_mw: number;
}

export interface FmSlot {
  period_start: string;
  period_end: string;
  output_base_mw: number;
  mileage_price: number;
}

export interface GenerateResult {
  energy_declaration: EnergySlot[];
  fm_declaration: FmSlot[];
  soc_trajectory: number[];
  spot_price_forecast: number[];
  arbitrage_executed: boolean;
  charge_hours: number[];
  discharge_hours: number[];
  p_charge_mw: number;
  p_discharge_mw: number;
  max_soc?: number;
  target_date: string;
  violations?: string[];
  discharge_depth?: number;
  fm_price_forecast_24?: number[];
  fm_price_basis?: Record<string, unknown>;
  forecast_revenue?: RevenueResult;
  generation_message?: string;
}

export interface StorageDeclaration {
  declaration_id: string;
  declaration_key: string;
  station_id: string;
  station_name: string;
  strategy_id: string;
  strategy_name: string;
  strategy_type: string;
  target_date: string;
  status: 'created' | 'submitted' | 'settled';
  declare_status: DeclareStatus;
  energy_declaration_96: number[];
  fm_declaration_24: number[];
  fm_output_base_24?: number[];
  energy_slots_96: EnergySlot[];
  fm_slots_24: FmSlot[];
  soc_trajectory_96: number[];
  spot_price_forecast_96: number[];
  arbitrage_executed: boolean;
  charge_hours: number[];
  discharge_hours: number[];
  p_charge_mw: number;
  p_discharge_mw: number;
  violations: string[];
  generation_message?: string;
  forecast_revenue?: RevenueResult;
  total_charge_mwh: number;
  total_discharge_mwh: number;
  fm_hours: number;
  station_snapshot: Record<string, unknown>;
  strategy_snapshot: Record<string, unknown>;
  params_snapshot: Record<string, unknown>;
  review_status?: '未复盘' | '已复盘';
  review_simulated_at?: string;
  review_node_realtime_price_96?: number[];
  review_fm_clearing_price_24?: number[];
  review_energy_slots_96?: ReviewEnergySlot[];
  review_fm_slots_24?: ReviewFmSlot[];
  review_metrics?: ReviewMetrics;
  review_readiness?: ReviewReadiness;
  forecast_date?: string;
  generated_at?: string;
  submitted_at?: string;
  settled_at?: string;
  created_at?: string;
  updated_at?: string;
}

export interface SaveDeclarationPayload {
  station_id: string;
  strategy_id: string;
  target_date: string;
  declare_status: DeclareStatus;
  energy_declaration: EnergySlot[];
  fm_declaration: FmSlot[];
  soc_trajectory: number[];
  spot_price_forecast: number[];
  params_snapshot: Record<string, unknown>;
  result_meta?: Record<string, unknown>;
}

export interface RevenueResult {
  energy_revenue: number;
  fm_revenue: number;
  operational_revenue: number;
  degradation_cost: number;
  net_revenue: number;
  total_charge_mwh: number;
  total_discharge_mwh: number;
  net_consumption_mwh: number;
  net_surcharge_per_mwh: number;
  peak_valley_spread?: number;
  grid_surcharge_detail: {
    transmission_distribution_price: number;
    government_fund: number;
    network_loss_price: number;
    system_op_cost_discount: number;
    peak_valley_bonus: number;
  };
  slot_pnl: number[];
  params: {
    beta: number;
    kp: number;
    clearing_price: number;
    degradation_cost_per_mwh: number;
    winning_capacity_mw: number;
  };
}

export interface CalculateRevenuePayload {
  station_id: string;
  target_date: string;
  energy_declaration: EnergySlot[];
  fm_declaration: FmSlot[];
  prices_96?: number[];
  beta?: number;
  kp?: number;
  clearing_price?: number;
  degradation_cost_per_mwh?: number;
}

export interface ReviewEnergySlot {
  time_point: string;
  power_mw: number;
  energy_mwh: number;
  charge_mwh: number;
  discharge_mwh: number;
  node_realtime_price: number;
  charge_fee: number;
  discharge_revenue: number;
  soc: number;
}

export interface ReviewFmSlot {
  hour: number;
  period_start: string;
  period_end: string;
  output_base_mw: number;
  mileage_price: number;
  clearing_compare_price?: number;
  intraday_clearing_price: number;
  is_winning: boolean;
  mileage: number;
  revenue: number;
}

export interface ReviewMetrics {
  total_revenue: number;
  energy_revenue: number;
  peak_valley_spread?: number;
  charge_mwh: number;
  charge_fee: number;
  discharge_mwh: number;
  discharge_revenue: number;
  loss_mwh: number;
  loss_fee: number;
  fm_revenue: number;
  winning_hours: number;
  fm_mileage: number;
  avg_clearing_price: number;
  fm_revenue_per_winning_hour?: number;
  charge_weighted_price?: number;
  loss_price?: number;
  fm_k_value?: number;
  default_mileage_beta?: number;
}

export interface StorageProfitSummary {
  start_date: string;
  end_date: string;
  reviewed_days: number;
  natural_days: number;
  total_revenue: number;
  energy_revenue: number;
  fm_revenue: number;
  avg_daily_revenue: number;
  charge_mwh: number;
  discharge_mwh: number;
  loss_mwh: number;
  winning_hours: number;
  fm_mileage: number;
  avg_clearing_price: number;
  fm_revenue_per_winning_hour: number;
}

export interface StorageProfitDailyRow {
  date: string;
  total_revenue: number;
  energy_revenue: number;
  fm_revenue: number;
  cumulative_revenue: number;
  cumulative_energy_revenue: number;
  cumulative_fm_revenue: number;
  charge_mwh: number;
  discharge_mwh: number;
  loss_mwh: number;
  winning_hours: number;
  fm_mileage: number;
  avg_clearing_price: number;
  fm_period_revenue: number;
}

export interface StorageProfitAnalysis {
  summary: StorageProfitSummary;
  rows: StorageProfitDailyRow[];
}

export interface ReviewReadiness {
  can_review: boolean;
  message: string;
  node_realtime_points: number;
  fm_intraday_hours: number;
}

export interface ReviewSimulatePayload {
  station_id: string;
  strategy_id: string;
  target_date: string;
}

// 工具方法
export const slotTimeLabel = (slotIndex: number): string => {
  const totalMinutes = (slotIndex + 1) * 15;
  if (totalMinutes >= 24 * 60) return '24:00';
  const hour = Math.floor(totalMinutes / 60);
  const minute = totalMinutes % 60;
  return `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`;
};

export interface StationOperationPoint {
  time: string;
  sced_mw: number;
  unit_power_mw: number;
  meter_power_mw: number;
  soc_percent: number;
}

export const STRATEGY_TYPE_OPTIONS: { value: string; label: string }[] = [
  { value: 'simple_peak_valley', label: '简单峰谷套利' },
  { value: 'threshold_arbitrage', label: '价差阈值套利' },
  { value: 'fm_priority', label: '调频优先' },
  { value: 'hybrid_opt', label: '混合优化' },
];

export const VOLTAGE_LEVEL_OPTIONS: string[] = ['10', '20', '35', '110', '220'];

export const simulateSoc = (
  slots: EnergySlot[],
  capacity: number,
  chargeEfficiency: number,
  dischargeEfficiency: number,
  socInit: number,
): number[] => {
  const dt = 0.25;
  const chargeEff = Math.max(chargeEfficiency, 1e-6);
  const dischargeEff = Math.max(dischargeEfficiency, 1e-6);
  const cap = Math.max(capacity, 1e-6);
  let soc = socInit;
  const result: number[] = [];
  for (const s of slots) {
    const power = s.power_mw || 0;
    if (power > 0) {
      soc -= (power * dt) / (cap * dischargeEff);
    } else if (power < 0) {
      soc += (Math.abs(power) * chargeEff * dt) / cap;
    }
    soc = Math.max(0, Math.min(1, soc));
    result.push(Number(soc.toFixed(6)));
  }
  return result;
};

// ─── 以下自 v1clone-e.ts 迁入（2026-06，储能申报（简版列表）） ───
// ─── E6 储能申报 ───
export interface StorageDeclarationItem {
  id: string;
  station_id: string;
  station_name: string;
  declared_date: string;
  charge_mw: number[];
  discharge_mw: number[];
  expected_revenue: number;
  strategy_note?: string | null;
  created_at: string;
}
export async function listStorageDeclarations(
  params: { station_id?: string; days?: number } = {},
): Promise<{ items: StorageDeclarationItem[] }> {
  const { data } = await apiClient.get('/api/v1/storage/declarations', { params });
  return data;
}
export async function genStorageDeclDemo(): Promise<{ declarations: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/storage/declarations/demo-data');
  return data;
}
