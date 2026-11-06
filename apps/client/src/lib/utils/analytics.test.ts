import { describe, expect, it } from 'bun:test';
import { countryName, fmt, formatChartDate, mergeByLabel, trendPercent } from './analytics';

describe('mergeByLabel', () => {
	it('sums rows that share a label', () => {
		const merged = mergeByLabel([
			{ label: '/pricing', count: 3 },
			{ label: '/', count: 10 },
			{ label: '/pricing', count: 4 }
		]);

		expect(merged).toEqual([
			{ label: '/', count: 10 },
			{ label: '/pricing', count: 7 }
		]);
	});

	it('re-ranks on the merged total, not on the largest fragment', () => {
		// Split across two rows, "france" outranks "spain" only once merged.
		const merged = mergeByLabel([
			{ label: 'spain', count: 9 },
			{ label: 'france', count: 6 },
			{ label: 'france', count: 5 }
		]);

		expect(merged.map((r) => r.label)).toEqual(['france', 'spain']);
	});

	it('keeps the empty label as its own bucket', () => {
		// "no referrer" is a real answer; dropping it makes the numbers stop adding up.
		const merged = mergeByLabel([
			{ label: '', count: 2 },
			{ label: '', count: 3 }
		]);

		expect(merged).toEqual([{ label: '', count: 5 }]);
	});

	it('does not mutate the rows it was given', () => {
		const rows = [
			{ label: 'a', count: 1 },
			{ label: 'a', count: 1 }
		];
		mergeByLabel(rows);

		expect(rows[0].count).toBe(1);
	});
});

describe('locale formatting', () => {
	it('groups numbers the way each locale does', () => {
		expect(fmt(1234567, 'en')).toBe('1,234,567');
		expect(fmt(1234567, 'de')).toBe('1.234.567');
	});

	it('names a country in the locale', () => {
		expect(countryName('DE', 'en')).toBe('Germany');
		expect(countryName('DE', 'fr')).toBe('Allemagne');
		expect(countryName('de', 'es')).toBe('Alemania');
	});

	it('labels a week with the locale prefix', () => {
		expect(formatChartDate('2026-12', 'week', 'de')).toBe('KW12');
		expect(formatChartDate('2026-12', 'week', 'en')).toBe('W12');
	});

	it('writes the trend with the locale decimal separator', () => {
		expect(trendPercent(112.5, 100, 'fr').text).toContain('12,5');
	});
});
