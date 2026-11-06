import { expect, test } from 'bun:test';
import { handle } from './hooks.server';

test('proxied responses drop upstream compression and framing headers', async () => {
	const originalFetch = globalThis.fetch;
	globalThis.fetch = (async () =>
		new Response('{"ok":true}', {
			status: 200,
			headers: {
				'content-type': 'application/json',
				'content-encoding': 'gzip',
				'content-length': '9999',
				'transfer-encoding': 'chunked',
				'x-upstream': 'kept'
			}
		})) as typeof fetch;

	try {
		const res = await handle({
			event: {
				url: new URL('http://localhost:3000/api/sites'),
				request: new Request('http://localhost:3000/api/sites')
			},
			resolve: async () => new Response(null)
		} as never);

		expect(res.headers.get('content-encoding')).toBeNull();
		expect(res.headers.get('content-length')).toBeNull();
		expect(res.headers.get('transfer-encoding')).toBeNull();
		expect(res.headers.get('content-type')).toBe('application/json');
		expect(res.headers.get('x-upstream')).toBe('kept');
		expect(await res.text()).toBe('{"ok":true}');
	} finally {
		globalThis.fetch = originalFetch;
	}
});

test('the tracker endpoint stays readable cross-origin whatever headers the API adds', async () => {
	const originalFetch = globalThis.fetch;
	globalThis.fetch = (async () =>
		new Response('GIF89a', {
			status: 200,
			headers: {
				'content-type': 'image/gif',
				'content-security-policy': "frame-ancestors 'none'",
				'cross-origin-opener-policy': 'same-origin',
				'x-frame-options': 'DENY'
			}
		})) as typeof fetch;

	try {
		const event = (method: string) => ({
			url: new URL('http://localhost:3000/api/e/p?data=%7B%7D'),
			request: new Request('http://localhost:3000/api/e/p?data=%7B%7D', { method })
		});
		const pixel = await handle({ event: event('GET'), resolve: async () => new Response(null) } as never);
		expect(pixel.headers.get('access-control-allow-origin')).toBe('*');
		expect(pixel.headers.get('content-security-policy')).toBeNull();
		expect(pixel.headers.get('x-frame-options')).toBeNull();

		const preflight = await handle({ event: event('OPTIONS'), resolve: async () => new Response(null) } as never);
		expect(preflight.status).toBe(204);
		expect(preflight.headers.get('access-control-allow-origin')).toBe('*');
		expect(preflight.headers.get('access-control-allow-methods')).toContain('POST');
	} finally {
		globalThis.fetch = originalFetch;
	}
});
