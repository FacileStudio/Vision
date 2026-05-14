<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Site } from '$lib';

	let sites = $state<Site[]>([]);

	onMount(async () => {
		sites = await api.sites.list();
	});
</script>

<svelte:head><title>Dashboard — Vision</title></svelte:head>

<h1 class="text-2xl font-bold mb-6">Dashboard</h1>

{#if sites.length === 0}
	<p class="text-muted-foreground">No sites yet. <a href="/sites" class="underline">Add one</a> to start tracking.</p>
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
