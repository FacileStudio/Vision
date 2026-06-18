<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';

	let name = $state('');
	let creating = $state(false);
	let error = $state('');

	async function create() {
		const trimmed = name.trim();
		if (!trimmed || trimmed.length < 2) {
			error = 'Name must be at least 2 characters.';
			return;
		}
		creating = true;
		error = '';
		try {
			const ws = await api.workspaces.create(trimmed);
			goto(`/team/${ws.id}`);
		} catch (e: any) {
			error = e.message || 'Failed to create space.';
		} finally {
			creating = false;
		}
	}
</script>

<svelte:head><title>New space — Vision</title></svelte:head>

<div class="mx-auto max-w-md">
	<a href="/team" class="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground">
		<ArrowLeft class="h-4 w-4" />
		Back to teams
	</a>

	<h1 class="mb-1 text-2xl font-bold">New space</h1>
	<p class="mb-6 text-sm text-muted-foreground">
		Create a space to collaborate with your team. You'll be the owner.
	</p>

	<form onsubmit={(e) => { e.preventDefault(); create(); }} class="space-y-4">
		<div>
			<label for="ws-name" class="mb-1.5 block text-sm font-medium">Name</label>
			<Input id="ws-name" bind:value={name} placeholder="My team" autofocus />
		</div>

		{#if error}
			<p class="text-sm text-destructive">{error}</p>
		{/if}

		<Button type="submit" class="w-full" disabled={creating || !name.trim()}>
			{creating ? 'Creating...' : 'Create space'}
		</Button>
	</form>
</div>
