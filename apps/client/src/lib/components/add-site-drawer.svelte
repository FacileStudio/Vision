<script lang="ts">
	import { api } from '$lib';
	import type { Site } from '$lib';
	import { Alert, Button, Drawer, Field, Input, Select, Spinner, icons } from '@facile/muse';
	import { workspaceStore } from '$lib/stores/workspace.svelte';

	let {
		open = $bindable(false),
		onCreated
	}: {
		open?: boolean;
		onCreated: (site: Site) => void;
	} = $props();

	let name = $state('');
	let domain = $state('');
	let workspaceId = $state('');
	let error = $state('');
	let submitting = $state(false);

	/* Reopening must never show the previous attempt's error or half-typed domain. */
	$effect(() => {
		if (!open) {
			name = '';
			domain = '';
			error = '';
			return;
		}
		if (workspaceStore.current) workspaceId = String(workspaceStore.current.id);
	});

	async function addSite(e: Event) {
		e.preventDefault();
		error = '';
		const id = Number(workspaceId);
		if (!id) {
			error = 'No space available — create one first.';
			return;
		}
		submitting = true;
		try {
			const site = await api.sites.create(name.trim(), domain.trim(), id);
			open = false;
			onCreated(site);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not add the site.';
		} finally {
			submitting = false;
		}
	}
</script>

<Drawer bind:open title="Add a site" description="Vision starts counting as soon as the script is on the page." showClose>
	<form id="add-site" class="flex flex-col gap-4" onsubmit={addSite}>
		{#if workspaceStore.all.length > 1}
			<Field label="Space" helper="Everyone in the space can read this site's stats.">
				<Select bind:value={workspaceId} disabled={submitting}>
					{#each workspaceStore.all as ws (ws.id)}
						<option value={String(ws.id)}>{ws.name}</option>
					{/each}
				</Select>
			</Field>
		{/if}

		<Field label="Name" helper="How it appears in your site list.">
			<Input bind:value={name} placeholder="My website" required disabled={submitting} />
		</Field>

		<Field label="Domain" helper="Events are verified against this domain — no subdomain, no scheme.">
			<Input bind:value={domain} placeholder="example.com" required disabled={submitting} />
		</Field>

		{#if error}
			<Alert tone="danger" title="Not added">{error}</Alert>
		{/if}
	</form>

	{#snippet footer()}
		<div class="flex gap-2">
			<Button variant="ghost" size="lg" class="flex-1" disabled={submitting} onclick={() => (open = false)}>
				Cancel
			</Button>
			<Button
				type="submit"
				form="add-site"
				size="lg"
				icon={icons.plus}
				class="flex-1"
				disabled={submitting || !name.trim() || !domain.trim()}
			>
				{#if submitting}<Spinner size="sm" />{/if}
				{submitting ? 'Adding…' : 'Add site'}
			</Button>
		</div>
	{/snippet}
</Drawer>
