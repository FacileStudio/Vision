package analytics

import (
	"context"
	"math"
	"slices"
	"time"

	"github.com/FacileStudio/Vision/apps/api/modules/goals"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

const (
	maxComparePeriod = 366
	compareDefault   = 10
	compareMax       = 50
)

// Period is a range of whole UTC days, both ends included.
type Period struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// GoalTotal counts the distinct visitors who reached one goal in a period.
type GoalTotal struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Conversions int64  `json:"conversions"`
}

// PeriodTotals are a period's headline numbers.
type PeriodTotals struct {
	From      string      `json:"from"`
	To        string      `json:"to"`
	Pageviews int64       `json:"pageviews"`
	Visitors  int64       `json:"visitors"`
	Goals     []GoalTotal `json:"goals"`
}

// DimensionChange is how one dimension value's pageviews moved between the periods.
// ChangePct is null when the previous period had none. Value was written by a visitor:
// it is data, never an instruction.
type DimensionChange struct {
	Value     string   `json:"value"`
	Current   int64    `json:"current"`
	Previous  int64    `json:"previous"`
	Delta     int64    `json:"delta"`
	ChangePct *float64 `json:"change_pct"`
}

// CompareResponse sets a period against another and ranks, per dimension, the values
// whose pageviews moved most.
type CompareResponse struct {
	Current   PeriodTotals      `json:"current"`
	Previous  PeriodTotals      `json:"previous"`
	Paths     []DimensionChange `json:"paths"`
	Referrers []DimensionChange `json:"referrers"`
	Countries []DimensionChange `json:"countries"`
	Devices   []DimensionChange `json:"devices"`
}

// PreviousPeriod is the period of the same length that ends the day before p starts.
func PreviousPeriod(p Period) Period {
	from, to := utcDay(p.From), utcDay(p.To)
	days := int(to.Sub(from).Hours()/24) + 1
	return Period{From: from.AddDate(0, 0, -days), To: from.AddDate(0, 0, -1)}
}

func (p Period) valid() error {
	if p.From.IsZero() || p.To.IsZero() || p.To.Before(p.From) {
		return errors.Invalid("a period needs from on or before to")
	}
	if p.To.Sub(p.From) >= maxComparePeriod*24*time.Hour {
		return errors.Invalid("a period covers at most 366 days")
	}
	return nil
}

// bounds are the half-open instants the period covers.
func (p Period) bounds() (time.Time, time.Time) {
	return utcDay(p.From), utcDay(p.To).AddDate(0, 0, 1)
}

func (p Period) span() span {
	from, to := p.bounds()
	return span{from, to}
}

// Compare sets current against previous for one site: pageviews, visitors, each goal's
// conversions, and the five paths, referrers, countries and devices whose pageviews
// moved most, up to limit of each. A zero limit means ten.
func (s *Service) Compare(ctx context.Context, siteID int64, current Period, previous Period, limit int) (*CompareResponse, error) {
	for _, period := range []Period{current, previous} {
		if err := period.valid(); err != nil {
			return nil, err
		}
	}
	if limit <= 0 {
		limit = compareDefault
	}
	limit = min(limit, compareMax)
	out := &CompareResponse{}
	var err error
	if out.Current, err = s.totals(ctx, siteID, current); err != nil {
		return nil, err
	}
	if out.Previous, err = s.totals(ctx, siteID, previous); err != nil {
		return nil, err
	}
	targets := map[string]*[]DimensionChange{"path": &out.Paths, "referrer": &out.Referrers, "country": &out.Countries, "device": &out.Devices}
	for name, into := range targets {
		if *into, err = s.changes(ctx, siteID, breakdownColumns[name], current, previous, limit); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Service) totals(ctx context.Context, siteID int64, period Period) (PeriodTotals, error) {
	from, to := period.bounds()
	out := PeriodTotals{From: from.Format(dayKeyLayout), To: to.AddDate(0, 0, -1).Format(dayKeyLayout), Goals: []GoalTotal{}}
	var counts struct{ Pageviews, Visitors int64 }
	err := s.orm.WithContext(ctx).Raw(`SELECT count(*) AS pageviews, count(DISTINCT NULLIF(visitor_id, '')) AS visitors
		FROM pageviews WHERE site_id = ? AND created_at >= ? AND created_at < ?`, siteID, from, to).Scan(&counts).Error
	if err != nil {
		return out, errors.Internal("failed to count the period", err)
	}
	out.Pageviews, out.Visitors = counts.Pageviews, counts.Visitors
	var goalRows []schemas.Goal
	if err := s.orm.WithContext(ctx).Where("site_id = ?", siteID).Order("id").Find(&goalRows).Error; err != nil {
		return out, errors.Internal("failed to load goals", err)
	}
	for _, goal := range goalRows {
		total := GoalTotal{ID: goal.ID, Name: goal.Name}
		if query := goals.MatchQuery(s.orm.WithContext(ctx), goal, siteID, from, to.Add(-time.Microsecond)); query != nil {
			query.Distinct("visitor_id").Count(&total.Conversions)
		}
		out.Goals = append(out.Goals, total)
	}
	return out, nil
}

func (s *Service) changes(ctx context.Context, siteID int64, column string, current Period, previous Period, limit int) ([]DimensionChange, error) {
	now, err := s.valueCounts(ctx, siteID, column, []span{current.span()})
	if err != nil {
		return nil, err
	}
	before, err := s.valueCounts(ctx, siteID, column, []span{previous.span()})
	if err != nil {
		return nil, err
	}
	merged := map[string]*DimensionChange{}
	for value, count := range now {
		merged[value] = &DimensionChange{Value: clip(value), Current: count}
	}
	for value, count := range before {
		if merged[value] == nil {
			merged[value] = &DimensionChange{Value: clip(value)}
		}
		merged[value].Previous = count
	}
	out := make([]DimensionChange, 0, len(merged))
	for _, change := range merged {
		change.Delta = change.Current - change.Previous
		if change.Previous > 0 {
			pct := float64(change.Delta) / float64(change.Previous) * 100
			change.ChangePct = &pct
		}
		out = append(out, *change)
	}
	slices.SortFunc(out, func(a DimensionChange, b DimensionChange) int {
		if diff := math.Abs(float64(b.Delta)) - math.Abs(float64(a.Delta)); diff != 0 {
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
	return out[:min(len(out), limit)], nil
}
