import { describe, expect, it } from 'bun:test';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('../../static/s.js', import.meta.url), 'utf8');

type Run = { sent: string[]; navigate: (path: string) => void; track: (name: string) => void };

/**
 * Runs the served tracker against a fake page and returns every beacon URL it built.
 * Timers are inert, so only the pageview, SPA navigations and custom events show up.
 */
function runTracker(href: string, opts: { mask?: boolean; referrer?: string } = {}): Run {
	const sent: string[] = [];
	const location = new URL(href);
	const store = new Map<string, string>([['_vs_id', 'visitor']]);
	const window: Record<string, unknown> = { innerWidth: 1280, addEventListener() {} };
	const history = {
		pushState(_state: unknown, _title: string, path: string) {
			location.pathname = path;
		}
	};
	const document = {
		currentScript: {
			src: 'https://vision.test/s.js',
			hasAttribute: (name: string) => name === 'data-mask' && !!opts.mask
		},
		referrer: opts.referrer ?? '',
		readyState: 'loading',
		hidden: false,
		addEventListener() {}
	};
	class Image {
		set src(value: string) {
			sent.push(value);
		}
	}
	const run = new Function(
		'location',
		'document',
		'window',
		'history',
		'navigator',
		'localStorage',
		'Image',
		'setInterval',
		'setTimeout',
		'performance',
		source
	);
	run(
		location,
		document,
		window,
		history,
		{ language: 'en' },
		{ getItem: (k: string) => store.get(k) ?? null, setItem: (k: string, v: string) => store.set(k, v) },
		Image,
		() => 0,
		() => 0,
		{ getEntriesByType: () => [] }
	);
	const vision = window.vision as { track: (name: string) => void } | undefined;
	return {
		sent,
		navigate: (path) => history.pushState(null, '', path),
		track: (name) => vision?.track(name)
	};
}

/** The JSON payload of a beacon URL. */
function payload(url: string): Record<string, unknown> {
	return JSON.parse(decodeURIComponent(new URL(url).searchParams.get('data') ?? '{}'));
}

const pageviews = (run: Run) => run.sent.filter((u) => u.includes('/api/e/p?')).map(payload);

describe('s.js', () => {
	it('stays silent on local and non-web pages', () => {
		for (const href of ['http://localhost:5173/a', 'http://127.0.0.1/a', 'http://app.localhost/a', 'file:///tmp/a.html']) {
			expect(runTracker(href).sent).toEqual([]);
		}
	});

	it('sends paths and referrers untouched without data-mask', () => {
		const run = runTracker('https://shop.test/orders/12345/a@b.c', {
			referrer: 'https://shop.test/users/42?token=x#y'
		});
		run.navigate('/u/550e8400-e29b-41d4-a716-446655440000');
		const [first, second] = pageviews(run);
		expect(first).toEqual({
			hostname: 'shop.test',
			path: '/orders/12345/a@b.c',
			referrer: 'https://shop.test/users/42?token=x#y',
			language: 'en',
			screen_width: 1280,
			visitor_id: 'visitor',
			utm_source: '',
			utm_medium: '',
			utm_campaign: '',
			utm_term: '',
			utm_content: ''
		});
		expect(second.path).toBe('/u/550e8400-e29b-41d4-a716-446655440000');
	});

	it('masks identifier segments in every path with data-mask', () => {
		const run = runTracker('https://shop.test/orders/12345/a%40b.c/settings', {
			mask: true,
			referrer: 'https://shop.test/users/42/reset/abcdefgh12345678?token=x#y'
		});
		run.navigate('/u/550E8400-E29B-41D4-A716-446655440000/profile');
		run.navigate('/blog/a-long-article-title-without-digits');
		run.track('signup');
		const [first, second, third] = pageviews(run);
		expect(first.path).toBe('/orders/:id/:id/settings');
		expect(first.referrer).toBe('https://shop.test/users/:id/reset/:id');
		expect(second.path).toBe('/u/:id/profile');
		expect(third.path).toBe('/blog/a-long-article-title-without-digits');
		const event = run.sent.find((u) => u.includes('/api/e/t?'));
		expect(payload(event ?? '').path).toBe('/blog/a-long-article-title-without-digits');
	});

	it('reduces a cross-origin referrer to its origin with data-mask', () => {
		const run = runTracker('https://shop.test/', {
			mask: true,
			referrer: 'https://search.test/q/someone@mail.test?x=1'
		});
		expect(pageviews(run)[0].referrer).toBe('https://search.test');
	});
});
