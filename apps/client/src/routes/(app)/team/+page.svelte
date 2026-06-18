<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Workspace } from '$lib';
	import { Button } from '$lib/components/ui/button/index.js';
	import Icon from '@iconify/svelte';
	import Plus from '@lucide/svelte/icons/plus';
	import Users from '@lucide/svelte/icons/users';
	import Globe from '@lucide/svelte/icons/globe';

	let workspaces = $state<Workspace[]>([]);
	let loading = $state(true);
	let search = $state('');

	const filtered = $derived(
		search.trim()
			? workspaces.filter((w) =>
					w.name.toLowerCase().includes(search.trim().toLowerCase())
				)
			: workspaces
	);

	onMount(async () => {
		try {
			workspaces = await api.workspaces.list();
		} catch {
		} finally {
			loading = false;
		}
	});

	function roleLabel(role: string): string {
		switch (role) {
			case 'owner':
				return 'Owner';
			case 'admin':
				return 'Admin';
			case 'editor':
				return 'Editor';
			case 'viewer':
				return 'Viewer';
			default:
				return role;
		}
	}

	function roleBadgeClass(role: string): string {
		switch (role) {
			case 'owner':
				return 'bg-foreground text-background';
			case 'admin':
				return 'bg-primary/10 text-primary';
			default:
				return 'bg-muted text-muted-foreground';
		}
	}
</script>

<svelte:head><title>Teams — Vision</title></svelte:head>

<div class="mx-auto max-w-2xl">
	<div class="mb-6 flex items-center justify-between">
		<h1 class="text-2xl font-bold">Teams</h1>
		<Button size="sm" onclick={() => goto('/team/new')}>
			<Plus class="mr-1.5 h-4 w-4" />
			New space
		</Button>
	</div>

	{#if workspaces.length > 3}
		<div class="mb-4">
			<input
				bind:value={search}
				type="text"
				placeholder="Search spaces..."
				class="h-9 w-full rounded-lg border border-border bg-background px-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
			/>
		</div>
	{/if}

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="h-5 w-5 animate-spin rounded-full border-2 border-border border-t-foreground"></div>
		</div>
	{:else if filtered.length === 0}
		<div class="rounded-lg border border-dashed p-8 text-center">
			<Icon icon="solar:users-group-rounded-linear" class="mx-auto mb-3 h-10 w-10 text-muted-foreground" />
			<p class="text-sm text-muted-foreground">
				{search ? 'No spaces match your search.' : 'No spaces yet. Create one to get started.'}
			</p>
		</div>
	{:else}
		<div class="space-y-2">
			{#each filtered as ws (ws.id)}
				<a
					href="/team/{ws.id}"
					class="group flex items-center gap-4 rounded-lg border border-border/60 p-4 transition-colors hover:bg-muted/50"
				>
					<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-foreground text-sm font-bold text-background">
						{ws.name.charAt(0).toUpperCase()}
					</div>
					<div class="min-w-0 flex-1">
						<div class="flex items-center gap-2">
							<p class="truncate text-sm font-semibold">{ws.name}</p>
							<span class="shrink-0 rounded-full px-2 py-0.5 text-[10px] font-medium {roleBadgeClass(ws.role)}">
								{roleLabel(ws.role)}
							</span>
						</div>
						<div class="mt-0.5 flex items-center gap-3 text-xs text-muted-foreground">
							<span class="flex items-center gap-1">
								<Users class="h-3 w-3" />
								{ws.member_count} {ws.member_count === 1 ? 'member' : 'members'}
							</span>
							<span class="flex items-center gap-1">
								<Globe class="h-3 w-3" />
								{ws.site_count} {ws.site_count === 1 ? 'site' : 'sites'}
							</span>
						</div>
					</div>
					<Icon icon="solar:alt-arrow-right-linear" class="h-4 w-4 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100" />
				</a>
			{/each}
		</div>
	{/if}
</div>
