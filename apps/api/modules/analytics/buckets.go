package analytics

import "time"

// summaryUnit buckets by hour for windows of two days or less, by day otherwise.
func summaryUnit(from time.Time, to time.Time) string {
	if to.Sub(from) <= 48*time.Hour {
		return "hour"
	}
	return "day"
}

// bucketKey formats a UTC bucket start the way Series dates read.
func bucketKey(t time.Time, unit string) string {
	if unit == "hour" {
		return t.UTC().Format("2006-01-02T15:00")
	}
	return t.UTC().Format("2006-01-02")
}

// bucketStarts lists every UTC bucket start the window touches, oldest first.
func bucketStarts(from time.Time, to time.Time, unit string) []time.Time {
	start := from.UTC().Truncate(time.Hour)
	step := time.Hour
	if unit == "day" {
		y, m, d := from.UTC().Date()
		start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
		step = 24 * time.Hour
	}
	var out []time.Time
	for t := start; !t.After(to); t = t.Add(step) {
		out = append(out, t)
	}
	return out
}

// bucketCounts indexes the grouped series rows by site, then by bucket key.
func bucketCounts(buckets []siteBucket, unit string) map[int64]map[string]int64 {
	counts := make(map[int64]map[string]int64)
	for _, row := range buckets {
		if counts[row.SiteID] == nil {
			counts[row.SiteID] = map[string]int64{}
		}
		counts[row.SiteID][bucketKey(row.Bucket, unit)] += row.Count
	}
	return counts
}
