import type { Workspace } from '$lib/backend';

let current = $state<Workspace | null>(null);
let all = $state<Workspace[]>([]);

export const workspaceStore = {
	get current() { return current; },
	set current(w: Workspace | null) { current = w; },
	get all() { return all; },
	set all(w: Workspace[]) { all = w; }
};
