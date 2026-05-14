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

	return &OverviewResponse{
		TotalPageviews:  totalPageviews,
		UniqueVisitors:  uniqueVisitors,
		TopPages:        topPages,
		TopReferrers:    topReferrers,
		TopCountries:    topCountries,
		PageviewsPerDay: pageviewsPerDay,
	}, nil
}
