import { afterEach, expect, mock, test } from 'bun:test';

const visits: string[] = [];
mock.module('$app/environment', () => ({ browser: true }));
mock.module('$app/navigation', () => ({
	goto: async (target: string) => {
		visits.push(target);
	}
}));
Object.assign(globalThis, { window: { location: { pathname: '/sites', search: '' } } });

const { api, currentUser, forgetStoredToken, logout } = await import('./backend');

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

function storage(entries: Record<string, string>) {
	const store = new Map(Object.entries(entries));
	Object.assign(globalThis, {
		localStorage: {
			getItem: (key: string) => store.get(key) ?? null,
			setItem: (key: string, value: string) => store.set(key, value),
			removeItem: (key: string) => store.delete(key)
		}
	});
	return store;
}

afterEach(() => {
	globalThis.fetch = originalFetch;
	visits.length = 0;
});

test('currentUser reports the signed-in user from a session the client cannot read', async () => {
	respondWith(200, { id: '1', email: 'a@b.c', name: 'A' });
	expect((await currentUser())?.email).toBe('a@b.c');
});

test('currentUser reports no session on 401 rather than throwing, and does not redirect', async () => {
	respondWith(401, { error: { message: 'missing auth' } });
	expect(await currentUser()).toBeNull();
	expect(visits).toEqual([]);
});

test('currentUser rethrows a server failure, so a bad round-trip is not a logout', async () => {
	respondWith(500, { error: { message: 'boom' } });
	expect(currentUser()).rejects.toThrow('boom');
});

test('a 401 on any other call sends the person to the login page', async () => {
	respondWith(401, { error: { message: 'missing auth' } });
	await expect(api.sites.list()).rejects.toThrow('missing auth');
	expect(visits).toEqual(['/login?next=%2Fsites']);
});

test('a mutation carries the CSRF header porte demands of a cookie session', async () => {
	const calls = respondWith(200, {});
	await api.workspaces.create('Acme');
	expect(calls[0].headers.get('X-Facile-CSRF')).toBe('1');
	expect(calls[0].headers.get('Content-Type')).toBe('application/json');
});

test('no request carries a bearer token, and logout deletes any an old login left', async () => {
	const store = storage({ 'vision.token': 't1', token: 't0', 'vision-theme': 'dark' });
	const calls = respondWith(200, { id: '1', email: 'a@b.c', name: 'A' });
	await currentUser();
	expect(calls[0].headers.get('Authorization')).toBeNull();
	await logout();
	expect([...store.keys()]).toEqual(['vision-theme']);
	store.set('token', 't2');
	forgetStoredToken();
	expect(store.has('token')).toBe(false);
});

test('invitations reach the user menu only with an href, and a failure is no row', async () => {
	respondWith(200, { count: 2, href: 'https://sso.facile.studio/spaces' });
	expect(await api.me.invitations()).toEqual({ count: 2, href: 'https://sso.facile.studio/spaces' });
	respondWith(200, { count: 0 });
	expect(await api.me.invitations()).toBeNull();
	respondWith(404, { error: { message: 'not found' } });
	expect(await api.me.invitations()).toBeNull();
	respondWith(401, { error: { message: 'missing auth' } });
	expect(await api.me.invitations()).toBeNull();
	expect(visits).toEqual([]);
});
