import { describe, expect, it } from 'bun:test';
import type { Anomaly } from '../bulle';
import { formatChange, formatDay, formatMetric, movers, signed } from './anomalies';

const rows = (deltas: number[]) => deltas.map((delta, i) => ({ value: `v${i}`, count: 10, baseline: 0, delta }));

describe('formatMetric', () => {
	it('shows a share as a percentage and a count as a number', () => {
		expect(formatMetric('top_referrer_share', 0.62, 'en')).toBe('62%');
		expect(formatMetric('visitors', 1240, 'en')).toBe('1,240');
		expect(formatMetric('goal_conversions', 2.5, 'en')).toBe('2.5');
	});
});

describe('formatChange', () => {
	it('signs the change and stays empty without a baseline', () => {
		expect(formatChange({ change_pct: 500 }, 'en')).toBe('+500%');
		expect(formatChange({ change_pct: -40 }, 'en')).toBe('-40%');
		expect(formatChange({ change_pct: null }, 'en')).toBe('');
	});
});

describe('formatDay', () => {
	it('reads the day in UTC whatever the machine zone is', () => {
		expect(formatDay('2026-11-15', 'en')).toContain('15');
		expect(formatDay('not a day', 'en')).toBe('not a day');
	});
});

describe('movers', () => {
	it('ranks by absolute change, drops zero moves and caps the list', () => {
		const anomaly = { breakdown: { referrers: rows([5, -40, 0, 20, 3]), paths: [], countries: [], devices: [] } } as Pick<Anomaly, 'breakdown'>;
		expect(movers(anomaly, 'referrers', 3).map((row) => row.delta)).toEqual([-40, 20, 5]);
		expect(movers(anomaly, 'paths')).toEqual([]);
	});

	it('keeps visitor-written text untouched, for the page to print as text', () => {
		const hostile = '<img src=x onerror=alert(1)>';
		const anomaly = { breakdown: { referrers: [{ value: hostile, count: 9, baseline: 0, delta: 9 }], paths: [], countries: [], devices: [] } } as Pick<Anomaly, 'breakdown'>;
		expect(movers(anomaly, 'referrers')[0]!.value).toBe(hostile);
	});
});

describe('signed', () => {
	it('prefixes a plus on gains', () => {
		expect(signed(200, 'en')).toBe('+200');
		expect(signed(-40, 'en')).toBe('-40');
		expect(signed(0, 'en')).toBe('0');
	});
});
