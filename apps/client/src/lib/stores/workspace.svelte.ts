import type { Workspace } from '$lib/backend';

const STORAGE_KEY = 'vision-workspace-id';

let current = $state<Workspace | null>(null);
let all = $state<Workspace[]>([]);

function persistId(id: number | null) {
	if (typeof window === 'undefined') return;
	if (id !== null) {
		localStorage.setItem(STORAGE_KEY, String(id));
	} else {
		localStorage.removeItem(STORAGE_KEY);
	}
}

function getSavedId(): number | null {
	if (typeof window === 'undefined') return null;
	const raw = localStorage.getItem(STORAGE_KEY);
	if (!raw) return null;
	const n = parseInt(raw, 10);
	return isNaN(n) ? null : n;
}

export const workspaceStore = {
	get current() {
		return current;
	},
	set current(w: Workspace | null) {
		current = w;
		persistId(w?.id ?? null);
	},
	get all() {
		return all;
	},
	set all(list: Workspace[]) {
		all = list;
	},

	hydrate(workspaces: Workspace[]) {
		all = workspaces;
		const savedId = getSavedId();
		const saved = savedId !== null ? workspaces.find((w) => w.id === savedId) : null;
		current = saved ?? workspaces[0] ?? null;
		persistId(current?.id ?? null);
	},

	switchTo(workspace: Workspace) {
		current = workspace;
		persistId(workspace.id);
	},

	get isOwnerOrAdmin() {
		return current?.role === 'owner' || current?.role === 'admin';
	}
};
