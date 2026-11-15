package analytics

import (
	"context"
	"encoding/json"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

const (
	listDefault = 50
	listMax     = 200
)

// AnomalyQuery narrows ListAnomalies. Zero values mean no bound, any metric and the
// default limit.
type AnomalyQuery struct {
	From   time.Time
	To     time.Time
	Metric string
	Limit  int
}

type anomalyRow struct {
	ID         int64
	SiteID     int64
	Metric     string
	GoalID     int64
	GoalName   *string
	Day        time.Time
	Direction  string
	Value      float64
	Baseline   float64
	MAD        float64
	Score      float64
	ChangePct  *float64
	Weeks      int
	Breakdown  string
	DetectedAt time.Time
}

const anomalySelect = `SELECT a.id, a.site_id, a.metric, a.goal_id, g.name AS goal_name, a.day, a.direction, a.value,
	a.baseline, a.mad, a.score, a.change_pct, a.weeks, a.breakdown, a.detected_at
	FROM anomalies a LEFT JOIN goals g ON g.id = a.goal_id AND a.goal_id <> 0`

// Anomaly reads one stored anomaly by id.
func (s *Service) Anomaly(ctx context.Context, id int64) (*Anomaly, error) {
	var rows []anomalyRow
	if err := s.orm.WithContext(ctx).Raw(anomalySelect+` WHERE a.id = ?`, id).Scan(&rows).Error; err != nil {
		return nil, errors.Internal("failed to read the anomaly", err)
	}
	if len(rows) == 0 {
		return nil, errors.NotFound("anomaly not found")
	}
	return rows[0].anomaly()
}

// ListAnomalies returns a site's anomalies, newest day first, then the biggest score.
func (s *Service) ListAnomalies(ctx context.Context, siteID int64, query AnomalyQuery) ([]Anomaly, error) {
	if query.Metric != "" && !ValidMetric(query.Metric) {
		return nil, errors.Invalid("metric must be visitors, pageviews, goal_conversions or top_referrer_share")
	}
	limit := query.Limit
	if limit <= 0 {
		limit = listDefault
	}
	limit = min(limit, listMax)
	var rows []anomalyRow
	err := s.orm.WithContext(ctx).Raw(anomalySelect+`
		WHERE a.site_id = @site AND (@metric = '' OR a.metric = @metric)
		  AND (CAST(@since AS timestamptz) IS NULL OR a.day >= @since) AND (CAST(@until AS timestamptz) IS NULL OR a.day <= @until)
		ORDER BY a.day DESC, abs(a.score) DESC, a.id DESC LIMIT @limit`,
		map[string]any{"site": siteID, "metric": query.Metric, "since": nullTime(query.From), "until": nullTime(query.To), "limit": limit}).
		Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to list anomalies", err)
	}
	out := make([]Anomaly, 0, len(rows))
	for _, row := range rows {
		anomaly, err := row.anomaly()
		if err != nil {
			return nil, err
		}
		out = append(out, *anomaly)
	}
	return out, nil
}

// nullTime is nil for the zero time, so the query binds NULL.
func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func (r anomalyRow) anomaly() (*Anomaly, error) {
	out := &Anomaly{
		ID: r.ID, SiteID: r.SiteID, Metric: r.Metric, GoalID: r.GoalID, Day: r.Day.UTC().Format(dayKeyLayout),
		Direction: r.Direction, Value: r.Value, Baseline: r.Baseline, MAD: r.MAD, Score: r.Score,
		ChangePct: r.ChangePct, Weeks: r.Weeks, DetectedAt: r.DetectedAt,
	}
	if r.GoalName != nil {
		out.GoalName = *r.GoalName
	}
	if err := json.Unmarshal([]byte(r.Breakdown), &out.Breakdown); err != nil {
		return nil, errors.Internal("failed to decode the breakdown", err)
	}
	return out, nil
}
