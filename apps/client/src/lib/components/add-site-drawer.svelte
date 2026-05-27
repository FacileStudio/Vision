<script lang="ts">
	import { api } from '$lib';
	import type { Site } from '$lib';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';

	let {
		open = $bindable(false),
		onCreated
	}: {
		open: boolean;
		onCreated: (site: Site) => void;
	} = $props();

	let name = $state('');
	let domain = $state('');
	let error = $state('');
	let submitting = $state(false);

	$effect(() => {
		if (!open) {
			name = '';
			domain = '';
			error = '';
		}
	});

	async function addSite() {
		error = '';
		submitting = true;
		try {
			const site = await api.sites.create(name, domain);
			open = false;
			onCreated(site);
		} catch (e: any) {
			error = e.message;
		} finally {
			submitting = false;
		}
	}
</script>

<Sheet.Root bind:open>
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
