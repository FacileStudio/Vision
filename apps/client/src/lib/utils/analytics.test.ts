import { describe, expect, it } from 'bun:test';
import { mergeByLabel } from './analytics';

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
