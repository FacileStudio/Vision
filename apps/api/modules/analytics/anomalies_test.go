package analytics

import (
	"log/slog"
	"testing"
)

func repeat(value float64, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = value
	}
	return out
}

func TestEvaluate(t *testing.T) {
	cases := []struct {
		name      string
		metric    string
		value     float64
		history   []float64
		want      bool
		direction string
	}{
		{"flat series stays quiet", MetricVisitors, 102, repeat(100, 8), false, ""},
		{"noise inside the usual range", MetricVisitors, 104, []float64{95, 103, 98, 101, 99, 105, 97, 100}, false, ""},
		{"spike", MetricVisitors, 300, []float64{95, 103, 98, 101, 99, 105, 97, 100}, true, DirectionUp},
		{"drop", MetricPageviews, 30, []float64{210, 190, 200, 205, 195, 200, 198, 202}, true, DirectionDown},
		{"outage to zero", MetricPageviews, 0, []float64{210, 190, 200, 205, 195, 200, 198, 202}, true, DirectionDown},
		{"a normal weekend value is not a drop", MetricVisitors, 21, []float64{20, 22, 19, 21, 20, 23, 18, 20}, false, ""},
		{"a weekday volume on a weekend is a spike", MetricVisitors, 100, []float64{20, 22, 19, 21, 20, 23, 18, 20}, true, DirectionUp},
		{"low volume never flags", MetricVisitors, 12, []float64{2, 3, 2, 4, 3, 2, 3, 2}, false, ""},
		{"low volume gate looks at the larger side", MetricVisitors, 5, []float64{80, 90, 85, 88, 82, 86, 84, 87}, true, DirectionDown},
		{"too few weeks", MetricVisitors, 500, []float64{100, 100, 100}, false, ""},
		{"zeros from missing days count as data", MetricVisitors, 100, []float64{100, 0, 100, 100, 0, 100, 100, 100}, false, ""},
		{"flat history under the relative gate", MetricVisitors, 130, repeat(100, 8), false, ""},
		{"flat history past the relative gate", MetricVisitors, 150, repeat(100, 8), true, DirectionUp},
		{"referrer share jump", MetricTopReferrerShare, 0.6, []float64{0.22, 0.25, 0.2, 0.24, 0.23, 0.21, 0.26, 0.22}, true, DirectionUp},
		{"referrer share wobble", MetricTopReferrerShare, 0.27, []float64{0.22, 0.25, 0.2, 0.24, 0.23, 0.21, 0.26, 0.22}, false, ""},
		{"goal conversions drop", MetricGoalConversions, 1, []float64{12, 10, 11, 13, 12, 11, 10, 12}, true, DirectionDown},
		{"goal conversions tiny numbers", MetricGoalConversions, 3, []float64{0, 1, 0, 1, 0, 0, 1, 0}, false, ""},
		{"unknown metric", "bounce_rate", 1, repeat(1, 8), false, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := evaluate(c.metric, c.value, c.history)
			if ok != c.want {
				t.Fatalf("flagged = %v, want %v (%+v)", ok, c.want, got)
			}
			if ok && got.Direction != c.direction {
				t.Fatalf("direction %s, want %s", got.Direction, c.direction)
			}
			if ok && got.Weeks != len(c.history) {
				t.Fatalf("weeks %d, want %d", got.Weeks, len(c.history))
			}
		})
	}
}

func TestMedian(t *testing.T) {
	cases := []struct {
		values []float64
		want   float64
	}{{nil, 0}, {[]float64{5}, 5}, {[]float64{3, 1, 2}, 2}, {[]float64{4, 1, 3, 2}, 2.5}}
	for _, c := range cases {
		original := append([]float64(nil), c.values...)
		if got := median(c.values); got != c.want {
			t.Errorf("median(%v) = %v, want %v", original, got, c.want)
		}
		for i := range original {
			if c.values[i] != original[i] {
				t.Errorf("median reordered its input %v", original)
			}
		}
	}
}

func TestRankContributionsKeepsTheBiggestMoves(t *testing.T) {
	today := map[string]int64{"/a": 100, "/b": 10, "/new": 40}
	before := map[string]int64{"/a": 400, "/b": 80, "/gone": 160}
	got := rankContributions(today, before, 4)
	if len(got) != 4 || got[0].Value != "/gone" || got[0].Delta != -40 || got[1].Value != "/new" || got[1].Delta != 40 {
		t.Fatalf("want /gone and /new first by absolute change, got %+v", got)
	}
	if got[3].Value != "/a" || got[3].Baseline != 100 || got[3].Delta != 0 {
		t.Fatalf("baseline is the average of the weeks, got %+v", got[3])
	}
}

func TestClipKeepsValuesShortAndValid(t *testing.T) {
	long := make([]rune, 500)
	for i := range long {
		long[i] = 'é'
	}
	if got := clip(string(long)); len([]rune(got)) != valueRunes {
		t.Fatalf("a long value must be cut to %d runes, got %d", valueRunes, len([]rune(got)))
	}
	if got := clip("ok\xff/path"); got != "ok/path" {
		t.Fatalf("invalid UTF-8 must be dropped, got %q", got)
	}
}

func discardLogger() *slog.Logger { return slog.New(slog.DiscardHandler) }
