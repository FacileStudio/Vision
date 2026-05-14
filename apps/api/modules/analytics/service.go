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

func (s *Service) Overview(ctx context.Context, siteID int64, from time.Time, to time.Time, granularity string) (*OverviewResponse, error) {
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
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Where("site_id = ? AND created_at >= ? AND created_at <= ?", siteID, from, to).
		Count(&totalPageviews).Error; err != nil {
		return nil, errors.Internal("failed to count pageviews", err)
	}

	var uniqueVisitors int64
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND visitor_id != ''", siteID, from, to).
		Distinct("visitor_id").
		Count(&uniqueVisitors).Error; err != nil {
		return nil, errors.Internal("failed to count visitors", err)
	}

	var topPages []PageStat
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Select("path, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ?", siteID, from, to).
		Group("path").
		Order("count desc").
		Limit(10).
		Scan(&topPages).Error; err != nil {
		return nil, errors.Internal("failed to query top pages", err)
	}

	var topReferrers []ReferrerStat
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Select("referrer, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND referrer != ''", siteID, from, to).
		Group("referrer").
		Order("count desc").
		Limit(10).
		Scan(&topReferrers).Error; err != nil {
		return nil, errors.Internal("failed to query top referrers", err)
	}

	var topCountries []CountryStat
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Select("country, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND country != ''", siteID, from, to).
		Group("country").
		Order("count desc").
		Limit(10).
		Scan(&topCountries).Error; err != nil {
		return nil, errors.Internal("failed to query top countries", err)
	}

	var topBrowsers []BrowserStat
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Select("browser, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND browser != ''", siteID, from, to).
		Group("browser").
		Order("count desc").
		Limit(10).
		Scan(&topBrowsers).Error; err != nil {
		return nil, errors.Internal("failed to query top browsers", err)
	}

	var topOS []OSStat
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Select("os, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND os != ''", siteID, from, to).
		Group("os").
		Order("count desc").
		Limit(10).
		Scan(&topOS).Error; err != nil {
		return nil, errors.Internal("failed to query top os", err)
	}

	var topDevices []DeviceStat
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Select("device, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND device != ''", siteID, from, to).
		Group("device").
		Order("count desc").
		Limit(10).
		Scan(&topDevices).Error; err != nil {
		return nil, errors.Internal("failed to query top devices", err)
	}

	var pageviewsPerDay []DayStat
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Select("to_char(created_at, '"+dateFormat+"') as date, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ?", siteID, from, to).
		Group("date").
		Order("date asc").
		Scan(&pageviewsPerDay).Error; err != nil {
		return nil, errors.Internal("failed to query pageviews per day", err)
	}

	var topScreens []ScreenStat
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Select("CASE WHEN screen_width < 768 THEN 'Mobile' WHEN screen_width < 1024 THEN 'Tablet' WHEN screen_width < 1440 THEN 'Laptop' ELSE 'Desktop' END as screen, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND screen_width > 0", siteID, from, to).
		Group("screen").
		Order("count desc").
		Scan(&topScreens).Error; err != nil {
		return nil, errors.Internal("failed to query top screens", err)
	}

	var uniqueVisitorsPerDay []DayStat
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Select("to_char(created_at, '"+dateFormat+"') as date, count(distinct visitor_id) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND visitor_id != ''", siteID, from, to).
		Group("date").
		Order("date asc").
		Scan(&uniqueVisitorsPerDay).Error; err != nil {
		return nil, errors.Internal("failed to query unique visitors per day", err)
	}

	var hourlyDistribution []HourStat
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Select("EXTRACT(HOUR FROM created_at)::int as hour, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ?", siteID, from, to).
		Group("hour").
		Order("hour asc").
		Scan(&hourlyDistribution).Error; err != nil {
		return nil, errors.Internal("failed to query hourly distribution", err)
	}

	duration := to.Sub(from)
	prevFrom := from.Add(-duration)
	prevTo := from.Add(-time.Nanosecond)

	var prevTotalPageviews int64
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Where("site_id = ? AND created_at >= ? AND created_at <= ?", siteID, prevFrom, prevTo).
		Count(&prevTotalPageviews).Error; err != nil {
		return nil, errors.Internal("failed to count previous pageviews", err)
	}

	var prevUniqueVisitors int64
	if err := s.orm.WithContext(ctx).
		Table("pageviews").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND visitor_id != ''", siteID, prevFrom, prevTo).
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
		Group("entry_path").Order("count desc").Limit(10).
		Scan(&topEntryPages)

	var topExitPages []PageStat
	s.orm.WithContext(ctx).Table("visitor_sessions").
		Select("exit_path as path, count(*) as count").
		Where("site_id = ? AND started_at >= ? AND started_at <= ?", siteID, from, to).
		Group("exit_path").Order("count desc").Limit(10).
		Scan(&topExitPages)

	var topUTMSources []UTMStat
	s.orm.WithContext(ctx).Table("pageviews").
		Select("utm_source as value, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND utm_source != ''", siteID, from, to).
		Group("utm_source").Order("count desc").Limit(10).
		Scan(&topUTMSources)

	var topUTMMediums []UTMStat
	s.orm.WithContext(ctx).Table("pageviews").
		Select("utm_medium as value, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND utm_medium != ''", siteID, from, to).
		Group("utm_medium").Order("count desc").Limit(10).
		Scan(&topUTMMediums)

	var topUTMCampaigns []UTMStat
	s.orm.WithContext(ctx).Table("pageviews").
		Select("utm_campaign as value, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND utm_campaign != ''", siteID, from, to).
		Group("utm_campaign").Order("count desc").Limit(10).
		Scan(&topUTMCampaigns)

	var prevPageviewsPerDay []DayStat
	s.orm.WithContext(ctx).Table("pageviews").
		Select("to_char(created_at, '"+dateFormat+"') as date, count(*) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ?", siteID, prevFrom, prevTo).
		Group("date").Order("date asc").
		Scan(&prevPageviewsPerDay)

	var prevUniqueVisitorsPerDay []DayStat
	s.orm.WithContext(ctx).Table("pageviews").
		Select("to_char(created_at, '"+dateFormat+"') as date, count(distinct visitor_id) as count").
		Where("site_id = ? AND created_at >= ? AND created_at <= ? AND visitor_id != ''", siteID, prevFrom, prevTo).
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
	}, nil
}

