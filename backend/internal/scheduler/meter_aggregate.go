// 表计负荷聚合作业（WP0.3）：每日 23:40 聚合近 2 天 raw_meter_data →
// user_load_data + unified_load_curve（跨 2 天是为兜住 23 点后补传的表计数据）。
package scheduler

import (
	"context"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

func AggregateMeterLoad(ctx context.Context, pool *db.Pool) error {
	repo := db.NewMeterRepository(pool)
	end := db.StartOfDay(time.Now())
	start := end.AddDate(0, 0, -1)
	written, skipped, err := repo.AggregateMeters(ctx, start, end)
	if err != nil {
		return err
	}
	log.Info().Str("start", start.Format("2006-01-02")).Str("end", end.Format("2006-01-02")).
		Int("written", written).Int("skipped", skipped).
		Msg("表计负荷聚合完成（raw_meter_data → user_load_data + unified_load_curve）")
	return nil
}
