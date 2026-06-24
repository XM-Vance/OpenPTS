// 合同文档 → 零售合同的跨页预填载体（sessionStorage，一次性消费）。
// 文档解析提取到的所有字段自动映射到合同表单，跳合同页由人工核对修改后保存。

const KEY = 'ptis.contract_prefill';

export interface ContractPrefill {
  fromDoc?: string;
  // 自动匹配客户
  customerCandidates?: string[];
  // 合同核心字段
  energyMwh?: string;
  greenRatio?: string;
  startMonth?: string; // YYYY-MM
  endMonth?: string;
  // 扩展字段（从文档提取，预填到表单供人工核对）
  price?: string;          // 电价（元/MWh 或 元/kWh）
  totalAmount?: string;    // 合同总金额（万元）
  contractNo?: string;     // 合同编号
  signDate?: string;       // 签订日期
  partyA?: string;         // 甲方
  partyB?: string;         // 乙方
  packageNameHint?: string; // 套餐名称提示
  voltageLevel?: string;   // 电压等级
  settlementMethod?: string; // 结算方式
  remarks?: string;        // 备注
  // 所有提取字段原文（供参考展示）
  allFields?: { key: string; label: string; value: string }[];
}

export function setContractPrefill(p: ContractPrefill): void {
  if (typeof window === 'undefined') return;
  try {
    window.sessionStorage.setItem(KEY, JSON.stringify(p));
  } catch {
    /* 忽略隐私模式等写入失败 */
  }
}

export function takeContractPrefill(): ContractPrefill | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = window.sessionStorage.getItem(KEY);
    if (!raw) return null;
    window.sessionStorage.removeItem(KEY);
    return JSON.parse(raw) as ContractPrefill;
  } catch {
    return null;
  }
}
