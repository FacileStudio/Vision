<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Site } from '$lib';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import Icon from '@iconify/svelte';

	let sites = $state<Site[]>([]);
	let drawerOpen = $state(false);
	let name = $state('');
	let domain = $state('');
	let error = $state('');
	let submitting = $state(false);

	onMount(async () => {
		sites = await api.sites.list();
	});

	async function addSite() {
		error = '';
		submitting = true;
		try {
			const site = await api.sites.create(name, domain);
			sites = [site, ...sites];
			name = '';
			domain = '';
			drawerOpen = false;
		} catch (e: any) {
			error = e.message;
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head><title>Dashboard — Vision</title></svelte:head>

<div class="flex items-center justify-between mb-6">
	<h1 class="text-2xl font-bold">Dashboard</h1>
	<Button variant="default" size="sm" onclick={() => (drawerOpen = true)}>
		<Icon icon="solar:add-circle-linear" class="h-4 w-4 mr-1.5" />
		Add site
	</Button>
</div>

{#if sites.length === 0}
	<div class="flex flex-col items-center justify-center py-16 text-center">
		<Icon icon="solar:chart-linear" class="h-12 w-12 text-muted-foreground mb-4" />
		<p class="text-muted-foreground mb-4">No sites yet. Add one to start tracking.</p>
		<Button onclick={() => (drawerOpen = true)}>
			<Icon icon="solar:add-circle-linear" class="h-4 w-4 mr-1.5" />
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

<Sheet.Root bind:open={drawerOpen}>
	<Sheet.Content side="right">
		<Sheet.Header>
			<Sheet.Title>Add a new site</Sheet.Title>
			<Sheet.Description>Enter the details of the website you want to track.</Sheet.Description>
		</Sheet.Header>
		<form onsubmit={(e) => { e.preventDefault(); addSite(); }} class="flex flex-col gap-4 px-4">
			<div class="space-y-2">
				<Label for="site-name">Name</Label>
				<Input id="site-name" bind:value={name} placeholder="My Website" required />
			</div>
			<div class="space-y-2">
				<Label for="site-domain">Domain</Label>
				<Input id="site-domain" bind:value={domain} placeholder="example.com" required />
			</div>
			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}
			<Button type="submit" disabled={submitting} class="w-full">
				{submitting ? 'Adding…' : 'Add site'}
			</Button>
		</form>
	</Sheet.Content>
</Sheet.Root>
