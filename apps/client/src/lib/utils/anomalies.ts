import type { Anomaly, Breakdown, Contribution } from '../bulle';

export type Dimension = keyof Breakdown;

export const DIMENSIONS: readonly Dimension[] = ['referrers', 'paths', 'countries', 'devices'];

/** A number the way the metric reads: a share as a percentage, everything else as a count. */
export function formatMetric(metric: Anomaly['metric'], value: number, locale: string): string {
	if (metric === 'top_referrer_share') {
		return new Intl.NumberFormat(locale, { style: 'percent', maximumFractionDigits: 0 }).format(value);
	}
	return new Intl.NumberFormat(locale, { maximumFractionDigits: value < 10 && value % 1 !== 0 ? 1 : 0 }).format(value);
}

/** The change against the baseline as a signed percentage, or empty when there is no baseline to divide by. */
export function formatChange(anomaly: Pick<Anomaly, 'change_pct'>, locale: string): string {
	if (anomaly.change_pct === null || anomaly.change_pct === undefined) return '';
	const formatter = new Intl.NumberFormat(locale, { style: 'percent', maximumFractionDigits: 0, signDisplay: 'exceptZero' });
	return formatter.format(anomaly.change_pct / 100);
}

/** A UTC day, such as 2026-11-15, as a short date in the locale. */
export function formatDay(day: string, locale: string): string {
	const parsed = new Date(`${day}T00:00:00Z`);
	if (Number.isNaN(parsed.getTime())) return day;
	return new Intl.DateTimeFormat(locale, { weekday: 'short', day: 'numeric', month: 'short', timeZone: 'UTC' }).format(parsed);
}

/** The values of one dimension that moved, biggest absolute change first, at most limit. */
export function movers(anomaly: Pick<Anomaly, 'breakdown'>, dimension: Dimension, limit = 3): Contribution[] {
	const rows = anomaly.breakdown?.[dimension] ?? [];
	return rows
		.filter((row) => row.delta !== 0)
		.sort((a, b) => Math.abs(b.delta) - Math.abs(a.delta))
		.slice(0, limit);
}

/** A signed whole number such as +200 or -40. */
export function signed(value: number, locale: string): string {
	return new Intl.NumberFormat(locale, { maximumFractionDigits: 0, signDisplay: 'exceptZero' }).format(value);
}

/** Where an anomaly's day sits among the chart's buckets: its own day, or its month. Hour and week buckets get no marker. */
export function anomalyIndex(day: string, buckets: string[], granularity: string): number {
	if (granularity === 'day') return buckets.indexOf(day);
	if (granularity === 'month') return buckets.indexOf(day.slice(0, 7));
	return -1;
}

/** The anomalies that land on the traffic chart, as LineChart markers: visitors on the visitors line, the rest on pageviews. */
export function anomalyMarkers(
	anomalies: Anomaly[],
	buckets: string[],
	granularity: string,
	label: (anomaly: Anomaly) => string
): { index: number; label: string; tone: 'info' | 'warning'; series: number; id: number }[] {
	return anomalies.flatMap((anomaly) => {
		const index = anomalyIndex(anomaly.day, buckets, granularity);
		if (index < 0) return [];
		return [
			{
				index,
				label: label(anomaly),
				tone: anomaly.direction === 'up' ? ('info' as const) : ('warning' as const),
				series: anomaly.metric === 'visitors' ? 1 : 0,
				id: anomaly.id
			}
		];
	});
}
