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

	const res = await fetch(`/api${path}`, {
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
			request<{ user_id: string; token: string }>('POST', '/auth/login', { email, password }),
		me: () => request<UserProfile>('GET', '/auth/me'),
		updateProfile: (name: string, email: string) =>
			request<UserProfile>('PUT', '/auth/me', { name, email }),
		changePassword: (currentPassword: string, newPassword: string) =>
			request<{ status: string }>('PUT', '/auth/password', {
				current_password: currentPassword,
				new_password: newPassword
			}),
		syncProfile: () => request<{ status: string }>('POST', '/auth/sync-profile')
	},
	workspaces: {
		list: () => request<Workspace[]>('GET', '/workspaces'),
		create: (name: string) => request<Workspace>('POST', '/workspaces', { name }),
		get: (id: number) => request<Workspace>('GET', `/workspaces/${id}`),
		update: (id: number, name: string) => request<Workspace>('PUT', `/workspaces/${id}`, { name }),
		delete: (id: number) => request<void>('DELETE', `/workspaces/${id}`),
		members: (id: number) => request<WorkspaceMember[]>('GET', `/workspaces/${id}/members`),
		addMember: (id: number, email: string, role: string) =>
			request<WorkspaceMember>('POST', `/workspaces/${id}/members`, { email, role }),
		updateMember: (id: number, userId: number, role: string) =>
			request<void>('PUT', `/workspaces/${id}/members/${userId}`, { role }),
		removeMember: (id: number, userId: number) =>
			request<void>('DELETE', `/workspaces/${id}/members/${userId}`),
		leave: (id: number) =>
			request<void>('POST', `/workspaces/${id}/leave`)
	},
	sites: {
		list: () => request<Site[]>('GET', '/sites'),
		get: (id: number) => request<Site>('GET', `/sites/${id}`),
		create: (name: string, domain: string, workspaceId: number) =>
			request<Site>('POST', '/sites', { name, domain, workspace_id: workspaceId }),
		update: (id: number, name: string, domain: string) =>
			request<Site>('PUT', `/sites/${id}`, { name, domain }),
		delete: (id: number) => request<void>('DELETE', `/sites/${id}`),
		share: (id: number) => request<Site>('POST', `/sites/${id}/share`),
		revokeShare: (id: number) => request<void>('DELETE', `/sites/${id}/share`)
	},
	webhooks: {
		list: () => request<Webhook[]>('GET', '/webhooks'),
		create: (data: CreateWebhookRequest) => request<Webhook>('POST', '/webhooks', data),
		update: (id: number, data: UpdateWebhookRequest) => request<Webhook>('PUT', `/webhooks/${id}`, data),
		delete: (id: number) => request<void>('DELETE', `/webhooks/${id}`),
		test: (id: number) => request<void>('POST', `/webhooks/${id}/test`)
	},
	analytics: {
		overview: (siteId: number, from?: string, to?: string, granularity?: string, filters?: Record<string, string>) => {
			const params = new URLSearchParams();
			if (from) params.set('from', from);
			if (to) params.set('to', to);
			if (granularity) params.set('granularity', granularity);
			if (filters) {
				for (const [k, v] of Object.entries(filters)) {
					if (v) params.set(k, v);
				}
			}
			const qs = params.toString();
			return request<AnalyticsOverview>('GET', `/analytics/${siteId}/overview${qs ? `?${qs}` : ''}`);
		},
		realtime: {
			visitors: (siteId: number) =>
				request<{ visitors: number }>('GET', `/analytics/${siteId}/realtime`)
		}
	},
	share: {
		overview: async (token: string, from?: string, to?: string, granularity?: string) => {
			const params = new URLSearchParams();
			if (from) params.set('from', from);
			if (to) params.set('to', to);
			if (granularity) params.set('granularity', granularity);
			const qs = params.toString();
			const res = await fetch(`/api/share/${token}${qs ? `?${qs}` : ''}`);
			if (!res.ok) throw new Error('Not found');
			return res.json() as Promise<{ site: { name: string; domain: string }; overview: AnalyticsOverview }>;
		},
		realtime: async (token: string) => {
			const res = await fetch(`/api/share/${token}/realtime`);
			if (!res.ok) throw new Error('Not found');
			return res.json() as Promise<{ visitors: number }>;
		}
	},
	goals: {
		list: (siteId: number) => request<Goal[]>('GET', `/goals?site_id=${siteId}`),
		create: (data: CreateGoalRequest) => request<Goal>('POST', '/goals', data),
		update: (id: number, data: UpdateGoalRequest) => request<Goal>('PUT', `/goals/${id}`, data),
		delete: (id: number) => request<void>('DELETE', `/goals/${id}`),
		conversions: (siteId: number, from?: string, to?: string) => {
			const params = new URLSearchParams();
			if (from) params.set('from', from);
			if (to) params.set('to', to);
			const qs = params.toString();
			return request<GoalConversionsResponse>('GET', `/goals/${siteId}/conversions${qs ? `?${qs}` : ''}`);
		}
	},
	apiKeys: {
		list: () => request<APIKeyItem[]>('GET', '/api-keys'),
		create: (data: CreateAPIKeyRequest) => request<CreateAPIKeyResponse>('POST', '/api-keys', data),
		revoke: (id: number) => request<void>('DELETE', `/api-keys/${id}`)
	},
	events: {}
};

