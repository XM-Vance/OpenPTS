// tradeDaySchedule 实现 cron.Schedule，在原 cron 触发时间基础上，
// 跳过周末与法定节假日（holidays 表 kind='public'），保留调休补班日（kind='makeup_workday'）。
// 用于 trade_day_only=true 的调度任务（P0-D4）。
package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/robfig/cron/v3"
)

// tradeDaySchedule 包装原 cron schedule 与节假日数据源。
type tradeDaySchedule struct {
	inner    cron.Schedule           // 原 cron 解析出的 schedule
	holidays *db.ForecastBaseRepository
	mu       sync.Mutex
	// 缓存：年份 → 该年非交易日集合（周末不在缓存内，由 weekday 判断；缓存仅含法定假日 date 字符串）
	cacheYear int
	holidaysSet map[string]bool
}

func newTradeDaySchedule(expr string, holidays *db.ForecastBaseRepository) (*tradeDaySchedule, error) {
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour |
		cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	sched, err := parser.Parse(expr)
	if err != nil {
		// 兜底：尝试用调度器默认的 6 段秒制解析器
		p2 := cron.NewParser(cron.Second | cron.Minute | cron.Hour |
			cron.Dom | cron.Month | cron.Dow)
		sched, err = p2.Parse(expr)
		if err != nil {
			return nil, err
		}
	}
	return &tradeDaySchedule{inner: sched, holidays: holidays}, nil
}

// isHoliday 判断某日是否为非交易日：周末 或 法定假日（且非调休补班）。
func (t *tradeDaySchedule) isHoliday(d time.Time) bool {
	wd := d.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		// 调休补班日即使是周末也应执行：需查库确认 kind='makeup_workday'
		return !t.isMakeupWorkday(d)
	}
	// 工作日：仅法定假日才跳过
	return t.isPublicHoliday(d)
}

// loadHolidaysForYear 懒加载某年的假日集合（含 public 跳过、makeup_workday 补班）。
func (t *tradeDaySchedule) loadHolidaysForYear(year int) {
	if t.holidays == nil {
		t.holidaysSet = map[string]bool{}
		t.cacheYear = year
		return
	}
	if t.holidaysSet != nil && t.cacheYear == year {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.holidaysSet != nil && t.cacheYear == year {
		return
	}
	list, err := t.holidays.ListHolidays(context.Background(), year)
	if err != nil {
		// 查库失败时退化为「不跳过任何工作日」（保守：宁可执行）
		t.holidaysSet = map[string]bool{}
		t.cacheYear = year
		return
	}
	set := make(map[string]bool, len(list))
	for _, h := range list {
		if h.Kind == "public" {
			set[h.HolidayDate.Format("2006-01-02")] = true
		}
		// makeup_workday 不入「跳过」集合（补班日要执行）
	}
	t.holidaysSet = set
	t.cacheYear = year
}

func (t *tradeDaySchedule) isPublicHoliday(d time.Time) bool {
	t.loadHolidaysForYear(d.Year())
	return t.holidaysSet[d.Format("2006-01-02")]
}

// isMakeupWorkday 判断周末调休补班（kind='makeup_workday'）。
// 由于 holidaysSet 只缓存 public，补班日需单独标记——这里用一个反向缓存。
// 简化：重新查 ListHolidays 已在上层缓存，补班日不在 holidaysSet 中，
// 但周末默认跳过，故需独立判断。为避免复杂化，补班日识别用第二缓存。
var (
	makeupMu       sync.Mutex
	makeupCache    map[string]bool
	makeupCacheYr  int
)

func (t *tradeDaySchedule) isMakeupWorkday(d time.Time) bool {
	if t.holidays == nil {
		return false
	}
	year := d.Year()
	makeupMu.Lock()
	if makeupCache == nil || makeupCacheYr != year {
		list, err := t.holidays.ListHolidays(context.Background(), year)
		if err != nil {
			makeupCache = map[string]bool{}
		} else {
			m := make(map[string]bool, len(list))
			for _, h := range list {
				if h.Kind == "makeup_workday" {
					m[h.HolidayDate.Format("2006-01-02")] = true
				}
			}
			makeupCache = m
		}
		makeupCacheYr = year
	}
	res := makeupCache[d.Format("2006-01-02")]
	makeupMu.Unlock()
	return res
}

// Next 返回 t 之后下一个交易日触发时间。
// 逻辑：先用 inner.Next 算出候选；若候选落在非交易日，则按天向后顺延，
// 每顺延一天用「次日同一时刻」重算，直到落到交易日。
func (t *tradeDaySchedule) Next(after time.Time) time.Time {
	candidate := t.inner.Next(after)
	for {
		if !t.isHoliday(candidate) {
			return candidate
		}
		// 落在非交易日：跳到次日凌晨，让 inner.Next 重算到下一个 cron 命中点
		nextDay := time.Date(candidate.Year(), candidate.Month(), candidate.Day(),
			0, 0, 0, 0, candidate.Location()).Add(24 * time.Hour)
		candidate = t.inner.Next(nextDay.Add(-1 * time.Second))
		if candidate.IsZero() {
			return candidate
		}
	}
}
