package analytics

import (
	"context"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/FacileStudio/Vision/apps/api/modules/goals"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

const (
	dayKeyLayout   = "2006-01-02"
	topContributed = 5
	valueLimit     = 500
	valueRunes     = 200
)

var breakdownColumns = map[string]string{
	"path": "path", "referrer": "referrer", "country": "country", "device": "device",
}

// seriesKey names one tracked series: a metric, and the goal for goal conversions.
type seriesKey struct {
	Metric string
	GoalID int64
}

// siteSeries holds every daily series of one site, keyed by UTC day as 2006-01-02. A day
// missing from a series had no events, which is zero, except for the referrer share where
// a day with no pageviews has no share at all.
type siteSeries struct {
	first     time.Time
	values    map[seriesKey]map[string]float64
	pageviews map[string]float64
}

type dayRow struct {
	Day   string
	Count float64
}

type trafficRow struct {
	Day       string
	Pageviews float64
	Visitors  float64
}

// utcDay truncates t to its UTC day.
func utcDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// firstDay is the UTC day of the site's first pageview, or false when it has none.
func (s *Service) firstDay(ctx context.Context, siteID int64) (time.Time, bool, error) {
	var first *time.Time
	err := s.orm.WithContext(ctx).Table("pageviews").Select("min(created_at)").Where("site_id = ?", siteID).Scan(&first).Error
	if err != nil {
		return time.Time{}, false, errors.Internal("failed to find the first pageview", err)
	}
	if first == nil {
		return time.Time{}, false, nil
	}
	return utcDay(*first), true, nil
}

// loadSeries reads the daily series of every tracked metric for [from, to).
func (s *Service) loadSeries(ctx context.Context, siteID int64, first time.Time, from time.Time, to time.Time) (*siteSeries, error) {
	out := &siteSeries{first: first, values: map[seriesKey]map[string]float64{}, pageviews: map[string]float64{}}
	var traffic []trafficRow
	err := s.orm.WithContext(ctx).Raw(`
		SELECT to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS day, count(*) AS pageviews,
			count(DISTINCT NULLIF(visitor_id, '')) AS visitors
		FROM pageviews WHERE site_id = ? AND created_at >= ? AND created_at < ? GROUP BY 1`, siteID, from, to).Scan(&traffic).Error
	if err != nil {
		return nil, errors.Internal("failed to read the traffic series", err)
	}
	visitors, pageviews := map[string]float64{}, map[string]float64{}
	for _, row := range traffic {
		visitors[row.Day], pageviews[row.Day] = row.Visitors, row.Pageviews
	}
	out.pageviews = pageviews
	out.values[seriesKey{Metric: MetricVisitors}] = visitors
	out.values[seriesKey{Metric: MetricPageviews}] = pageviews
	share, err := s.referrerShare(ctx, siteID, from, to, pageviews)
	if err != nil {
		return nil, err
	}
	out.values[seriesKey{Metric: MetricTopReferrerShare}] = share
	return out, s.loadGoalSeries(ctx, siteID, from, to, out)
}

// referrerShare is, per day, the pageviews of the busiest named referrer over every
// pageview of that day. Direct visits count in the total and never as a referrer.
func (s *Service) referrerShare(ctx context.Context, siteID int64, from time.Time, to time.Time, totals map[string]float64) (map[string]float64, error) {
	var rows []dayRow
	err := s.orm.WithContext(ctx).Raw(`
		SELECT day, count FROM (
			SELECT day, n AS count, row_number() OVER (PARTITION BY day ORDER BY n DESC, referrer) AS rank FROM (
				SELECT to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS day, referrer, count(*) AS n
				FROM pageviews WHERE site_id = ? AND created_at >= ? AND created_at < ? AND COALESCE(referrer, '') <> ''
				GROUP BY 1, 2) per_referrer) ranked
		WHERE rank = 1`, siteID, from, to).Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to read the referrer series", err)
	}
	out := map[string]float64{}
	for _, row := range rows {
		if totals[row.Day] > 0 {
			out[row.Day] = row.Count / totals[row.Day]
		}
	}
	return out, nil
}

