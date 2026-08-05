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
