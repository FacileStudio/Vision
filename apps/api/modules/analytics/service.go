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

func (s *Service) overview(ctx context.Context, siteID int64, from time.Time, to time.Time) (*OverviewResponse, error) {
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
		Select("to_char(created_at, 'YYYY-MM-DD') as date, count(*) as count").
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

	return &OverviewResponse{
		TotalPageviews:  totalPageviews,
		UniqueVisitors:  uniqueVisitors,
		TopPages:        topPages,
		TopReferrers:    topReferrers,
		TopCountries:    topCountries,
		TopBrowsers:     topBrowsers,
		TopOS:           topOS,
		TopDevices:      topDevices,
		PageviewsPerDay: pageviewsPerDay,
		TopScreens:      topScreens,
	}, nil
}

func (s *Service) realtimeVisitors(ctx context.Context, siteID int64) (int64, error) {
	var count int64
	err := s.orm.WithContext(ctx).
		Table("pageviews").
		Where("site_id = ? AND created_at >= ? AND visitor_id != ''", siteID, time.Now().Add(-5*time.Minute)).
		Distinct("visitor_id").
		Count(&count).Error
	if err != nil {
		return 0, errors.Internal("failed to count realtime visitors", err)
	}
	return count, nil
}
