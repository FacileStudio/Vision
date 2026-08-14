import { afterEach, expect, test } from 'bun:test';
import { currentUser } from './backend';

const originalFetch = globalThis.fetch;

function respondWith(status: number, body: unknown) {
	const calls: Request[] = [];
	globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		calls.push(new Request(typeof input === 'string' ? `http://localhost${input}` : input, init));
		return new Response(JSON.stringify(body), {
			status,
			headers: { 'content-type': 'application/json' }
		});
	}) as typeof fetch;
	return calls;
}

afterEach(() => {
	globalThis.fetch = originalFetch;
});

test('currentUser reports the signed-in user from a session the client cannot read', async () => {
	respondWith(200, { id: '1', email: 'a@b.c', name: 'A' });
	expect((await currentUser())?.email).toBe('a@b.c');
});

test('currentUser reports no session on 401 rather than throwing', async () => {
	respondWith(401, { error: { message: 'missing auth' } });
	expect(await currentUser()).toBeNull();
});

test('currentUser rethrows a server failure, so a bad round-trip is not a logout', async () => {
	respondWith(500, { error: { message: 'boom' } });
	expect(currentUser()).rejects.toThrow('boom');
});

test('every request carries the CSRF header porte demands of a cookie session', async () => {
	const calls = respondWith(200, { id: '1', email: 'a@b.c', name: 'A' });
	await currentUser();
	expect(calls[0].headers.get('X-Facile-CSRF')).toBe('1');
});
