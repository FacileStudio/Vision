package analytics

import (
	"context"
	"sort"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

const maxSummaryWindow = 366 * 24 * time.Hour

type siteTotals struct {
	SiteID             int64
	Pageviews          int64
	UniqueVisitors     int64
	PrevPageviews      int64
	PrevUniqueVisitors int64
}

type siteBucket struct {
	SiteID int64
	Bucket time.Time
	Count  int64
}

// Summary returns headline numbers for every site in the workspaces userID belongs to,
// narrowed to one workspace when workspaceID is positive. A workspace the user is not a
// member of yields no sites. Three grouped queries cover every site, whatever the count.
func (s *Service) Summary(ctx context.Context, userID int64, workspaceID int64, from time.Time, to time.Time) (*SummaryResponse, error) {
	if !to.After(from) || to.Sub(from) > maxSummaryWindow {
		return nil, errors.Invalid("from must be before to, at most 366 days apart")
	}
	out := &SummaryResponse{From: from, To: to, Granularity: summaryUnit(from, to), Sites: []SiteSummary{}}
	if err := s.readableSites(ctx, userID, workspaceID, &out.Sites); err != nil {
		return nil, err
	}
	if len(out.Sites) == 0 {
		return out, nil
	}
	ids := make([]int64, len(out.Sites))
	for i, site := range out.Sites {
		ids[i] = site.SiteID
	}
	totals, err := s.summaryTotals(ctx, ids, from, to)
	if err != nil {
		return nil, err
	}
	buckets, err := s.summaryBuckets(ctx, ids, from, to, out.Granularity)
	if err != nil {
		return nil, err
	}
	fillSummary(out, totals, buckets)
	return out, nil
}

// readableSites loads the sites of every workspace userID is a member of, or of the one
// workspace when workspaceID is positive and userID belongs to it.
func (s *Service) readableSites(ctx context.Context, userID int64, workspaceID int64, into *[]SiteSummary) error {
	err := s.orm.WithContext(ctx).Raw(`
		SELECT s.id AS site_id, s.name, s.domain, s.workspace_id, w.name AS workspace_name
		FROM sites s
		JOIN workspaces w ON w.id = s.workspace_id
		WHERE EXISTS (SELECT 1 FROM workspace_members m WHERE m.workspace_id = s.workspace_id AND m.user_id = ?)
		  AND (? = 0 OR s.workspace_id = ?)
		ORDER BY s.id`, userID, workspaceID, workspaceID).Scan(into).Error
	if err != nil {
		return errors.Internal("failed to list sites", err)
	}
	return nil
}

// summaryTotals counts pageviews and unique visitors per site in the window and in the
// window of the same length right before it, in one grouped query.
func (s *Service) summaryTotals(ctx context.Context, ids []int64, from time.Time, to time.Time) ([]siteTotals, error) {
	var totals []siteTotals
	err := s.orm.WithContext(ctx).Raw(`
		SELECT site_id,
			count(*) FILTER (WHERE created_at >= @from) AS pageviews,
			count(DISTINCT NULLIF(visitor_id, '')) FILTER (WHERE created_at >= @from) AS unique_visitors,
			count(*) FILTER (WHERE created_at < @from) AS prev_pageviews,
			count(DISTINCT NULLIF(visitor_id, '')) FILTER (WHERE created_at < @from) AS prev_unique_visitors
		FROM pageviews
		WHERE site_id IN @ids AND created_at >= @prev AND created_at <= @to
		GROUP BY site_id`,
		map[string]any{"ids": ids, "from": from, "to": to, "prev": from.Add(-to.Sub(from))}).Scan(&totals).Error
	if err != nil {
		return nil, errors.Internal("failed to count pageviews", err)
	}
	return totals, nil
}

// summaryBuckets counts pageviews per site and UTC hour or day, in one grouped query.
func (s *Service) summaryBuckets(ctx context.Context, ids []int64, from time.Time, to time.Time, unit string) ([]siteBucket, error) {
	var buckets []siteBucket
	err := s.orm.WithContext(ctx).Raw(`
		SELECT site_id, date_trunc(@unit, created_at AT TIME ZONE 'UTC') AS bucket, count(*) AS count
		FROM pageviews
		WHERE site_id IN @ids AND created_at >= @from AND created_at <= @to
		GROUP BY 1, 2`,
		map[string]any{"ids": ids, "from": from, "to": to, "unit": unit}).Scan(&buckets).Error
	if err != nil {
		return nil, errors.Internal("failed to query the pageview series", err)
	}
	return buckets, nil
}

// fillSummary merges the grouped rows into the site list, zero-fills each series and
// sorts the sites.
func fillSummary(out *SummaryResponse, totals []siteTotals, buckets []siteBucket) {
	byTotals := make(map[int64]siteTotals, len(totals))
	for _, row := range totals {
		byTotals[row.SiteID] = row
	}
	counts := bucketCounts(buckets, out.Granularity)
	starts := bucketStarts(out.From, out.To, out.Granularity)
	for i := range out.Sites {
		site := &out.Sites[i]
		row := byTotals[site.SiteID]
		site.Pageviews, site.UniqueVisitors = row.Pageviews, row.UniqueVisitors
		site.PrevPageviews, site.PrevUniqueVisitors = row.PrevPageviews, row.PrevUniqueVisitors
		site.Series = make([]DayStat, len(starts))
		for j, start := range starts {
			key := bucketKey(start, out.Granularity)
			site.Series[j] = DayStat{Date: key, Count: counts[site.SiteID][key]}
		}
	}
	sort.SliceStable(out.Sites, func(i, j int) bool { return busier(out.Sites[i], out.Sites[j]) })
}

// busier orders sites by visitors, then pageviews, then name.
func busier(a SiteSummary, b SiteSummary) bool {
	if a.UniqueVisitors != b.UniqueVisitors {
		return a.UniqueVisitors > b.UniqueVisitors
	}
	if a.Pageviews != b.Pageviews {
		return a.Pageviews > b.Pageviews
	}
	return a.Name < b.Name
}
