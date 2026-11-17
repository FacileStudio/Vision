package events

import (
	"context"
	"strings"
	"time"

	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// recordBotHit counts one request from a classified bot against the site, path, category and
// bot name for the current UTC day. It never keeps an address or a visitor id.
func (s *Service) recordBotHit(ctx context.Context, site *schemas.Site, path, category, name string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}
	day := time.Now().UTC().Truncate(24 * time.Hour)
	err := s.orm.WithContext(ctx).Exec(
		`INSERT INTO bot_hits (site_id, day, category, name, path, hits) VALUES (?, ?, ?, ?, ?, 1)
		ON CONFLICT ON CONSTRAINT bot_hits_unique DO UPDATE SET hits = bot_hits.hits + 1`,
		site.ID, day, category, truncateRunes(name, 128), truncateRunes(path, 512),
	).Error
	if err != nil {
		return errors.Internal("failed to record bot hit", err)
	}
	return nil
}

// markInteracted flags the visitor's newest open session, one that ended in the last 30
// minutes, as having seen a real gesture.
func (s *Service) markInteracted(ctx context.Context, siteID int64, visitorID string) error {
	cutoff := time.Now().Add(-30 * time.Minute)
	return s.orm.WithContext(ctx).Exec(
		`UPDATE visitor_sessions SET interacted = true WHERE id = (
			SELECT id FROM visitor_sessions WHERE site_id = ? AND visitor_id = ? AND ended_at >= ?
			ORDER BY ended_at DESC LIMIT 1)`,
		siteID, visitorID, cutoff,
	).Error
}

// truncateRunes cuts value to at most limit runes without splitting a character.
func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
