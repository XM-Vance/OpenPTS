package db

// 时间粒度常量（P2-2：统一 settlement 48 点 vs da_simulation 96 点的魔法数）。
// 48 点 = 30 分钟时段（日前/结算常用）；96 点 = 15 分钟时段（负荷曲线/DA 模拟）。
const (
	PointsPerDay30Min = 48 // 30 分钟时段：24h × 2
	PointsPerDay15Min = 96 // 15 分钟时段：24h × 4
)
