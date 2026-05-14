package events

import (
	"context"
	stderrors "errors"
	"net/url"
	"strings"
	"time"

	"encoding/json"

	"api/internal/errors"
	"api/schemas"

	"gorm.io/gorm"
)

type Service struct {
	orm *gorm.DB
	hub *Hub
}

func NewService(orm *gorm.DB) *Service {
	return &Service{orm: orm}
}

func (s *Service) SetHub(hub *Hub) {
	s.hub = hub
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

	browser, os, device := parseUserAgent(userAgent)

	record := &schemas.Pageview{
		SiteID:    site.ID,
		Path:      path,
		Referrer:  strings.TrimSpace(req.Referrer),
		UserAgent: userAgent,
		Browser:   browser,
		OS:        os,
		Device:    device,
		Language:  strings.TrimSpace(req.Language),
		Country:   country,
		VisitorID:   strings.TrimSpace(req.VisitorID),
		ScreenWidth: req.ScreenWidth,
		UTMSource:   strings.TrimSpace(req.UTMSource),
		UTMMedium:   strings.TrimSpace(req.UTMMedium),
		UTMCampaign: strings.TrimSpace(req.UTMCampaign),
		UTMTerm:     strings.TrimSpace(req.UTMTerm),
		UTMContent:  strings.TrimSpace(req.UTMContent),
	}
	if req.Performance != nil {
		record.PerfDNS = &req.Performance.DNS
		record.PerfTCP = &req.Performance.TCP
		record.PerfTTFB = &req.Performance.TTFB
		record.PerfDOMLoad = &req.Performance.DOMLoad
		record.PerfPageLoad = &req.Performance.PageLoad
	}
	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		return errors.Internal("failed to record pageview", err)
	}

	s.updateSession(ctx, site.ID, record.VisitorID, record.Path, record.CreatedAt)

	if s.hub != nil {
		s.hub.Broadcast(PageviewEvent{
			SiteID:    site.ID,
			Path:      record.Path,
			Referrer:  record.Referrer,
			Country:   record.Country,
			VisitorID: record.VisitorID,
			Timestamp: record.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	return nil
}

func (s *Service) updateSession(ctx context.Context, siteID int64, visitorID string, path string, timestamp time.Time) {
	if visitorID == "" {
		return
	}

	sessionGap := 30 * time.Minute
	cutoff := timestamp.Add(-sessionGap)

	_ = s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing schemas.VisitorSession
		err := tx.
			Where("site_id = ? AND visitor_id = ? AND ended_at >= ?", siteID, visitorID, cutoff).
			Order("ended_at DESC").
			First(&existing).Error

		if err == nil {
			existing.EndedAt = timestamp
			existing.ExitPath = path
			existing.PageviewCount++
			existing.Duration = int(existing.EndedAt.Sub(existing.StartedAt).Seconds())
			existing.IsBounce = false
			return tx.Save(&existing).Error
		}

		session := &schemas.VisitorSession{
			SiteID:        siteID,
			VisitorID:     visitorID,
			StartedAt:     timestamp,
			EndedAt:       timestamp,
			EntryPath:     path,
			ExitPath:      path,
			PageviewCount: 1,
			Duration:      0,
			IsBounce:      true,
		}
		return tx.Create(session).Error
	})
}

func (s *Service) recordCustomEvent(ctx context.Context, site *schemas.Site, req *CustomEventRequest) error {
	name := strings.TrimSpace(req.EventName)
	if name == "" {
		return errors.Invalid("event name is required")
	}

	propsJSON := "{}"
	if req.EventProps != nil {
		b, err := json.Marshal(req.EventProps)
		if err == nil {
			propsJSON = string(b)
		}
	}

	record := &schemas.CustomEvent{
		SiteID:    site.ID,
		VisitorID: strings.TrimSpace(req.VisitorID),
		Path:      strings.TrimSpace(req.Path),
		Name:      name,
		Props:     propsJSON,
	}
	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		return errors.Internal("failed to record custom event", err)
	}
	return nil
}

func (s *Service) updatePerformance(ctx context.Context, site *schemas.Site, visitorID string, path string, perf *PerformanceData) {
	if perf == nil || visitorID == "" {
		return
	}
	s.orm.WithContext(ctx).
		Model(&schemas.Pageview{}).
		Where("site_id = ? AND visitor_id = ? AND path = ? AND perf_page_load IS NULL", site.ID, visitorID, path).
		Order("created_at DESC").
		Limit(1).
		Updates(map[string]interface{}{
			"perf_dns":       perf.DNS,
			"perf_tcp":       perf.TCP,
			"perf_ttfb":      perf.TTFB,
			"perf_dom_load":  perf.DOMLoad,
			"perf_page_load": perf.PageLoad,
		})
}
