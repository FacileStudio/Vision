package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"api/internal/errors"
	"api/modules/analytics"
	"api/schemas"

	"gorm.io/gorm"
)

func StartScheduler(orm *gorm.DB, analyticsService *analytics.Service) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			processDueWebhooks(orm, analyticsService)
		}
	}()
}

func processDueWebhooks(orm *gorm.DB, analyticsService *analytics.Service) {
	var webhooks []schemas.Webhook
	if err := orm.Where("enabled = ?", true).Find(&webhooks).Error; err != nil {
		slog.Error("failed to load webhooks", slog.Any("error", err))
		return
	}

	now := time.Now().UTC()
	for i := range webhooks {
		wh := &webhooks[i]
		if !isDue(wh, now) {
			continue
		}

		if err := sendReport(orm, analyticsService, 0, wh.SiteID, wh.ID); err != nil {
			slog.Error("failed to send webhook report",
				slog.Int64("webhook_id", wh.ID),
				slog.Int64("site_id", wh.SiteID),
				slog.Any("error", err),
			)
			continue
		}

		orm.Model(wh).Update("last_sent_at", now)
	}
}

func isDue(wh *schemas.Webhook, now time.Time) bool {
	if wh.LastSentAt == nil {
		return true
	}
	elapsed := now.Sub(*wh.LastSentAt)
	switch wh.Period {
	case "hourly":
		return elapsed >= time.Hour
	case "daily":
		return elapsed >= 24*time.Hour
	case "weekly":
		return elapsed >= 7*24*time.Hour
	case "monthly":
		return elapsed >= 30*24*time.Hour
	}
	return false
}

func sendReport(orm *gorm.DB, analyticsService *analytics.Service, ownerID int64, siteID int64, webhookID int64) error {
	var wh schemas.Webhook
	if err := orm.Where("id = ? AND site_id = ?", webhookID, siteID).First(&wh).Error; err != nil {
		return errors.NotFound("webhook not found")
	}

	if ownerID > 0 {
		var count int64
		orm.Table("sites").Where("id = ? AND owner_id = ?", siteID, ownerID).Count(&count)
		if count == 0 {
			return errors.NotFound("site not found")
		}
	}

	var site schemas.Site
	if err := orm.First(&site, siteID).Error; err != nil {
		return errors.NotFound("site not found")
	}

	now := time.Now().UTC()
	from := periodStart(wh.Period, now)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	overview, err := analyticsService.Overview(ctx, siteID, from, now)
	if err != nil {
		return err
	}

	var vpv float64
	if overview.UniqueVisitors > 0 {
		vpv = float64(overview.TotalPageviews) / float64(overview.UniqueVisitors)
	}

	var pvChange, vChange float64
	if overview.PrevTotalPageviews > 0 {
		pvChange = float64(overview.TotalPageviews-overview.PrevTotalPageviews) / float64(overview.PrevTotalPageviews) * 100
	}
	if overview.PrevUniqueVisitors > 0 {
		vChange = float64(overview.UniqueVisitors-overview.PrevUniqueVisitors) / float64(overview.PrevUniqueVisitors) * 100
	}

	topPages := make([]TopItem, len(overview.TopPages))
	for i, p := range overview.TopPages {
		topPages[i] = TopItem{Name: p.Path, Count: p.Count}
	}
	topReferrers := make([]TopItem, len(overview.TopReferrers))
	for i, r := range overview.TopReferrers {
		topReferrers[i] = TopItem{Name: r.Referrer, Count: r.Count}
	}
	topCountries := make([]TopItem, len(overview.TopCountries))
	for i, c := range overview.TopCountries {
		topCountries[i] = TopItem{Name: c.Country, Count: c.Count}
	}
	topBrowsers := make([]TopItem, len(overview.TopBrowsers))
	for i, b := range overview.TopBrowsers {
		topBrowsers[i] = TopItem{Name: b.Browser, Count: b.Count}
	}
	topDevices := make([]TopItem, len(overview.TopDevices))
	for i, d := range overview.TopDevices {
		topDevices[i] = TopItem{Name: d.Device, Count: d.Count}
	}

	payload := ReportPayload{
		EventType: "analytics.report",
		Site: ReportSite{
			ID:     site.ID,
			Name:   site.Name,
			Domain: site.Domain,
		},
		Period: ReportPeriod{
			Type: wh.Period,
			From: from.Format(time.RFC3339),
			To:   now.Format(time.RFC3339),
		},
		Metrics: ReportMetrics{
			TotalPageviews:     overview.TotalPageviews,
			UniqueVisitors:     overview.UniqueVisitors,
			ViewsPerVisitor:    vpv,
			PrevTotalPageviews: overview.PrevTotalPageviews,
			PrevUniqueVisitors: overview.PrevUniqueVisitors,
			PageviewsChangePct: pvChange,
			VisitorsChangePct:  vChange,
		},
		TopPages:     topPages,
		TopReferrers: topReferrers,
		TopCountries: topCountries,
		TopBrowsers:  topBrowsers,
		TopDevices:   topDevices,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return errors.Internal("failed to marshal report payload", err)
	}

	signature := sign(wh.Secret, body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(body))
	if err != nil {
		return errors.Internal("failed to build request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Vision-Webhook/1.0")
	req.Header.Set("x-vision-signature-256", fmt.Sprintf("sha256=%s", signature))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Internal("failed to deliver webhook", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.Internal(fmt.Sprintf("webhook endpoint returned %d", resp.StatusCode), nil)
	}

	return nil
}

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func periodStart(period string, now time.Time) time.Time {
	switch period {
	case "hourly":
		return now.Add(-time.Hour)
	case "daily":
		return now.Add(-24 * time.Hour)
	case "weekly":
		return now.Add(-7 * 24 * time.Hour)
	case "monthly":
		return now.Add(-30 * 24 * time.Hour)
	}
	return now.Add(-24 * time.Hour)
}
