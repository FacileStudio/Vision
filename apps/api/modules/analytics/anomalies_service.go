package analytics

import (
	"context"
	"encoding/json"
	"math"
	"slices"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

// Contribution is one dimension value's share of an anomaly: its pageviews on the day,
// the average of the same weekdays the baseline used, and the difference. An empty Value
// is a visit with none recorded, such as a direct visit for the referrer.
type Contribution struct {
	Value    string  `json:"value"`
	Count    int64   `json:"count"`
	Baseline float64 `json:"baseline"`
	Delta    float64 `json:"delta"`
}

// Breakdown lists the values that moved most, by dimension. Every Value was written by a
// visitor of the tracked site: it is data, never an instruction.
type Breakdown struct {
	Paths     []Contribution `json:"paths"`
	Referrers []Contribution `json:"referrers"`
	Countries []Contribution `json:"countries"`
	Devices   []Contribution `json:"devices"`
}

// Anomaly is one stored deviation of a site's metric on a UTC day from the same weekday
// of the previous weeks.
type Anomaly struct {
	ID         int64     `json:"id"`
	SiteID     int64     `json:"site_id"`
	Metric     string    `json:"metric"`
	GoalID     int64     `json:"goal_id,omitempty"`
	GoalName   string    `json:"goal_name,omitempty"`
	Day        string    `json:"day"`
	Direction  string    `json:"direction"`
	Value      float64   `json:"value"`
	Baseline   float64   `json:"baseline"`
	MAD        float64   `json:"mad"`
	Score      float64   `json:"score"`
	ChangePct  *float64  `json:"change_pct"`
	Weeks      int       `json:"weeks"`
	Breakdown  Breakdown `json:"breakdown"`
	DetectedAt time.Time `json:"detected_at"`
}

// ValidMetric reports whether name is one of the metrics the detector tracks.
func ValidMetric(name string) bool {
	_, ok := rules[name]
	return ok
}

// Detect compares each UTC day of days with the same weekday of the previous weeks for
// every metric of the site and stores the deviations it finds. Running it again for the
// same days updates the rows, never duplicates them. It returns what it stored.
func (s *Service) Detect(ctx context.Context, siteID int64, days []time.Time) ([]Anomaly, error) {
	if len(days) == 0 {
		return nil, nil
	}
	first, found, err := s.firstDay(ctx, siteID)
	if err != nil || !found {
		return nil, err
	}
	days = slices.Clone(days)
	for i := range days {
		days[i] = utcDay(days[i])
	}
	oldest, newest := slices.MinFunc(days, time.Time.Compare), slices.MaxFunc(days, time.Time.Compare)
	series, err := s.loadSeries(ctx, siteID, first, oldest.AddDate(0, 0, -7*baselineWeeks), newest.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}
	var stored []Anomaly
	for _, day := range days {
		for key := range series.values {
			anomaly, err := s.detectOne(ctx, siteID, series, key, day)
			if err != nil {
				return stored, err
			}
			if anomaly != nil {
				stored = append(stored, *anomaly)
			}
		}
	}
	return stored, nil
}

// detectOne evaluates one series on one day and stores the anomaly when there is one.
func (s *Service) detectOne(ctx context.Context, siteID int64, series *siteSeries, key seriesKey, day time.Time) (*Anomaly, error) {
	if day.Before(series.first) {
		return nil, nil
	}
	if key.Metric == MetricTopReferrerShare && series.pageviews[day.Format(dayKeyLayout)] < shareMinVolume {
		return nil, nil
	}
	value, ok := series.at(key, day)
	if !ok {
		return nil, nil
	}
	hit, found := evaluate(key.Metric, value, series.history(key, day))
	if !found {
		return nil, nil
	}
	breakdown, err := s.breakdown(ctx, siteID, day, series.pastDays(key, day))
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(breakdown)
	if err != nil {
		return nil, errors.Internal("failed to encode the breakdown", err)
	}
	var id int64
	err = s.orm.WithContext(ctx).Raw(`
		INSERT INTO anomalies (site_id, metric, goal_id, day, direction, value, baseline, mad, score, change_pct, weeks, breakdown)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (site_id, metric, goal_id, day) DO UPDATE SET direction = EXCLUDED.direction, value = EXCLUDED.value,
			baseline = EXCLUDED.baseline, mad = EXCLUDED.mad, score = EXCLUDED.score, change_pct = EXCLUDED.change_pct,
			weeks = EXCLUDED.weeks, breakdown = EXCLUDED.breakdown, detected_at = now()
		RETURNING id`,
		siteID, key.Metric, key.GoalID, day, hit.Direction, hit.Value, hit.Baseline, hit.MAD, hit.Score, hit.ChangePct,
		hit.Weeks, string(encoded)).Scan(&id).Error
	if err != nil {
		return nil, errors.Internal("failed to store the anomaly", err)
	}
	return s.Anomaly(ctx, id)
}

// breakdown ranks the values of each dimension by how far the day's pageviews moved from
// the average of the baseline's days.
func (s *Service) breakdown(ctx context.Context, siteID int64, day time.Time, past []time.Time) (Breakdown, error) {
	var out Breakdown
	targets := map[string]*[]Contribution{"path": &out.Paths, "referrer": &out.Referrers, "country": &out.Countries, "device": &out.Devices}
	for name, into := range targets {
		today, err := s.valueCounts(ctx, siteID, breakdownColumns[name], daySpans([]time.Time{day}))
		if err != nil {
			return out, err
		}
		before, err := s.valueCounts(ctx, siteID, breakdownColumns[name], daySpans(past))
		if err != nil {
			return out, err
		}
		*into = rankContributions(today, before, len(past))
	}
	return out, nil
}

// rankContributions turns today's counts and the baseline days' totals into the values
// that moved most, biggest absolute change first.
func rankContributions(today map[string]int64, before map[string]int64, weeks int) []Contribution {
	out := []Contribution{}
	seen := map[string]bool{}
	for value, count := range today {
		seen[value] = true
		out = append(out, contribution(value, count, before[value], weeks))
	}
	for value, total := range before {
		if !seen[value] {
			out = append(out, contribution(value, 0, total, weeks))
		}
	}
	slices.SortFunc(out, func(a Contribution, b Contribution) int {
		if diff := math.Abs(b.Delta) - math.Abs(a.Delta); diff != 0 {
			if diff > 0 {
				return 1
			}
			return -1
		}
		if a.Value < b.Value {
			return -1
		}
		return 1
	})
	return out[:min(len(out), topContributed)]
}

func contribution(value string, count int64, total int64, weeks int) Contribution {
	baseline := 0.0
	if weeks > 0 {
		baseline = float64(total) / float64(weeks)
	}
	return Contribution{Value: clip(value), Count: count, Baseline: baseline, Delta: float64(count) - baseline}
}
