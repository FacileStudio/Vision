export type RangeKey = 'today' | '7d' | '30d' | '90d' | 'this_year' | 'custom';

export type Granularity = 'hour' | 'day' | 'week' | 'month';

export const ranges: { key: RangeKey; label: string }[] = [
	{ key: 'today', label: 'Today' },
	{ key: '7d', label: '7d' },
	{ key: '30d', label: '30d' },
	{ key: '90d', label: '90d' },
	{ key: 'this_year', label: 'Year' },
	{ key: 'custom', label: 'Custom' }
];

export const granularities: { key: Granularity; label: string }[] = [
	{ key: 'hour', label: 'Hourly' },
	{ key: 'day', label: 'Daily' },
	{ key: 'week', label: 'Weekly' },
	{ key: 'month', label: 'Monthly' }
];

export function rangeDates(
	key: RangeKey,
	customFrom?: string,
	customTo?: string
): { from: string; to: string } {
	const now = new Date();
	const toStr = now.toISOString().slice(0, 10);

	if (key === 'today') return { from: toStr, to: toStr };
	if (key === 'this_year') return { from: `${now.getFullYear()}-01-01`, to: toStr };
	if (key === 'custom') return { from: customFrom || toStr, to: customTo || toStr };

	const daysMap: Record<string, number> = { '7d': 7, '30d': 30, '90d': 90 };
	const from = new Date(now);
	from.setDate(from.getDate() - (daysMap[key] ?? 30));
	return { from: from.toISOString().slice(0, 10), to: toStr };
}

export function formatDateRange(from: string, to: string): string {
	const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric', year: 'numeric' };
	const f = new Date(from + 'T00:00:00');
	const t = new Date(to + 'T00:00:00');
	if (from === to) return f.toLocaleDateString('en-US', opts);
	return `${f.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })} – ${t.toLocaleDateString('en-US', opts)}`;
}

export function trendPercent(current: number, prev: number): { text: string; color: string } {
	if (prev === 0) return { text: '—', color: 'text-fc-fg-muted' };
	const pct = ((current - prev) / prev) * 100;
	if (pct > 0) return { text: `↑ ${pct.toFixed(1)}%`, color: 'text-fc-success' };
	if (pct < 0) return { text: `↓ ${Math.abs(pct).toFixed(1)}%`, color: 'text-fc-danger' };
	return { text: '—', color: 'text-fc-fg-muted' };
}

export function fmt(n: number): string {
	return n.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

const SEARCH_ENGINES = ['google', 'bing', 'duckduckgo', 'yahoo', 'baidu', 'yandex'];
const SOCIAL_NETWORKS = [
	'twitter',
	'x.com',
	'facebook',
	'instagram',
	'linkedin',
	'reddit',
	'tiktok',
	'youtube'
];

export function classifyReferrer(ref: string): 'search' | 'social' | 'other' {
	const lower = ref.toLowerCase();
	if (SEARCH_ENGINES.some((s) => lower.includes(s))) return 'search';
	if (SOCIAL_NETWORKS.some((s) => lower.includes(s))) return 'social';
	return 'other';
}

/*
 * Series colour comes from muse — six slots, assigned by index and never by rank, so a
 * filter that drops a series cannot repaint the survivors. Import `chartColor` straight
 * from `@facile/muse` at the call site: this module used to re-export it for convenience,
 * and nothing ever used the re-export, while it dragged muse's barrel — and therefore its
 * `$state` rune modules — into every plain-TS import of this file. That makes the module
 * unloadable outside the Svelte compiler, so it could not be unit-tested at all.
 */

const ALLOWED_GRANULARITIES: Record<RangeKey, Granularity[]> = {
	today: ['hour'],
	'7d': ['hour', 'day'],
	'30d': ['day', 'week'],
	'90d': ['day', 'week', 'month'],
	this_year: ['day', 'week', 'month'],
	custom: ['hour', 'day', 'week', 'month']
};

const DEFAULT_GRANULARITY: Record<RangeKey, Granularity> = {
	today: 'hour',
	'7d': 'day',
	'30d': 'day',
	'90d': 'day',
	this_year: 'week',
	custom: 'day'
};

export function allowedGranularities(range: RangeKey) {
	const allowed = ALLOWED_GRANULARITIES[range];
	return granularities.filter((g) => allowed.includes(g.key));
}

export function defaultGranularity(range: RangeKey): Granularity {
	return DEFAULT_GRANULARITY[range];
}

export function formatChartDate(d: string, granularity: Granularity): string {
	if (granularity === 'hour') {
		const parts = d.split(' ');
		return parts[1] ?? d;
	}
	if (granularity === 'week') {
		return `W${d.split('-').pop()}`;
	}
	if (granularity === 'month') {
		const date = new Date(d + '-01T00:00:00');
		return date.toLocaleDateString('en-US', { month: 'short' });
	}
	const date = new Date(d + 'T00:00:00');
	return `${date.getMonth() + 1}/${date.getDate()}`;
}

/*
 * Collapse rows that carry the same label, summing their counts.
 *
 * Two reasons, and the second one is a crash. A breakdown list showing "/pricing" twice is
 * meaningless to a reader — whatever the API grouped by, the label is all the UI shows, so
 * rows sharing one are one row. And every list here is rendered with `(item.label)` as the
 * `{#each}` key, so a repeat is a hard `each_key_duplicate` error that takes the whole page
 * down rather than rendering it slightly wrong.
 *
 * Merge before slicing a top-N, never after: summing the tail is the entire point, and a
 * post-slice merge would drop counts that belong in the total.
 *
 * Empty labels are kept and merged together rather than dropped — "(none)" is a real bucket
 * in every one of these dimensions (no referrer, unknown OS), and silently discarding it
 * makes the numbers stop adding up.
 */
export function mergeByLabel<T extends { label: string; count: number }>(rows: T[]): T[] {
	const seen = new Map<string, T>();
	for (const row of rows) {
		const found = seen.get(row.label);
		if (found) found.count += row.count;
		else seen.set(row.label, { ...row });
	}
	return [...seen.values()].sort((a, b) => b.count - a.count);
}
