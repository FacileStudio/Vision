import { describe, expect, it } from 'bun:test';
import { askPrompt, pageContext, runActive, safeHref, upgradeLink } from './bulle';

describe('safeHref', () => {
	it('keeps http and https links and drops everything else', () => {
		expect(safeHref('https://bulle.facile.studio/s/1')).toBe('https://bulle.facile.studio/s/1');
		expect(safeHref('javascript:alert(1)')).toBeUndefined();
		expect(safeHref('data:text/html,<b>x</b>')).toBeUndefined();
		expect(safeHref('//evil.test')).toBeUndefined();
		expect(safeHref('')).toBeUndefined();
	});
});

describe('upgradeLink', () => {
	it('only sends people to https addresses', () => {
		expect(upgradeLink('https://sso.facile.studio/spaces')).toBe('https://sso.facile.studio/spaces');
		expect(upgradeLink('javascript:alert(1)')).toBe('https://sso.facile.studio/spaces');
		expect(upgradeLink(undefined)).toBe('https://sso.facile.studio/spaces');
	});
});

describe('askPrompt', () => {
	it('names the four read tools, asks for periods and honesty, and ends with the question', () => {
		const prompt = askPrompt(7, '  Why did visitors drop?  ', 'fr');
		for (const tool of ['vision_get_stats', 'vision_compare_periods', 'vision_list_anomalies', 'vision_list_goals']) {
			expect(prompt).toContain(tool);
		}
		expect(prompt).toContain('site_id 7');
		expect(prompt).toContain('period');
		expect(prompt).toContain('say so');
		expect(prompt).toContain('French');
		expect(prompt.endsWith('Question: Why did visitors drop?')).toBe(true);
	});

	it('stays inside Bulle’s 2000 character cap for a 500 character question', () => {
		expect(askPrompt(1, 'x'.repeat(500), 'en').length).toBeLessThan(2000);
	});
});

describe('pageContext', () => {
	it('is sent for https pages only', () => {
		expect(pageContext(3, 'Shop', 'https://vision.facile.studio/sites/3')).toEqual({
			app: 'vision',
			id: '3',
			title: 'Shop',
			url: 'https://vision.facile.studio/sites/3'
		});
		expect(pageContext(3, 'Shop', 'http://localhost:5173/sites/3')).toBeUndefined();
	});
});

describe('runActive', () => {
	it('polls until the run is done or failed', () => {
		const run = (state: 'running' | 'needs_approval' | 'done' | 'failed') => ({ id: 'r', session_id: 's', state, summary: '', continue_url: '' });
		expect(runActive(null)).toBe(true);
		expect(runActive(run('running'))).toBe(true);
		expect(runActive(run('needs_approval'))).toBe(true);
		expect(runActive(run('done'))).toBe(false);
		expect(runActive(run('failed'))).toBe(false);
	});
});
