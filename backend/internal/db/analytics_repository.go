// 客户分析仓储：异动告警 + 客户特征。
package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type CustomerAlert struct {
	ID           uuid.UUID       `json:"id"`
	CustomerID   uuid.UUID       `json:"customer_id"`
	CustomerName string          `json:"customer_name"`
	AlertDate    time.Time       `json:"alert_date"`
	AlertType    string          `json:"alert_type"`
	Severity     string          `json:"severity"`
	Confidence   *float64        `json:"confidence,omitempty"`
	Reason       *string         `json:"reason,omitempty"`
	Metrics      json.RawMessage `json:"metrics,omitempty"`
	RuleID       *string         `json:"rule_id,omitempty"`
	Acknowledged bool            `json:"acknowledged"`
	AckBy        *uuid.UUID      `json:"acknowledged_by,omitempty"`
	AckAt        *time.Time      `json:"acknowledged_at,omitempty"`
	Note         *string         `json:"note,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type CustomerCharacteristic struct {
	ID              uuid.UUID       `json:"id"`
	CustomerID      uuid.UUID       `json:"customer_id"`
	CustomerName    string          `json:"customer_name"`
	DataDate        time.Time       `json:"data_date"`
	LongTerm        json.RawMessage `json:"long_term"`
	ShortTerm       json.RawMessage `json:"short_term"`
	Tags            []string        `json:"tags"`
	RegularityScore *float64        `json:"regularity_score,omitempty"`
	QualityRating   *string         `json:"quality_rating,omitempty"`
}

type AlertStats struct {
	Total        int `json:"total"`
	Pending      int `json:"pending"`
	Acknowledged int `json:"acknowledged"`
	Critical     int `json:"critical"`
}

var ErrAlertNotFound = errors.New("告警不存在")

type AnalyticsRepository struct {
	pool *Pool
}

func NewAnalyticsRepository(pool *Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

// ─── 告警 ───

const alertSelect = `SELECT a.id, a.customer_id, c.user_name, a.alert_date, a.alert_type, a.severity,
		a.confidence, a.reason, a.metrics, a.rule_id, a.acknowledged,
		a.acknowledged_by, a.acknowledged_at, a.note, a.created_at
		FROM customer_anomaly_alerts a
		JOIN customers c ON c.id = a.customer_id`

func (r *AnalyticsRepository) ListAlerts(
	ctx context.Context, limit int, includeAcked bool,
) ([]*CustomerAlert, error) {
	where := " WHERE 1=1"
	if !includeAcked {
		where += " AND a.acknowledged = FALSE"
	}
	q := alertSelect + where + " ORDER BY a.alert_date DESC, a.created_at DESC LIMIT $1"

	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*CustomerAlert, 0, limit)
	for rows.Next() {
		var a CustomerAlert
		var metrics []byte
		if err := rows.Scan(&a.ID, &a.CustomerID, &a.CustomerName,
			&a.AlertDate, &a.AlertType, &a.Severity,
			&a.Confidence, &a.Reason, &metrics, &a.RuleID,
			&a.Acknowledged, &a.AckBy, &a.AckAt, &a.Note, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Metrics = json.RawMessage(metrics)
		list = append(list, &a)
	}
	return list, rows.Err()
}

func (r *AnalyticsRepository) AckAlert(ctx context.Context, id, userID uuid.UUID) error {
	const q = `UPDATE customer_anomaly_alerts
		SET acknowledged = TRUE, acknowledged_by = $2, acknowledged_at = NOW()
		WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAlertNotFound
	}
	return nil
}

func (r *AnalyticsRepository) InsertAlert(
	ctx context.Context, custID uuid.UUID, d time.Time,
	alertType, severity, ruleID, reason string, confidence float64,
) error {
	const q = `INSERT INTO customer_anomaly_alerts
		(customer_id, alert_date, alert_type, severity, confidence, reason, rule_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(ctx, q, custID, d, alertType, severity, confidence, reason, ruleID)
	return err
}

// GetAlertStats 一次 SQL 取四项统计。
func (r *AnalyticsRepository) GetAlertStats(ctx context.Context) (*AlertStats, error) {
	const q = `SELECT
		COUNT(*)                                       AS total,
		COUNT(*) FILTER (WHERE NOT acknowledged)       AS pending,
		COUNT(*) FILTER (WHERE acknowledged)           AS acked,
		COUNT(*) FILTER (WHERE severity = 'critical') AS critical
		FROM customer_anomaly_alerts`
	var s AlertStats
	if err := r.pool.QueryRow(ctx, q).Scan(
		&s.Total, &s.Pending, &s.Acknowledged, &s.Critical,
	); err != nil {
		return nil, err
	}
	return &s, nil
}

// ─── 客户特征 ───

// ListLatestCharacteristics 每个客户取最近一条 characteristic。
func (r *AnalyticsRepository) ListLatestCharacteristics(
	ctx context.Context, limit int,
) ([]*CustomerCharacteristic, error) {
	const q = `
		SELECT DISTINCT ON (cc.customer_id)
			cc.id, cc.customer_id, cust.user_name, cc.data_date,
			cc.long_term, cc.short_term, cc.tags,
			cc.regularity_score, cc.quality_rating
		FROM customer_characteristics cc
		JOIN customers cust ON cust.id = cc.customer_id
		ORDER BY cc.customer_id, cc.data_date DESC
		LIMIT $1`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*CustomerCharacteristic, 0, limit)
	for rows.Next() {
		var x CustomerCharacteristic
		var lt, st []byte
		if err := rows.Scan(&x.ID, &x.CustomerID, &x.CustomerName, &x.DataDate,
			&lt, &st, &x.Tags, &x.RegularityScore, &x.QualityRating); err != nil {
			return nil, err
		}
		x.LongTerm = json.RawMessage(lt)
		x.ShortTerm = json.RawMessage(st)
		list = append(list, &x)
	}
	return list, rows.Err()
}

func (r *AnalyticsRepository) UpsertCharacteristic(
	ctx context.Context, custID uuid.UUID, d time.Time,
	longTerm, shortTerm json.RawMessage, tags []string,
	regularity float64, quality string,
) error {
	const q = `INSERT INTO customer_characteristics
		(customer_id, data_date, long_term, short_term, tags, regularity_score, quality_rating)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (customer_id, data_date) DO UPDATE SET
			long_term        = EXCLUDED.long_term,
			short_term       = EXCLUDED.short_term,
			tags             = EXCLUDED.tags,
			regularity_score = EXCLUDED.regularity_score,
			quality_rating   = EXCLUDED.quality_rating`
	_, err := r.pool.Exec(ctx, q, custID, d, longTerm, shortTerm, tags, regularity, quality)
	return err
}
