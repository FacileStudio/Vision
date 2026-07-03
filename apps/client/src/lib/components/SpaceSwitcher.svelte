<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { api } from '$lib';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import type { Workspace } from '$lib/backend';
	import Icon from '@iconify/svelte';
	import ChevronsUpDown from '@lucide/svelte/icons/chevrons-up-down';
	import Check from '@lucide/svelte/icons/check';
	import Plus from '@lucide/svelte/icons/plus';

	let open = $state(false);
	let loading = $state(false);

	async function loadWorkspaces() {
		loading = true;
		try {
			const ws = await api.workspaces.list();
			workspaceStore.all = ws;
		} catch {
		} finally {
			loading = false;
		}
	}

	function toggle() {
		open = !open;
		if (open) loadWorkspaces();
	}

	async function switchTo(ws: Workspace) {
		workspaceStore.switchTo(ws);
		open = false;
		await invalidateAll();
		window.location.reload();
	}

	function handleClickOutside(e: MouseEvent) {
		const el = document.getElementById('ws-switcher');
		if (el && !el.contains(e.target as Node)) open = false;
	}

	$effect(() => {
		if (open) {
			document.addEventListener('click', handleClickOutside);
			return () => document.removeEventListener('click', handleClickOutside);
		}
	});
</script>

<div class="relative" id="ws-switcher">
	<button
		onclick={toggle}
		class="flex h-9 w-full items-center gap-2 rounded-lg border border-border/60 bg-muted/40 px-3 text-sm font-medium transition-colors hover:bg-muted"
	>
		<Icon icon="solar:share-circle-linear" class="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
		<span class="min-w-0 flex-1 truncate text-left">
			{workspaceStore.current?.name ?? 'Space'}
		</span>
		<ChevronsUpDown class="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
	</button>

	{#if open}
		<div class="absolute left-0 right-0 z-50 mt-1 overflow-hidden rounded-xl border border-border bg-background shadow-lg">
			{#if loading}
				<div class="flex items-center justify-center px-3 py-4">
					<div class="h-4 w-4 animate-spin rounded-full border-2 border-border border-t-foreground"></div>
				</div>
			{:else}
				<div class="py-1">
					{#each workspaceStore.all as ws (ws.id)}
						{@const active = workspaceStore.current?.id === ws.id}
						<button
							onclick={() => switchTo(ws)}
							class="flex h-9 w-full items-center gap-2 px-3 text-sm transition-colors {active
								? 'bg-muted font-medium text-foreground'
								: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
						>
							<Icon icon="solar:share-circle-linear" class="h-3.5 w-3.5 shrink-0" />
							<span class="min-w-0 flex-1 truncate text-left">{ws.name}</span>
							{#if active}
								<Check class="ml-auto h-3.5 w-3.5 shrink-0" />
							{/if}
						</button>
					{/each}

					<div class="my-1 border-t border-border"></div>

					<a
						href="/team/new"
						onclick={() => (open = false)}
						class="flex h-9 w-full items-center gap-2 px-3 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
					>
						<Plus class="h-3.5 w-3.5 shrink-0" />
						<span>New space</span>
					</a>
				</div>
			{/if}
		</div>
	{/if}
</div>
