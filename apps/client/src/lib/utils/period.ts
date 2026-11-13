export type Period = '24h' | '7d' | '30d';

export const periods: Period[] = ['24h', '7d', '30d'];

const HOURS: Record<Period, number> = { '24h': 24, '7d': 24 * 7, '30d': 24 * 30 };

/** The period a page carries in `?period=`, `7d` when absent or unknown. */
export function periodFilter(url: URL): Period {
	const raw = url.searchParams.get('period');
	return periods.includes(raw as Period) ? (raw as Period) : '7d';
}

/** The rolling window that ends now, as RFC 3339 instants. */
export function periodWindow(period: Period, now = new Date()): { from: string; to: string } {
	const from = new Date(now.getTime() - HOURS[period] * 3_600_000);
	return { from: from.toISOString(), to: now.toISOString() };
}
