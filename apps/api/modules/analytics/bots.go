package analytics

import (
	"context"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

const (
	botMaxDays    = 366
	botCategories = `SELECT category AS name, SUM(hits) AS hits FROM bot_hits
WHERE site_id = ? AND day >= ? AND day <= ? GROUP BY category ORDER BY hits DESC, name ASC`
	botNames = `SELECT name, category, SUM(hits) AS hits FROM bot_hits
WHERE site_id = ? AND day >= ? AND day <= ? GROUP BY name, category ORDER BY hits DESC, name ASC, category ASC LIMIT 20`
	botPaths = `SELECT path, SUM(hits) AS hits FROM bot_hits
WHERE site_id = ? AND day >= ? AND day <= ? GROUP BY path ORDER BY hits DESC, path ASC LIMIT 20`
	botDays = `SELECT to_char(day AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS day, SUM(hits) AS hits FROM bot_hits
WHERE site_id = ? AND day >= ? AND day <= ? GROUP BY 1`
	interactedVisitorsSQL = `SELECT COUNT(DISTINCT visitor_id) FROM visitor_sessions
WHERE site_id = ? AND started_at >= ? AND started_at <= ? AND interacted AND visitor_id != ''`
)

// BotCategoryStat sums the hits of one bot category.
type BotCategoryStat struct {
	Name string `json:"name"`
	Hits int64  `json:"hits"`
}

// BotStat sums the hits of one named bot.
type BotStat struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Hits     int64  `json:"hits"`
}

// BotPathStat sums the bot hits on one path.
type BotPathStat struct {
	Path string `json:"path"`
	Hits int64  `json:"hits"`
}

// BotDayStat sums the bot hits of one UTC day.
type BotDayStat struct {
	Day  string `json:"day"`
	Hits int64  `json:"hits"`
}

// BotsResponse is GET /analytics/{siteId}/bots: bot and crawler traffic, kept apart
// from pageviews. Lists are sorted by hits, then name; bots and paths keep the top 20.
type BotsResponse struct {
	Total      int64             `json:"total"`
	Categories []BotCategoryStat `json:"categories"`
	Bots       []BotStat         `json:"bots"`
	Paths      []BotPathStat     `json:"paths"`
	Series     []BotDayStat      `json:"series"`
}

// Bots aggregates the site's bot hits over the UTC days from..to, both included.
func (s *Service) Bots(ctx context.Context, siteID int64, from time.Time, to time.Time) (*BotsResponse, error) {
	first, last := utcDay(from), utcDay(to)
	if last.Sub(first) > botMaxDays*24*time.Hour {
		return nil, errors.Invalid("the period is at most 366 days")
	}
	resp := &BotsResponse{}
	args := []any{siteID, first, last}
	var days []BotDayStat
	for _, step := range []struct {
		sql  string
		into any
	}{{botCategories, &resp.Categories}, {botNames, &resp.Bots}, {botPaths, &resp.Paths}, {botDays, &days}} {
		if err := s.orm.WithContext(ctx).Raw(step.sql, args...).Scan(step.into).Error; err != nil {
			return nil, errors.Internal("failed to query bot hits", err)
		}
	}
	resp.Categories, resp.Bots, resp.Paths = nonNil(resp.Categories), nonNil(resp.Bots), nonNil(resp.Paths)
	resp.Total = sumHits(resp.Categories)
	resp.Series = fillBotSeries(first, last, days)
	return resp, nil
}

// sumHits adds up the hits of every category.
func sumHits(categories []BotCategoryStat) int64 {
	var total int64
	for _, category := range categories {
		total += category.Hits
	}
	return total
}

// fillBotSeries lists every day from first to last, with zero for a day without hits.
func fillBotSeries(first time.Time, last time.Time, days []BotDayStat) []BotDayStat {
	hits := make(map[string]int64, len(days))
	for _, day := range days {
		hits[day.Day] = day.Hits
	}
	series := []BotDayStat{}
	for day := first; !day.After(last); day = day.AddDate(0, 0, 1) {
		key := day.Format(dayKeyLayout)
		series = append(series, BotDayStat{Day: key, Hits: hits[key]})
	}
	return series
}

// nonNil turns a nil slice into an empty one so it encodes as [].
func nonNil[T any](list []T) []T {
	if list == nil {
		return []T{}
	}
	return list
}

// interactedVisitors counts the distinct visitors with a session started in the
// period that made a real gesture. It ignores the pageview filters, like the other
// session figures of the overview.
func (s *Service) interactedVisitors(ctx context.Context, siteID int64, from time.Time, to time.Time) (int64, error) {
	var count int64
	if err := s.orm.WithContext(ctx).Raw(interactedVisitorsSQL, siteID, from, to).Scan(&count).Error; err != nil {
		return 0, errors.Internal("failed to count interacted visitors", err)
	}
	return count, nil
}
