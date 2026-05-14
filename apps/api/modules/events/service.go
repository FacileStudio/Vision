package events

import (
	"context"
	stderrors "errors"
	"strings"

	"api/internal/errors"
	"api/schemas"

	"gorm.io/gorm"
)

type Service struct {
	orm *gorm.DB
}

func NewService(orm *gorm.DB) *Service {
	return &Service{orm: orm}
}

func (s *Service) resolveSite(ctx context.Context, apiKey string) (*schemas.Site, error) {
	if apiKey == "" {
		return nil, errors.Unauthorized("missing API key")
	}

	var site schemas.Site
	err := s.orm.WithContext(ctx).Where("api_key = ?", apiKey).First(&site).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Unauthorized("invalid API key")
	}
	if err != nil {
		return nil, errors.Internal("failed to resolve site", err)
	}
	return &site, nil
}

func (s *Service) recordPageview(ctx context.Context, site *schemas.Site, req *PageviewRequest, userAgent string, country string) error {
	path := strings.TrimSpace(req.Path)
	if path == "" {
		path = "/"
	}

	record := &schemas.Pageview{
		SiteID:    site.ID,
		Path:      path,
		Referrer:  strings.TrimSpace(req.Referrer),
		UserAgent: userAgent,
		Language:  strings.TrimSpace(req.Language),
		Country:   country,
		VisitorID: strings.TrimSpace(req.VisitorID),
	}
	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		return errors.Internal("failed to record pageview", err)
	}
	return nil
}
