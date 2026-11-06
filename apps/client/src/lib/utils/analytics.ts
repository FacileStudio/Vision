export type RangeKey = 'today' | '7d' | '30d' | '90d' | 'this_year' | 'custom';

export type Granularity = 'hour' | 'day' | 'week' | 'month';

export const rangeKeys: RangeKey[] = ['today', '7d', '30d', '90d', 'this_year', 'custom'];

export const granularityKeys: Granularity[] = ['hour', 'day', 'week', 'month'];

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

export function formatDateRange(from: string, to: string, locale: string): string {
	const withYear = new Intl.DateTimeFormat(locale, { month: 'short', day: 'numeric', year: 'numeric' });
	const withoutYear = new Intl.DateTimeFormat(locale, { month: 'short', day: 'numeric' });
	const f = new Date(from + 'T00:00:00');
	const t = new Date(to + 'T00:00:00');
	if (from === to) return withYear.format(f);
	return `${withoutYear.format(f)} – ${withYear.format(t)}`;
}

export function trendPercent(current: number, prev: number, locale: string): { text: string; color: string } {
	if (prev === 0) return { text: '—', color: 'text-fc-fg-muted' };
	const pct = ((current - prev) / prev) * 100;
	const one = new Intl.NumberFormat(locale, { minimumFractionDigits: 1, maximumFractionDigits: 1 });
	if (pct > 0) return { text: `↑ ${one.format(pct)}%`, color: 'text-fc-success' };
	if (pct < 0) return { text: `↓ ${one.format(Math.abs(pct))}%`, color: 'text-fc-danger' };
	return { text: '—', color: 'text-fc-fg-muted' };
}

export function fmt(n: number, locale: string): string {
	return new Intl.NumberFormat(locale).format(n);
}

/** One decimal place, in the locale's own separators. */
export function decimal(n: number, locale: string): string {
	return new Intl.NumberFormat(locale, { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format(n);
}

/** An hour of the day, 0 to 23, written the way the locale writes a clock hour. */
export function formatHour(hour: number, locale: string): string {
	return new Intl.DateTimeFormat(locale, { hour: 'numeric' }).format(new Date(2000, 0, 1, hour));
}

/** A region code as a country name, falling back to the code itself. */
export function countryName(code: string, locale: string): string {
	try {
		return new Intl.DisplayNames([locale], { type: 'region' }).of(code.toUpperCase()) ?? code.toUpperCase();
	} catch {
		return code.toUpperCase();
	}
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

export function allowedGranularities(range: RangeKey): Granularity[] {
	const allowed = ALLOWED_GRANULARITIES[range];
	return granularityKeys.filter((key) => allowed.includes(key));
}

export function defaultGranularity(range: RangeKey): Granularity {
	return DEFAULT_GRANULARITY[range];
}

const WEEK_PREFIX: Record<string, string> = { fr: 'S', es: 'S', de: 'KW' };

export function formatChartDate(d: string, granularity: Granularity, locale: string): string {
	if (granularity === 'hour') {
		const parts = d.split(' ');
		return parts[1] ?? d;
	}
	if (granularity === 'week') {
		return `${WEEK_PREFIX[locale] ?? 'W'}${d.split('-').pop()}`;
	}
	if (granularity === 'month') {
		const date = new Date(d + '-01T00:00:00');
		return new Intl.DateTimeFormat(locale, { month: 'short' }).format(date);
	}
	const date = new Date(d + 'T00:00:00');
	return new Intl.DateTimeFormat(locale, { month: 'numeric', day: 'numeric' }).format(date);
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
