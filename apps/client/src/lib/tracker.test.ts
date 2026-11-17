import { describe, expect, it } from 'bun:test';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('../../static/s.js', import.meta.url), 'utf8');

type Run = {
	sent: string[];
	navigate: (path: string) => void;
	replace: (path: string) => void;
	track: (name: string) => void;
	fire: (type: string) => void;
};

/**
 * Runs the served tracker against a fake page and returns every beacon URL it built.
 * Timers are inert, so only the pageview, SPA navigations and custom events show up.
 */
function runTracker(href: string, opts: { mask?: boolean; exclude?: string; referrer?: string; webdriver?: boolean } = {}): Run {
	const sent: string[] = [];
	const location = new URL(href);
	const store = new Map<string, string>([['_vs_id', 'visitor']]);
	const listeners = new Map<string, Set<() => void>>();
	const window: Record<string, unknown> = {
		innerWidth: 1280,
		outerWidth: 1280,
		outerHeight: 800,
		addEventListener(type: string, fn: () => void) {
			listeners.set(type, (listeners.get(type) ?? new Set()).add(fn));
		},
		removeEventListener(type: string, fn: () => void) {
			listeners.get(type)?.delete(fn);
		}
	};
	const history = {
		pushState(_state: unknown, _title: string, path: string) {
			location.pathname = path;
		},
		replaceState(_state: unknown, _title: string, path: string) {
			location.pathname = path;
		}
	};
	const document = {
		currentScript: {
			src: 'https://vision.test/s.js',
			hasAttribute: (name: string) => name === 'data-mask' && !!opts.mask,
			getAttribute: (name: string) => (name === 'data-exclude' ? (opts.exclude ?? null) : null)
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
		{ language: 'en', webdriver: opts.webdriver ?? false },
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
		replace: (path) => history.replaceState(null, '', path),
		track: (name) => vision?.track(name),
		fire: (type) => [...(listeners.get(type) ?? [])].forEach((fn) => fn())
	};
}

/** The JSON payload of a beacon URL. */
function payload(url: string): Record<string, unknown> {
	return JSON.parse(decodeURIComponent(new URL(url).searchParams.get('data') ?? '{}'));
}

const pageviews = (run: Run) => run.sent.filter((u) => u.includes('/api/e/p?')).map(payload);

const beacons = (run: Run, kind: 'p' | 'h' | 't') => run.sent.filter((u) => u.includes(`/api/e/${kind}?`));

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
			b: 0,
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

	it('sends nothing for paths under a data-exclude prefix, on load and on navigation', () => {
		const run = runTracker('https://field.test/room/abc1234', { exclude: ' /room/ ,/share/' });
		run.track('joined');
		expect(run.sent).toEqual([]);
		run.navigate('/room/xyz9876');
		run.track('joined');
		expect(run.sent).toEqual([]);
		run.navigate('/dashboard');
		expect(pageviews(run).map((p) => p.path)).toEqual(['/dashboard']);
		run.replace('/share/token');
		run.track('opened');
		expect(beacons(run, 't')).toEqual([]);
		run.replace('/rooms');
		run.track('listed');
		expect(beacons(run, 't').map((u) => payload(u).path)).toEqual(['/rooms']);
	});

	it('still sends the first pageview and heartbeat outside excluded prefixes', () => {
		const run = runTracker('https://field.test/', { exclude: '/room/' });
		expect(beacons(run, 'p')).toHaveLength(1);
		expect(beacons(run, 'h')).toHaveLength(1);
	});

	it('flags a webdriver browser in every payload', () => {
		const run = runTracker('https://shop.test/', { webdriver: true });
		run.track('signup');
		expect(pageviews(run)[0].b).toBe(1);
		expect(payload(beacons(run, 'h')[0]).b).toBe(1);
		expect(payload(beacons(run, 't')[0]).b).toBe(1);
		expect(pageviews(runTracker('https://shop.test/'))[0].b).toBe(0);
	});

	it('sends one interacted heartbeat on the first gesture only', () => {
		const run = runTracker('https://shop.test/');
		const interacted = () => beacons(run, 'h').map(payload).filter((p) => p.interacted === true);
		run.fire('scroll');
		expect(interacted()).toHaveLength(0);
		run.fire('pointerdown');
		expect(interacted()).toEqual([{ hostname: 'shop.test', visitor_id: 'visitor', b: 0, interacted: true }]);
		run.fire('keydown');
		run.fire('pointerdown');
		expect(interacted()).toHaveLength(1);
	});

	it('sends no interacted heartbeat on an excluded path', () => {
		const run = runTracker('https://field.test/room/abc1234', { exclude: '/room/' });
		run.fire('pointerdown');
		expect(run.sent).toEqual([]);
	});
});
