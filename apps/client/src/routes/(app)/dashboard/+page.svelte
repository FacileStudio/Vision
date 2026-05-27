<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Site } from '$lib';
	import { Button } from '$lib/components/ui/button/index.js';
	import Icon from '@iconify/svelte';
	import AddSiteDrawer from '$lib/components/add-site-drawer.svelte';

	let sites = $state<Site[]>([]);
	let drawerOpen = $state(false);

	onMount(async () => {
		sites = await api.sites.list();
	});
</script>

<svelte:head><title>Dashboard — Vision</title></svelte:head>

<AddSiteDrawer bind:open={drawerOpen} onCreated={(site) => { sites = [site, ...sites]; }} />

<div class="flex items-center justify-between mb-6">
	<h1 class="text-2xl font-bold">Dashboard</h1>
	<Button variant="default" size="sm" onclick={() => (drawerOpen = true)}>
		<Icon icon="mdi:plus" class="h-4 w-4 mr-1.5" />
		Add site
	</Button>
</div>

{#if sites.length === 0}
	<div class="flex flex-col items-center justify-center py-16 text-center">
		<Icon icon="solar:chart-linear" class="h-12 w-12 text-muted-foreground mb-4" />
		<p class="text-muted-foreground mb-4">No sites yet. Add one to start tracking.</p>
		<Button onclick={() => (drawerOpen = true)}>
			<Icon icon="mdi:plus" class="h-4 w-4 mr-1.5" />
			Add your first site
		</Button>
	</div>
{:else}
	<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
		{#each sites as site}
			<a
				href="/sites/{site.id}"
				class="flex items-start gap-3 rounded-lg border bg-card p-6 hover:shadow-sm transition-shadow"
			>
				<img src="https://www.google.com/s2/favicons?domain={site.domain}&sz=32" alt="" class="h-6 w-6 shrink-0 rounded mt-0.5" />
				<div>
					<h3 class="font-semibold">{site.name}</h3>
					<p class="text-sm text-muted-foreground">{site.domain}</p>
				</div>
			</a>
		{/each}
	</div>
{/if}
