package analytics

import (
	"context"
	"time"

	"api/internal/errors"

	"gorm.io/gorm"
)

type Service struct {
	orm *gorm.DB
}

func NewService(orm *gorm.DB) *Service {
	return &Service{orm: orm}
}

func applyFilters(db *gorm.DB, siteID int64, from time.Time, to time.Time, filters Filters) *gorm.DB {
	q := db.Where("site_id = ? AND created_at >= ? AND created_at <= ?", siteID, from, to)
	if filters.Country != "" {
		q = q.Where("country = ?", filters.Country)
	}
	if filters.Browser != "" {
		q = q.Where("browser = ?", filters.Browser)
	}
	if filters.OS != "" {
		q = q.Where("os = ?", filters.OS)
	}
	if filters.Device != "" {
		q = q.Where("device = ?", filters.Device)
	}
	if filters.Path != "" {
		q = q.Where("path LIKE ?", "%"+filters.Path+"%")
	}
	if filters.Referrer != "" {
		q = q.Where("referrer LIKE ?", "%"+filters.Referrer+"%")
	}
	return q
}

func (s *Service) Overview(ctx context.Context, siteID int64, from time.Time, to time.Time, granularity string, filters Filters) (*OverviewResponse, error) {
	var dateFormat string
	switch granularity {
	case "hour":
		dateFormat = "YYYY-MM-DD HH24:00"
	case "week":
		dateFormat = "IYYY-IW"
	case "month":
		dateFormat = "YYYY-MM"
	default:
		dateFormat = "YYYY-MM-DD"
	}

	var totalPageviews int64
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews"), siteID, from, to, filters).
		Count(&totalPageviews).Error; err != nil {
		return nil, errors.Internal("failed to count pageviews", err)
	}

	var uniqueVisitors int64
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews"), siteID, from, to, filters).
		Where("visitor_id != ''").
		Distinct("visitor_id").
		Count(&uniqueVisitors).Error; err != nil {
		return nil, errors.Internal("failed to count visitors", err)
	}

	var topPages []PageStat
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("path, count(*) as count"), siteID, from, to, filters).
		Group("path").
		Order("count desc").
		Limit(100).
		Scan(&topPages).Error; err != nil {
		return nil, errors.Internal("failed to query top pages", err)
	}

	var topReferrers []ReferrerStat
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("referrer, count(*) as count"), siteID, from, to, filters).
		Where("referrer != ''").
		Group("referrer").
		Order("count desc").
		Limit(100).
		Scan(&topReferrers).Error; err != nil {
		return nil, errors.Internal("failed to query top referrers", err)
	}

	var topCountries []CountryStat
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("country, count(*) as count"), siteID, from, to, filters).
		Where("country != ''").
		Group("country").
		Order("count desc").
		Limit(100).
		Scan(&topCountries).Error; err != nil {
		return nil, errors.Internal("failed to query top countries", err)
	}

	var topBrowsers []BrowserStat
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("browser, count(*) as count"), siteID, from, to, filters).
		Where("browser != ''").
		Group("browser").
		Order("count desc").
		Limit(100).
		Scan(&topBrowsers).Error; err != nil {
		return nil, errors.Internal("failed to query top browsers", err)
	}

	var topOS []OSStat
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("os, count(*) as count"), siteID, from, to, filters).
		Where("os != ''").
		Group("os").
		Order("count desc").
		Limit(100).
		Scan(&topOS).Error; err != nil {
		return nil, errors.Internal("failed to query top os", err)
	}

	var topDevices []DeviceStat
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("device, count(*) as count"), siteID, from, to, filters).
		Where("device != ''").
		Group("device").
		Order("count desc").
		Limit(100).
		Scan(&topDevices).Error; err != nil {
		return nil, errors.Internal("failed to query top devices", err)
	}

	var pageviewsPerDay []DayStat
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("to_char(created_at, '"+dateFormat+"') as date, count(*) as count"), siteID, from, to, filters).
		Group("date").
		Order("date asc").
		Scan(&pageviewsPerDay).Error; err != nil {
		return nil, errors.Internal("failed to query pageviews per day", err)
	}

	var topScreens []ScreenStat
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("CASE WHEN screen_width < 768 THEN 'Mobile' WHEN screen_width < 1024 THEN 'Tablet' WHEN screen_width < 1440 THEN 'Laptop' ELSE 'Desktop' END as screen, count(*) as count"), siteID, from, to, filters).
		Where("screen_width > 0").
		Group("screen").
		Order("count desc").
		Scan(&topScreens).Error; err != nil {
		return nil, errors.Internal("failed to query top screens", err)
	}

	var uniqueVisitorsPerDay []DayStat
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("to_char(created_at, '"+dateFormat+"') as date, count(distinct visitor_id) as count"), siteID, from, to, filters).
		Where("visitor_id != ''").
		Group("date").
		Order("date asc").
		Scan(&uniqueVisitorsPerDay).Error; err != nil {
		return nil, errors.Internal("failed to query unique visitors per day", err)
	}

	var hourlyDistribution []HourStat
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("EXTRACT(HOUR FROM created_at)::int as hour, count(*) as count"), siteID, from, to, filters).
		Group("hour").
		Order("hour asc").
		Scan(&hourlyDistribution).Error; err != nil {
		return nil, errors.Internal("failed to query hourly distribution", err)
	}

	duration := to.Sub(from)
	prevFrom := from.Add(-duration)
	prevTo := from.Add(-time.Nanosecond)

	var prevTotalPageviews int64
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews"), siteID, prevFrom, prevTo, filters).
		Count(&prevTotalPageviews).Error; err != nil {
		return nil, errors.Internal("failed to count previous pageviews", err)
	}

	var prevUniqueVisitors int64
	if err := applyFilters(s.orm.WithContext(ctx).Table("pageviews"), siteID, prevFrom, prevTo, filters).
		Where("visitor_id != ''").
		Distinct("visitor_id").
		Count(&prevUniqueVisitors).Error; err != nil {
		return nil, errors.Internal("failed to count previous visitors", err)
	}

	var totalSessions int64
	s.orm.WithContext(ctx).Table("visitor_sessions").
		Where("site_id = ? AND started_at >= ? AND started_at <= ?", siteID, from, to).
		Count(&totalSessions)

	var bounceSessions int64
	s.orm.WithContext(ctx).Table("visitor_sessions").
		Where("site_id = ? AND started_at >= ? AND started_at <= ? AND is_bounce = ?", siteID, from, to, true).
		Count(&bounceSessions)

	var bounceRate float64
	if totalSessions > 0 {
		bounceRate = float64(bounceSessions) / float64(totalSessions) * 100
	}

	type SessionAvgs struct {
		AvgDuration float64
		AvgPages    float64
	}
	var avgs SessionAvgs
	s.orm.WithContext(ctx).Table("visitor_sessions").
		Select("COALESCE(AVG(duration), 0) as avg_duration, COALESCE(AVG(pageview_count), 0) as avg_pages").
		Where("site_id = ? AND started_at >= ? AND started_at <= ?", siteID, from, to).
		Scan(&avgs)

	var prevTotalSessions int64
	s.orm.WithContext(ctx).Table("visitor_sessions").
		Where("site_id = ? AND started_at >= ? AND started_at <= ?", siteID, prevFrom, prevTo).
		Count(&prevTotalSessions)

	var prevBounceSessions int64
	s.orm.WithContext(ctx).Table("visitor_sessions").
		Where("site_id = ? AND started_at >= ? AND started_at <= ? AND is_bounce = ?", siteID, prevFrom, prevTo, true).
		Count(&prevBounceSessions)

	var prevBounceRate float64
	if prevTotalSessions > 0 {
		prevBounceRate = float64(prevBounceSessions) / float64(prevTotalSessions) * 100
	}

	var prevAvgs SessionAvgs
	s.orm.WithContext(ctx).Table("visitor_sessions").
		Select("COALESCE(AVG(duration), 0) as avg_duration, COALESCE(AVG(pageview_count), 0) as avg_pages").
		Where("site_id = ? AND started_at >= ? AND started_at <= ?", siteID, prevFrom, prevTo).
		Scan(&prevAvgs)

	var topEntryPages []PageStat
	s.orm.WithContext(ctx).Table("visitor_sessions").
		Select("entry_path as path, count(*) as count").
		Where("site_id = ? AND started_at >= ? AND started_at <= ?", siteID, from, to).
		Group("entry_path").Order("count desc").Limit(100).
		Scan(&topEntryPages)

	var topExitPages []PageStat
	s.orm.WithContext(ctx).Table("visitor_sessions").
		Select("exit_path as path, count(*) as count").
		Where("site_id = ? AND started_at >= ? AND started_at <= ?", siteID, from, to).
		Group("exit_path").Order("count desc").Limit(100).
		Scan(&topExitPages)

	var topUTMSources []UTMStat
	applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("utm_source as value, count(*) as count"), siteID, from, to, filters).
		Where("utm_source != ''").
		Group("utm_source").Order("count desc").Limit(100).
		Scan(&topUTMSources)

	var topUTMMediums []UTMStat
	applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("utm_medium as value, count(*) as count"), siteID, from, to, filters).
		Where("utm_medium != ''").
		Group("utm_medium").Order("count desc").Limit(100).
		Scan(&topUTMMediums)

	var topUTMCampaigns []UTMStat
	applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("utm_campaign as value, count(*) as count"), siteID, from, to, filters).
		Where("utm_campaign != ''").
		Group("utm_campaign").Order("count desc").Limit(100).
		Scan(&topUTMCampaigns)

	var prevPageviewsPerDay []DayStat
	applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("to_char(created_at, '"+dateFormat+"') as date, count(*) as count"), siteID, prevFrom, prevTo, filters).
		Group("date").Order("date asc").
		Scan(&prevPageviewsPerDay)

	var prevUniqueVisitorsPerDay []DayStat
	applyFilters(s.orm.WithContext(ctx).Table("pageviews").Select("to_char(created_at, '"+dateFormat+"') as date, count(distinct visitor_id) as count"), siteID, prevFrom, prevTo, filters).
		Where("visitor_id != ''").
		Group("date").Order("date asc").
		Scan(&prevUniqueVisitorsPerDay)

	if topPages == nil {
		topPages = []PageStat{}
	}
	if topReferrers == nil {
		topReferrers = []ReferrerStat{}
	}
	if topCountries == nil {
		topCountries = []CountryStat{}
	}
	if pageviewsPerDay == nil {
		pageviewsPerDay = []DayStat{}
	}
	if topBrowsers == nil {
		topBrowsers = []BrowserStat{}
	}
	if topOS == nil {
		topOS = []OSStat{}
	}
	if topDevices == nil {
		topDevices = []DeviceStat{}
	}
	if topScreens == nil {
		topScreens = []ScreenStat{}
	}
	if uniqueVisitorsPerDay == nil {
		uniqueVisitorsPerDay = []DayStat{}
	}
	if hourlyDistribution == nil {
		hourlyDistribution = []HourStat{}
	}
	if topEntryPages == nil {
		topEntryPages = []PageStat{}
	}
	if topExitPages == nil {
		topExitPages = []PageStat{}
	}
	if topUTMSources == nil {
		topUTMSources = []UTMStat{}
	}
	if topUTMMediums == nil {
		topUTMMediums = []UTMStat{}
	}
	if topUTMCampaigns == nil {
		topUTMCampaigns = []UTMStat{}
	}
	if prevPageviewsPerDay == nil {
		prevPageviewsPerDay = []DayStat{}
	}
	if prevUniqueVisitorsPerDay == nil {
		prevUniqueVisitorsPerDay = []DayStat{}
	}

	var perf PerformanceStats
	s.orm.WithContext(ctx).Table("pageviews").
		Select("COALESCE(AVG(perf_dns), 0) as avg_dns, COALESCE(AVG(perf_tcp), 0) as avg_tcp, COALESCE(AVG(perf_ttfb), 0) as avg_ttfb, COALESCE(AVG(perf_dom_load), 0) as avg_dom_load, COALESCE(AVG(perf_page_load), 0) as avg_page_load, COUNT(perf_page_load) as sample_count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND perf_page_load IS NOT NULL", siteID, from, to).
		Scan(&perf)

	var topEvents []EventStat
	s.orm.WithContext(ctx).Table("custom_events").
		Select("name, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ?", siteID, from, to).
		Group("name").
		Order("count desc").
		Limit(100).
		Scan(&topEvents)

	if topEvents == nil {
		topEvents = []EventStat{}
	}

	return &OverviewResponse{
		TotalPageviews:  totalPageviews,
		UniqueVisitors:  uniqueVisitors,
		TopPages:        topPages,
		TopReferrers:    topReferrers,
		TopCountries:    topCountries,
		TopBrowsers:     topBrowsers,
		TopOS:           topOS,
		TopDevices:      topDevices,
		PageviewsPerDay:      pageviewsPerDay,
		TopScreens:           topScreens,
		UniqueVisitorsPerDay: uniqueVisitorsPerDay,
		HourlyDistribution:   hourlyDistribution,
		PrevTotalPageviews:     prevTotalPageviews,
		PrevUniqueVisitors:     prevUniqueVisitors,
		BounceRate:             bounceRate,
		AvgSessionDuration:     avgs.AvgDuration,
		PagesPerSession:        avgs.AvgPages,
		PrevBounceRate:           prevBounceRate,
		PrevAvgSessionDuration:   prevAvgs.AvgDuration,
		PrevPagesPerSession:      prevAvgs.AvgPages,
		TopEntryPages:            topEntryPages,
		TopExitPages:             topExitPages,
		TopUTMSources:            topUTMSources,
		TopUTMMediums:            topUTMMediums,
		TopUTMCampaigns:          topUTMCampaigns,
		PrevPageviewsPerDay:      prevPageviewsPerDay,
		PrevUniqueVisitorsPerDay: prevUniqueVisitorsPerDay,
		Performance:              &perf,
		TopEvents:                topEvents,
	}, nil
}
