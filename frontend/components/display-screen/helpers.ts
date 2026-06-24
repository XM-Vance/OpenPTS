export function fmtNum(n: number): string {
  return Math.round(n).toLocaleString('zh-CN');
}

export function fmtMoney(n: number): string {
  if (n >= 10_000_000) return (n / 10_000_000).toFixed(2) + ' 千万';
  if (n >= 10_000) return (n / 10_000).toFixed(2) + ' 万';
  return fmtNum(n);
}

export function fmtEnergy(n: number): string {
  if (n >= 10_000) return (n / 10_000).toFixed(2) + ' 万 MWh';
  return n.toFixed(1) + ' MWh';
}
