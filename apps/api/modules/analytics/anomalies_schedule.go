package analytics

import (
	"context"
	"log/slog"
	"time"
)

const anomalyRetention = 180 * 24 * time.Hour

// Sweep detects anomalies for the given UTC days on every site that had a pageview in
// the last two weeks, so a site that went quiet is still checked, and drops anomalies
// older than the retention. One site failing does not stop the others.
func (s *Service) Sweep(ctx context.Context, days []time.Time, logger *slog.Logger) {
	var sites []int64
	err := s.orm.WithContext(ctx).Raw(`SELECT DISTINCT site_id FROM pageviews WHERE created_at >= ?`, time.Now().Add(-14*24*time.Hour)).Scan(&sites).Error
	if err != nil {
		logger.Error("anomalies: failed to list sites", slog.Any("error", err))
		return
	}
	for _, siteID := range sites {
		if _, err := s.Detect(ctx, siteID, days); err != nil {
			logger.Error("anomalies: detection failed", slog.Int64("site_id", siteID), slog.Any("error", err))
		}
	}
	if err := s.orm.WithContext(ctx).Exec(`DELETE FROM anomalies WHERE day < ?`, time.Now().Add(-anomalyRetention)).Error; err != nil {
		logger.Error("anomalies: failed to prune", slog.Any("error", err))
	}
}

// StartAnomalies runs a sweep over the last week at boot, then one over the last two
// completed UTC days every hour, until ctx ends. Vision runs one replica, so this is the
// one detector.
func (s *Service) StartAnomalies(ctx context.Context, logger *slog.Logger) {
	go func() {
		s.Sweep(ctx, completedDays(time.Now(), 7), logger)
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				s.Sweep(ctx, completedDays(now, 2), logger)
			}
		}
	}()
}

// completedDays lists the n UTC days before the one now falls in, oldest first.
func completedDays(now time.Time, n int) []time.Time {
	today := utcDay(now)
	days := make([]time.Time, n)
	for i := range days {
		days[i] = today.AddDate(0, 0, i-n)
	}
	return days
}
