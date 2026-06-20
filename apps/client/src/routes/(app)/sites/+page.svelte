<script lang="ts">
	import { api } from '$lib';
	import type { Site } from '$lib';
	import { Button } from '$lib/components/ui/button/index.js';
	import Icon from '@iconify/svelte';
	import AddSiteDrawer from '$lib/components/add-site-drawer.svelte';
	import SiteFavicon from '$lib/components/site-favicon.svelte';
	import { workspaceStore } from '$lib/stores/workspace.svelte';

	let sites = $state<Site[]>([]);
	let drawerOpen = $state(false);

	async function loadSites() {
		sites = await api.sites.list(workspaceStore.current?.id);
	}

	$effect(() => {
		workspaceStore.current;
		loadSites();
	});

	async function deleteSite(id: number) {
		await api.sites.delete(id);
		sites = sites.filter((s) => s.id !== id);
	}
</script>

<svelte:head><title>Sites — Vision</title></svelte:head>

<AddSiteDrawer bind:open={drawerOpen} onCreated={(site) => { sites = [site, ...sites]; }} />

<div class="flex items-center justify-between mb-6">
	<h1 class="text-2xl font-bold">Sites</h1>
	<Button variant="default" size="sm" onclick={() => (drawerOpen = true)}>
		<Icon icon="mdi:plus" class="h-4 w-4 mr-1.5" />
		Add site
	</Button>
</div>

<div class="space-y-2">
	{#each sites as site}
		<div class="flex items-center justify-between rounded-lg border p-4">
			<a href="/sites/{site.id}" class="flex items-center gap-3 hover:underline">
				<SiteFavicon domain={site.domain} name={site.name} class="h-5 w-5" />
				<span class="font-medium">{site.name}</span>
				<span class="text-sm text-muted-foreground">{site.domain}</span>
			</a>
			<button onclick={() => deleteSite(site.id)} class="flex items-center gap-1 rounded-md p-1.5 text-red-500 transition-colors hover:bg-red-500/10">
				<Icon icon="solar:trash-bin-trash-linear" class="h-4 w-4" />
			</button>
		</div>
	{/each}
</div>
