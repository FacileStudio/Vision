package analytics

import (
	"math"
	"slices"
)

// Metric names stored on an anomaly.
const (
	MetricVisitors         = "visitors"
	MetricPageviews        = "pageviews"
	MetricGoalConversions  = "goal_conversions"
	MetricTopReferrerShare = "top_referrer_share"
)

// Directions of a deviation from the baseline.
const (
	DirectionUp   = "up"
	DirectionDown = "down"
)

const (
	baselineWeeks  = 8
	minWeeks       = 4
	madToSigma     = 1.4826
	minScore       = 3.5
	minRelative    = 0.3
	shareMinVolume = 100
)

// rule holds the volume and size gates of one metric. A day is a candidate only when the
// larger of its value and its baseline reaches MinVolume and the two differ by at least
// MinDelta. The spread is never smaller than MinScale or RelScale of the baseline, so a
// perfectly flat history does not turn a small wobble into a huge score.
type rule struct {
	MinVolume float64
	MinDelta  float64
	MinScale  float64
	RelScale  float64
}

var rules = map[string]rule{
	MetricVisitors:         {MinVolume: 30, MinDelta: 15, MinScale: 1, RelScale: 0.1},
	MetricPageviews:        {MinVolume: 50, MinDelta: 25, MinScale: 1, RelScale: 0.1},
	MetricGoalConversions:  {MinVolume: 5, MinDelta: 3, MinScale: 1, RelScale: 0.1},
	MetricTopReferrerShare: {MinVolume: 0, MinDelta: 0.1, MinScale: 0.02, RelScale: 0.1},
}

// finding is one metric's deviation on one day.
type finding struct {
	Direction string
	Value     float64
	Baseline  float64
	MAD       float64
	Score     float64
	ChangePct *float64
	Weeks     int
}

// evaluate compares value with the same weekday of the previous weeks, oldest or newest
// first. It reports false with fewer than minWeeks of history, below the metric's volume
// or size gates, or when the robust z-score stays under minScore or the relative change
// under minRelative.
func evaluate(metric string, value float64, history []float64) (finding, bool) {
	gate, known := rules[metric]
	if !known || len(history) < minWeeks {
		return finding{}, false
	}
	baseline := median(history)
	deviations := make([]float64, len(history))
	for i, past := range history {
		deviations[i] = math.Abs(past - baseline)
	}
	mad := median(deviations)
	delta := value - baseline
	if math.Max(value, baseline) < gate.MinVolume || math.Abs(delta) < gate.MinDelta {
		return finding{}, false
	}
	scale := math.Max(madToSigma*mad, math.Max(gate.MinScale, gate.RelScale*math.Abs(baseline)))
	score := delta / scale
	if math.Abs(score) < minScore {
		return finding{}, false
	}
	out := finding{Direction: DirectionUp, Value: value, Baseline: baseline, MAD: mad, Score: score, Weeks: len(history)}
	if delta < 0 {
		out.Direction = DirectionDown
	}
	if baseline != 0 {
		change := delta / baseline * 100
		out.ChangePct = &change
		if math.Abs(delta/baseline) < minRelative {
			return finding{}, false
		}
	}
	return out, true
}

// median is the middle of values, the mean of the two middle ones for an even count, and
// 0 for none. It leaves values as they were.
func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := slices.Clone(values)
	slices.Sort(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / 2
}