export interface UserProfile {
	id: string;
	email: string;
	name: string;
	avatar_url: string;
	avatar_source: string;
	created_at: string;
}

export interface Site {
	id: number;
	name: string;
	domain: string;
	owner_id: number;
	workspace_id: number;
	share_token: string | null;
	created_at: string;
	updated_at: string;
}

export interface Workspace {
	id: number;
	name: string;
	role: string;
	member_count: number;
	site_count: number;
	created_at: string;
	updated_at: string;
}

export interface WorkspaceMember {
	id: number;
	user_id: number;
	email: string;
	name: string;
	avatar_url: string;
	role: string;
	created_at: string;
}

export interface Webhook {
	id: number;
	url: string;
	period: string;
	interval_hours: number;
	enabled: boolean;
	last_sent_at: string | null;
	created_at: string;
	updated_at: string;
}

export interface CreateWebhookRequest {
	url: string;
	secret: string;
	interval_hours: number;
}

export interface UpdateWebhookRequest {
	url: string;
	secret: string;
	interval_hours: number;
	enabled: boolean;
}

export interface Goal {
	id: number;
	site_id: number;
	name: string;
	goal_type: string;
	event_name: string | null;
	page_path: string | null;
	match_type: string;
	created_at: string;
}

export interface CreateGoalRequest {
	site_id: number;
	name: string;
	goal_type: string;
	event_name?: string;
	page_path?: string;
	match_type?: string;
}

export interface UpdateGoalRequest {
	name: string;
	goal_type: string;
	event_name?: string;
	page_path?: string;
	match_type?: string;
}

export interface GoalConversion {
	id: number;
	name: string;
	goal_type: string;
	conversions: number;
	conversion_rate: number;
}

export interface GoalConversionsResponse {
	goals: GoalConversion[];
	total_visitors: number;
}

export interface APIKeyItem {
	id: number;
	name: string;
	prefix: string;
	key_hint: string;
	scopes: string;
	site_id: number | null;
	is_active: boolean;
	last_used_at: string | null;
	expires_at: string | null;
	created_at: string;
}

export interface CreateAPIKeyRequest {
	name: string;
	scopes: string;
	site_id?: number;
}

export interface CreateAPIKeyResponse {
	key: string;
	api_key: APIKeyItem;
}

export interface AnalyticsOverview {
	total_pageviews: number;
	unique_visitors: number;
	prev_total_pageviews: number;
	prev_unique_visitors: number;
	top_pages: { path: string; count: number }[];
	top_referrers: { referrer: string; count: number }[];
	top_countries: { country: string; count: number }[];
	top_browsers: { browser: string; count: number }[];
	top_os: { os: string; count: number }[];
	top_devices: { device: string; count: number }[];
	top_screens: { screen: string; count: number }[];
	top_entry_pages: { path: string; count: number }[];
	top_exit_pages: { path: string; count: number }[];
	top_utm_sources: { value: string; count: number }[];
	top_utm_mediums: { value: string; count: number }[];
	top_utm_campaigns: { value: string; count: number }[];
	pageviews_per_day: { date: string; count: number }[];
	unique_visitors_per_day: { date: string; count: number }[];
	prev_pageviews_per_day: { date: string; count: number }[];
	prev_unique_visitors_per_day: { date: string; count: number }[];
	hourly_distribution: { hour: number; count: number }[];
	performance: {
		avg_dns: number;
		avg_tcp: number;
		avg_ttfb: number;
		avg_dom_load: number;
		avg_page_load: number;
		sample_count: number;
	} | null;
	top_events: { name: string; count: number }[];
	bounce_rate: number;
	avg_session_duration: number;
	pages_per_session: number;
	prev_bounce_rate: number;
	prev_avg_session_duration: number;
	prev_pages_per_session: number;
}
