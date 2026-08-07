<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import { Alert, Button, Card, Field, Input, Spinner, icons, toast } from '@facile/muse';

	let name = $state('');
	let creating = $state(false);
	let error = $state('');

	const tooShort = $derived(name.trim().length > 0 && name.trim().length < 2);

	async function create(e: Event) {
		e.preventDefault();
		const trimmed = name.trim();
		if (trimmed.length < 2) {
			error = 'A space name needs at least two characters.';
			return;
		}
		creating = true;
		error = '';
		try {
			const ws = await api.workspaces.create(trimmed);
			toast.success(`${ws.name} is ready.`);
			goto(`/team/${ws.id}`);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not create the space.';
		} finally {
			creating = false;
		}
	}
</script>

<svelte:head><title>New space — Vision</title></svelte:head>

<div class="mx-auto flex w-full max-w-fc-sm flex-col gap-8">
	<a
		href="/team"
		class="inline-flex items-center gap-1.5 self-start rounded-fc-sm text-fc-sm text-fc-fg-muted transition-colors hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
	>
		<iconify-icon icon={icons.chevronLeft} width="16" height="16" class="block"></iconify-icon>
		Back to teams
	</a>

	<div class="flex flex-col gap-2">
		<h1 class="text-fc-2xl font-semibold text-fc-fg">New space</h1>
		<p class="text-fc-sm text-fc-fg-muted">
			A space groups sites and the people who can read them. You will be its owner.
		</p>
	</div>

	<Card>
		<form class="flex flex-col gap-4" onsubmit={create}>
			<Field
				label="Name"
				error={tooShort ? 'At least two characters.' : undefined}
				helper="Usually a client, a team or a product."
			>
				<Input bind:value={name} placeholder="Acme Studio" disabled={creating} />
			</Field>

			{#if error}
				<Alert tone="danger" title="Not created">{error}</Alert>
			{/if}

			<Button
				type="submit"
				size="lg"
				icon={icons.plus}
				class="self-start"
				disabled={creating || name.trim().length < 2}
			>
				{#if creating}<Spinner size="sm" />{/if}
				{creating ? 'Creating…' : 'Create space'}
			</Button>
		</form>
	</Card>
</div>
