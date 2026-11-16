import { ApiError, request } from './backend';
import type { Anomaly, BulleProfile, BulleRun, Explanation, IdeaRunResponse, IdeasResponse, PageContext } from './bulle';

const quiet = <T>(method: string, path: string, body?: unknown) => request<T>(method, path, body, true);

const spaceQuery = (spaceId: number | null) => (spaceId ? `?space_id=${spaceId}` : '');

/** Vision's calls to its own Bulle routes. Every one is quiet: the callers show the failure themselves. */
export const bulleApi = {
	anomalies: (siteId: number, from?: string, to?: string) => {
		const params = new URLSearchParams({ limit: '30' });
		if (from) params.set('from', from);
		if (to) params.set('to', to);
		return quiet<{ anomalies: Anomaly[] }>('GET', `/analytics/${siteId}/anomalies?${params}`);
	},
	explain: (anomalyId: number, lang: string) => quiet<Explanation>('POST', `/anomalies/${anomalyId}/explain?lang=${lang}`),
	profile: () => quiet<BulleProfile>('GET', '/bulle/profile'),
	startRun: (spaceId: number | null, body: { title: string; prompt: string; page?: PageContext }) =>
		quiet<{ run_id: string; session_id: string }>('POST', `/bulle/runs${spaceQuery(spaceId)}`, body),
	run: (runId: string) => quiet<BulleRun>('GET', `/bulle/runs/${encodeURIComponent(runId)}`),
	ideas: {
		list: (spaceId: number | null) => quiet<IdeasResponse>('GET', `/ideas${spaceQuery(spaceId)}`),
		generate: (spaceId: number | null) => quiet<{ batch_id: string }>('POST', `/ideas/generate${spaceQuery(spaceId)}`),
		run: (id: string, spaceId: number | null) =>
			quiet<{ run_id: string }>('POST', `/ideas/${encodeURIComponent(id)}/run${spaceQuery(spaceId)}`),
		dismiss: (id: string) => quiet<void>('POST', `/ideas/${encodeURIComponent(id)}/dismiss`),
		status: (runId: string) => quiet<IdeaRunResponse>('GET', `/ideas/runs/${encodeURIComponent(runId)}`),
		decide: (runId: string, callId: string, approve: boolean) =>
			quiet<void>('POST', `/ideas/runs/${encodeURIComponent(runId)}/approvals/${encodeURIComponent(callId)}`, { approve })
	}
};

export { ApiError };
