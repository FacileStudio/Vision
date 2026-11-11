package mcp

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Vision/apps/api/modules/analytics"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	topN        = 10
	listsLimit  = 100
	defaultSize = "day"
)

var errGranularity = stderrors.New("granularity must be hour, day, week or month")

// granularities are the time series buckets the dashboard offers.
var granularities = map[string]bool{"hour": true, "day": true, "week": true, "month": true}

// getStats is vision_get_stats: the dashboard's numbers for one site and range, through
// the same access check and query as GET /analytics/{siteId}/overview.
func (t *toolset) getStats(ctx context.Context, req *sdk.CallToolRequest, in statsInput) (*sdk.CallToolResult, statsOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, statsOutput{}, err
	}
	detailed, err := mcpkit.Detailed(in.ResponseFormat)
	if err != nil {
		return nil, statsOutput{}, err
	}
	granularity, err := granularityOf(in.Granularity)
	if err != nil {
		return nil, statsOutput{}, err
	}
	span, err := windowOf(in.From, in.To, time.Now())
	if err != nil {
		return nil, statsOutput{}, err
	}
	site, err := t.site(ctx, userID, in.SiteID)
	if err != nil {
		return nil, statsOutput{}, err
	}
	overview, err := t.Analytics.Overview(ctx, site.ID, span.from, span.to, granularity, filtersOf(in))
	if err != nil {
		return nil, statsOutput{}, mcpkit.Failure(err, "")
	}
	out := summaryOf(overview)
	out.Site, out.Range, out.LiveVisitors = t.siteOf(site), span.out(), t.Live.Count(site.ID)
	if detailed {
		out = withDetail(out, overview)
	}
	return nil, out, nil
}

// granularityOf reads the time series bucket, day by default.
func granularityOf(granularity string) (string, error) {
	if granularity == "" {
		return defaultSize, nil
	}
	if !granularities[granularity] {
		return "", errGranularity
	}
	return granularity, nil
}

// filtersOf maps the filter arguments to the dashboard's filters.
func filtersOf(in statsInput) analytics.Filters {
	return analytics.Filters{
		Country: in.Country, Browser: in.Browser, OS: in.OS, Device: in.Device, Path: in.Path, Referrer: in.Referrer,
	}
}

// summaryOf is the concise report: totals, the previous period and the top 10 lists.
func summaryOf(o *analytics.OverviewResponse) statsOutput {
	out := statsOutput{
		Pageviews: o.TotalPageviews, Visitors: o.UniqueVisitors, PrevPageviews: o.PrevTotalPageviews,
		PrevVisitors: o.PrevUniqueVisitors, BounceRate: o.BounceRate, AvgSessionSeconds: o.AvgSessionDuration,
		PagesPerSession: o.PagesPerSession,
	}
	out.TopPages = top(o.TopPages, func(s analytics.PageStat) countOut { return countOut{s.Path, s.Count} })
	out.TopReferrers = top(o.TopReferrers, func(s analytics.ReferrerStat) countOut { return countOut{s.Referrer, s.Count} })
	out.TopCountries = top(o.TopCountries, func(s analytics.CountryStat) countOut { return countOut{s.Country, s.Count} })
	out.TopDevices = top(o.TopDevices, func(s analytics.DeviceStat) countOut { return countOut{s.Device, s.Count} })
	out.TopBrowsers = top(o.TopBrowsers, func(s analytics.BrowserStat) countOut { return countOut{s.Browser, s.Count} })
	out.TopEvents = top(o.TopEvents, func(s analytics.EventStat) countOut { return countOut{s.Name, s.Count} })
	return out
}

// top renders the first topN rows of a list.
func top[S any](rows []S, render func(S) countOut) []countOut {
	out := make([]countOut, 0, min(len(rows), topN))
	for _, row := range rows[:min(len(rows), topN)] {
		out = append(out, render(row))
	}
	return out
}

// withDetail adds the whole overview, each list cut to the longest length that keeps the
// result within the budget.
func withDetail(out statsOutput, o *analytics.OverviewResponse) statsOutput {
	n := mcpkit.FitCount(listsLimit, func(n int) int {
		out.Detailed = clipped(o, n)
		return mcpkit.EncodedSize(out)
	})
	out.Detailed, out.DetailedListsLimit = clipped(o, n), n
	return out
}

// clipped is o with every top list cut to n rows; the time series stay whole.
func clipped(o *analytics.OverviewResponse, n int) *analytics.OverviewResponse {
	c := *o
	c.TopPages, c.TopReferrers, c.TopCountries = head(o.TopPages, n), head(o.TopReferrers, n), head(o.TopCountries, n)
	c.TopBrowsers, c.TopOS, c.TopDevices, c.TopScreens = head(o.TopBrowsers, n), head(o.TopOS, n), head(o.TopDevices, n), head(o.TopScreens, n)
	c.TopEntryPages, c.TopExitPages, c.TopEvents = head(o.TopEntryPages, n), head(o.TopExitPages, n), head(o.TopEvents, n)
	c.TopUTMSources, c.TopUTMMediums, c.TopUTMCampaigns = head(o.TopUTMSources, n), head(o.TopUTMMediums, n), head(o.TopUTMCampaigns, n)
	return &c
}

// head is the first n rows of a list.
func head[S any](rows []S, n int) []S {
	return rows[:min(len(rows), n)]
}
