<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Workspace } from '$lib';
	import { Avatar, Badge, Button, Card, EmptyState, Input, Skeleton, icons, toast } from '@facile/muse';

	let workspaces = $state<Workspace[]>([]);
	let loading = $state(true);
	let search = $state('');

	const filtered = $derived(
		search.trim()
			? workspaces.filter((w) => w.name.toLowerCase().includes(search.trim().toLowerCase()))
			: workspaces
	);

	onMount(async () => {
		try {
			workspaces = await api.workspaces.list();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Could not load your spaces.');
		} finally {
			loading = false;
		}
	});

	const roleLabels: Record<string, string> = {
		owner: 'Owner',
		admin: 'Admin',
		editor: 'Editor',
		viewer: 'Viewer'
	};

	/* `owner` and `admin` are the two role tones in the shared vocabulary; everything else
	   is a plain member and gets `neutral`. */
	const roleTones = { owner: 'owner', admin: 'admin' } as const;

	function roleTone(role: string) {
		return roleTones[role as keyof typeof roleTones] ?? 'neutral';
	}
</script>

<svelte:head><title>Teams — Vision</title></svelte:head>

<div class="flex flex-col gap-10">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex min-w-0 flex-col gap-2">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">Teams</h1>
			<p class="text-fc-sm text-fc-fg-muted">
				Spaces you belong to. Everyone in a space sees its sites and their stats.
			</p>
		</div>
		<Button icon={icons.plus} onclick={() => goto('/team/new')}>New space</Button>
	</div>

	<section class="flex flex-col gap-4">
		{#if workspaces.length > 3}
			<Input bind:value={search} placeholder="Search spaces…" aria-label="Search spaces" />
		{/if}

		{#if loading}
			{#each [0, 1, 2] as row (row)}
				<Skeleton class="h-20 w-full rounded-fc-md" />
			{/each}
		{:else if filtered.length === 0}
			<EmptyState
				icon={icons.usersGroup}
				title={search ? `No space matches “${search}”` : 'No spaces yet'}
				description={search
					? 'Try another name, or clear the search.'
					: 'A space holds sites and the people who can read them.'}
			/>
		{:else}
			{#each filtered as ws (ws.id)}
				<a
					href="/team/{ws.id}"
					class="rounded-fc-md focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
				>
					<Card class="flex items-center gap-4 transition-colors hover:bg-fc-surface">
						<Avatar name={ws.name} />
						<div class="flex min-w-0 flex-1 flex-col gap-1">
							<div class="flex min-w-0 items-center gap-2">
								<span class="truncate text-fc-sm font-medium text-fc-fg">{ws.name}</span>
								<Badge tone={roleTone(ws.role)}>{roleLabels[ws.role] ?? ws.role}</Badge>
							</div>
							<span class="text-fc-xs text-fc-fg-muted">
								{ws.member_count}
								{ws.member_count === 1 ? 'member' : 'members'} · {ws.site_count}
								{ws.site_count === 1 ? 'site' : 'sites'}
							</span>
						</div>
						<iconify-icon
							icon={icons.arrow}
							width="18"
							height="18"
							class="block shrink-0 text-fc-fg-muted"
						></iconify-icon>
					</Card>
				</a>
			{/each}
		{/if}
	</section>
</div>
