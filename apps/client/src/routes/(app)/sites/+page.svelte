<script lang="ts">
	import { api } from '$lib';
	import type { Site } from '$lib';
	import { Button, Card, ConfirmModal, Skeleton, icons, toast } from '@facile/muse';
	import AddSiteDrawer from '$lib/components/add-site-drawer.svelte';
	import SiteFavicon from '$lib/components/site-favicon.svelte';
	import { workspaceStore } from '$lib/stores/workspace.svelte';

	let sites = $state<Site[]>([]);
	let loading = $state(true);
	let drawerOpen = $state(false);

	let pendingDelete = $state<Site | null>(null);
	let confirmOpen = $state(false);

	async function loadSites() {
		loading = true;
		try {
			sites = await api.sites.list(workspaceStore.current?.id);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Could not load your sites.');
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		workspaceStore.current;
		loadSites();
	});

	async function deleteSite() {
		const site = pendingDelete;
		if (!site) return;
		await api.sites.delete(site.id);
		sites = sites.filter((s) => s.id !== site.id);
		pendingDelete = null;
		toast.neutral(`Deleted ${site.name}.`);
	}
</script>

<svelte:head><title>Sites — Vision</title></svelte:head>

<AddSiteDrawer
	bind:open={drawerOpen}
	onCreated={(site) => {
		sites = [site, ...sites];
		toast.success(`${site.name} is now being tracked.`);
	}}
/>

<div class="flex flex-col gap-10">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex min-w-0 flex-col gap-2">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">Sites</h1>
			<p class="text-fc-sm text-fc-fg-muted">
				Everything {workspaceStore.current?.name ?? 'this space'} is tracking.
			</p>
		</div>
		<Button icon={icons.plus} onclick={() => (drawerOpen = true)}>Add site</Button>
	</div>

	<section class="flex flex-col gap-4">
		{#if loading}
			{#each [0, 1, 2] as row (row)}
				<Skeleton class="h-16 w-full rounded-fc-md" />
			{/each}
		{:else if sites.length === 0}
			<Card class="flex flex-col items-center gap-3 py-12 text-center">
				<p class="text-fc-sm font-medium text-fc-fg">No sites yet</p>
				<p class="max-w-sm text-fc-sm text-fc-fg-muted">
					Add a domain and drop one script tag on it. Numbers start arriving on the next
					pageview.
				</p>
				<Button size="lg" icon={icons.plus} onclick={() => (drawerOpen = true)}>Add site</Button>
			</Card>
		{:else}
			{#each sites as site (site.id)}
				<Card class="flex items-center gap-4">
					<a
						href="/sites/{site.id}"
						class="flex min-w-0 flex-1 items-center gap-3 rounded-fc-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
					>
						<SiteFavicon domain={site.domain} name={site.name} class="size-8" />
						<span class="flex min-w-0 flex-col">
							<span class="truncate text-fc-sm font-medium text-fc-fg">{site.name}</span>
							<span class="truncate text-fc-xs text-fc-fg-muted">{site.domain}</span>
						</span>
					</a>
					<Button
						variant="ghost-danger"
						icon={icons.remove}
						aria-label="Delete {site.name}"
						onclick={() => {
							pendingDelete = site;
							confirmOpen = true;
						}}
					>
						Delete
					</Button>
				</Card>
			{/each}
		{/if}
	</section>
</div>

<ConfirmModal
	bind:open={confirmOpen}
	tone="danger"
	title="Delete {pendingDelete?.name ?? 'this site'}?"
	description="Every pageview, session and goal recorded for {pendingDelete?.domain ??
		'this domain'} goes with it, and the script on the site stops being accepted. This cannot be undone."
	confirmLabel="Delete site"
	cancelLabel="Keep it"
	onConfirm={deleteSite}
	onCancel={() => (pendingDelete = null)}
/>
