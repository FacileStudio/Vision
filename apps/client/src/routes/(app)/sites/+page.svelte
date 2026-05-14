<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Site } from '$lib';
	import Icon from '@iconify/svelte';

	let sites = $state<Site[]>([]);
	let name = $state('');
	let domain = $state('');
	let error = $state('');

	onMount(async () => {
		sites = await api.sites.list();
	});

	async function addSite() {
		error = '';
		try {
			const site = await api.sites.create(name, domain);
			sites = [site, ...sites];
			name = '';
			domain = '';
		} catch (e: any) {
			error = e.message;
		}
	}

	async function deleteSite(id: number) {
		await api.sites.delete(id);
		sites = sites.filter((s) => s.id !== id);
	}
</script>

<h1 class="text-2xl font-bold mb-6">Sites</h1>

<form onsubmit={addSite} class="flex gap-2 mb-6">
	<input bind:value={name} placeholder="Site name" class="rounded-md border bg-background px-3 py-2 text-sm" required />
	<input bind:value={domain} placeholder="example.com" class="rounded-md border bg-background px-3 py-2 text-sm" required />
	<button type="submit" class="flex items-center gap-1.5 rounded-md bg-primary px-4 py-2 text-sm text-primary-foreground">
		<Icon icon="mdi:plus" class="h-4 w-4" />
		Add
	</button>
</form>

{#if error}
	<p class="text-destructive text-sm mb-4">{error}</p>
{/if}

<div class="space-y-2">
	{#each sites as site}
		<div class="flex items-center justify-between rounded-lg border p-4">
			<a href="/sites/{site.id}" class="hover:underline">
				<span class="font-medium">{site.name}</span>
				<span class="text-sm text-muted-foreground ml-2">{site.domain}</span>
			</a>
			<button onclick={() => deleteSite(site.id)} class="flex items-center gap-1 rounded-md p-1.5 text-red-500 transition-colors hover:bg-red-500/10">
				<Icon icon="solar:trash-bin-trash-linear" class="h-4 w-4" />
			</button>
		</div>
	{/each}
</div>
