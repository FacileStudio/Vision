import type { Workspace } from '$lib/backend';

let all = $state<Workspace[]>([]);

const WRITE_ROLES = ['owner', 'admin', 'editor'];

/** The workspaces the signed-in person belongs to, loaded once by the app layout. */
export const workspaceStore = {
	get all() {
		return all;
	},
	set all(list: Workspace[]) {
		all = list;
	},

	hydrate(workspaces: Workspace[]) {
		all = workspaces;
	},

	/** The workspaces the person can add sites, keys and webhooks to. */
	get writable() {
		return all.filter((w) => WRITE_ROLES.includes(w.role));
	}
};

/** The space filter a page carries in `?space=`: `all` when absent, otherwise a workspace id. */
export function spaceFilter(url: URL): string {
	return url.searchParams.get('space') || 'all';
}