// loadGoalSeries adds one distinct-visitor series per goal of the site.
func (s *Service) loadGoalSeries(ctx context.Context, siteID int64, from time.Time, to time.Time, out *siteSeries) error {
	var goalRows []schemas.Goal
	if err := s.orm.WithContext(ctx).Where("site_id = ?", siteID).Find(&goalRows).Error; err != nil {
		return errors.Internal("failed to load goals", err)
	}
	for _, goal := range goalRows {
		query := goals.MatchQuery(s.orm.WithContext(ctx), goal, siteID, from, to.Add(-time.Microsecond))
		if query == nil {
			continue
		}
		var rows []dayRow
		err := query.Select("to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS day, count(DISTINCT visitor_id) AS count").
			Group("day").Scan(&rows).Error
		if err != nil {
			return errors.Internal("failed to read the goal series", err)
		}
		series := map[string]float64{}
		for _, row := range rows {
			series[row.Day] = row.Count
		}
		out.values[seriesKey{Metric: MetricGoalConversions, GoalID: goal.ID}] = series
	}
	return nil
}

// history lists the value of the series on the same weekday of up to baselineWeeks
// weeks before day, newest first. Weeks before the site's first pageview are left out,
// and so are days a share is undefined on.
func (d *siteSeries) history(key seriesKey, day time.Time) []float64 {
	var out []float64
	for week := 1; week <= baselineWeeks; week++ {
		past := day.AddDate(0, 0, -7*week)
		if past.Before(d.first) {
			break
		}
		if value, ok := d.at(key, past); ok {
			out = append(out, value)
		}
	}
	return out
}

// pastDays are the weeks of history the baseline used, for the breakdown.
func (d *siteSeries) pastDays(key seriesKey, day time.Time) []time.Time {
	var out []time.Time
	for week := 1; week <= baselineWeeks; week++ {
		past := day.AddDate(0, 0, -7*week)
		if past.Before(d.first) {
			break
		}
		if _, ok := d.at(key, past); ok {
			out = append(out, past)
		}
	}
	return out
}

// at reads one day of a series. Counts default to zero; the share needs pageviews.
func (d *siteSeries) at(key seriesKey, day time.Time) (float64, bool) {
	name := day.Format(dayKeyLayout)
	if key.Metric == MetricTopReferrerShare {
		if d.pageviews[name] == 0 {
			return 0, false
		}
	}
	return d.values[key][name], true
}

// span is a half-open range of instants.
type span struct {
	from, to time.Time
}

// daySpans are the whole UTC days as spans.
func daySpans(days []time.Time) []span {
	out := make([]span, len(days))
	for i, day := range days {
		out[i] = span{day, day.AddDate(0, 0, 1)}
	}
	return out
}

// valueCounts counts each value of one dimension column over the spans, as pageviews,
// most frequent first. The column is always one of breakdownColumns.
func (s *Service) valueCounts(ctx context.Context, siteID int64, column string, spans []span) (map[string]int64, error) {
	if len(spans) == 0 {
		return map[string]int64{}, nil
	}
	clauses := make([]string, len(spans))
	args := []any{siteID}
	for i, one := range spans {
		clauses[i] = "(created_at >= ? AND created_at < ?)"
		args = append(args, one.from, one.to)
	}
	var rows []struct {
		Value string
		Count int64
	}
	err := s.orm.WithContext(ctx).Raw(`SELECT COALESCE(`+column+`, '') AS value, count(*) AS count FROM pageviews
		WHERE site_id = ? AND (`+strings.Join(clauses, " OR ")+`) GROUP BY 1 ORDER BY count DESC LIMIT `+strconv.Itoa(valueLimit), args...).Scan(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to read the breakdown", err)
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.Value] = row.Count
	}
	return out, nil
}

// clip keeps a visitor-written value short and valid UTF-8 before it is stored.
func clip(value string) string {
	value = strings.ToValidUTF8(value, "")
	if utf8.RuneCountInString(value) <= valueRunes {
		return value
	}
	return string([]rune(value)[:valueRunes])
}
