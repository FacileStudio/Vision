package events

import (
	"context"
	stderrors "errors"
	"net/url"
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

func (s *Service) resolveSiteByOrigin(ctx context.Context, origin string, referer string) (*schemas.Site, error) {
	host := extractHost(origin)
	if host == "" {
		host = extractHost(referer)
	}
	if host == "" {
		return nil, errors.Forbidden("missing origin")
	}

	var site schemas.Site
	err := s.orm.WithContext(ctx).Where("LOWER(domain) = ?", host).First(&site).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Forbidden("unregistered domain")
	}
	if err != nil {
		return nil, errors.Internal("failed to resolve site", err)
	}
	return &site, nil
}

func extractHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(parsed.Hostname())
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
