const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:4000';

function getToken(): string | null {
	if (typeof window === 'undefined') return null;
	return localStorage.getItem('token');
}

export function setToken(token: string) {
	localStorage.setItem('token', token);
}

export function clearToken() {
	localStorage.removeItem('token');
}

export function isAuthenticated(): boolean {
	return getToken() !== null;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const headers: Record<string, string> = {
		'Content-Type': 'application/json'
	};

	const token = getToken();
	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	const res = await fetch(`${API_BASE}${path}`, {
		method,
		headers,
		body: body ? JSON.stringify(body) : undefined
	});

	if (res.status === 204) return undefined as T;

	const data = await res.json();
	if (!res.ok) {
		throw new Error(data?.error?.message ?? 'Request failed');
	}
	return data as T;
}

export const api = {
	auth: {
		register: (email: string, password: string) =>
			request<{ user_id: string; token: string }>('POST', '/auth/register', { email, password }),
		login: (email: string, password: string) =>
			request<{ user_id: string; token: string }>('POST', '/auth/login', { email, password })
	},
	sites: {
		list: () => request<Site[]>('GET', '/sites'),
		get: (id: number) => request<Site>('GET', `/sites/${id}`),
		create: (name: string, domain: string) => request<Site>('POST', '/sites', { name, domain }),
		update: (id: number, name: string, domain: string) =>
			request<Site>('PUT', `/sites/${id}`, { name, domain }),
		delete: (id: number) => request<void>('DELETE', `/sites/${id}`)
	},
	analytics: {
		overview: (siteId: number, from?: string, to?: string) => {
			const params = new URLSearchParams();
			if (from) params.set('from', from);
			if (to) params.set('to', to);
			const qs = params.toString();
			return request<AnalyticsOverview>('GET', `/analytics/${siteId}/overview${qs ? `?${qs}` : ''}`);
		}
	}
};

export interface Site {
	id: number;
	name: string;
	domain: string;
	owner_id: number;
	created_at: string;
	updated_at: string;
}

export interface AnalyticsOverview {
	total_pageviews: number;
	unique_visitors: number;
	top_pages: { path: string; count: number }[];
	top_referrers: { referrer: string; count: number }[];
	top_countries: { country: string; count: number }[];
	pageviews_per_day: { date: string; count: number }[];
}
