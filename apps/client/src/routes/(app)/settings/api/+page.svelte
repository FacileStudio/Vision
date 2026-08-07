<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { APIKeyItem } from '$lib';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import {
		Alert,
		Badge,
		Button,
		ConfirmModal,
		Drawer,
		Field,
		Input,
		SecretField,
		Select,
		SettingsSection,
		Spinner,
		Table,
		icons,
		toast
	} from '@facile/muse';

	let keys = $state<APIKeyItem[]>([]);
	let loading = $state(true);

	let createOpen = $state(false);
	let creating = $state(false);
	let createdKey = $state('');
	let newName = $state('');
	let newScopes = $state('read');
	let createError = $state('');

	let revokeTarget = $state<APIKeyItem | null>(null);
	let revokeOpen = $state(false);

	const endpoint = $derived(`${page.url.origin}/api`);

	async function load() {
		loading = true;
		try {
			keys = await api.apiKeys.list(workspaceStore.current?.id);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Could not load your API keys.');
		} finally {
			loading = false;
		}
	}

	onMount(load);

	$effect(() => {
		workspaceStore.current;
		load();
	});

	/* Reset first: reopening the drawer must never re-show a key from a previous run. */
	function openCreate() {
		createdKey = '';
		newName = '';
		newScopes = 'read';
		createError = '';
		createOpen = true;
	}

	async function create(e: Event) {
		e.preventDefault();
		creating = true;
		createError = '';
		try {
			const resp = await api.apiKeys.create({
				name: newName.trim(),
				scopes: newScopes,
				workspace_id: workspaceStore.current?.id
			});
			createdKey = resp.key;
			newName = '';
			await load();
		} catch (e) {
			createError = e instanceof Error ? e.message : 'Could not create the key.';
		} finally {
			creating = false;
		}
	}

	async function revoke() {
		const target = revokeTarget;
		if (!target) return;
		await api.apiKeys.revoke(target.id);
		revokeTarget = null;
		toast.neutral('Key revoked.');
		await load();
	}

	function when(value: string | null) {
		return value ? new Date(value).toLocaleDateString() : 'Never';
	}
</script>

<div class="flex flex-col gap-10">
	<SettingsSection
		title="Endpoint"
		description="Point a script or an HTTP client here. Not a secret — the key is."
	>
		<SecretField value={endpoint} sensitive={false} label="Base URL" />
	</SettingsSection>

	<SettingsSection
		title="API keys"
		description="One key per machine. Revoking one never touches the others."
		bare
	>
		{#snippet actions()}
			<Button icon={icons.plus} onclick={openCreate}>New key</Button>
		{/snippet}

		{#if loading}
			<p class="text-fc-sm text-fc-fg-muted">Loading…</p>
		{:else if keys.length === 0}
			<Alert tone="info">
				No keys yet. A script needs one before it can read this space's analytics.
			</Alert>
		{:else}
			<Table>
				<thead>
					<tr>
						<th scope="col">Name</th>
						<th scope="col">Key</th>
						<th scope="col">Scope</th>
						<th scope="col">Last used</th>
						<th scope="col" class="text-right">Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each keys as key (key.id)}
						<!-- Revoked rows stay listed, dimmed, so the audit trail still names them. -->
						<tr class={key.is_active ? '' : 'opacity-55'}>
							<td class="font-medium text-fc-fg">{key.name}</td>
							<td class="whitespace-nowrap font-fc-mono text-fc-xs text-fc-fg-muted">
								{key.prefix}_****{key.key_hint}
							</td>
							<td>
								<Badge tone={key.scopes.includes('write') ? 'accent' : 'neutral'}>
									{key.scopes.includes('write') ? 'read + write' : 'read'}
								</Badge>
							</td>
							<td class="whitespace-nowrap text-fc-fg-muted">{when(key.last_used_at)}</td>
							<td class="text-right">
								{#if key.is_active}
									<Button
										variant="ghost-danger"
										size="sm"
										icon={icons.revoke}
										aria-label="Revoke {key.name}"
										onclick={() => {
											revokeTarget = key;
											revokeOpen = true;
										}}
									>
										Revoke
									</Button>
								{:else}
									<Badge>Revoked</Badge>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</Table>
		{/if}
	</SettingsSection>
</div>

<Drawer bind:open={createOpen} title="New API key" showClose>
	{#if createdKey}
		<div class="flex flex-col gap-4">
			<Alert tone="warning" title="Copy it now">
				This is the only time the key is shown. Vision stores a hash, so it cannot be shown to
				you again — losing it means issuing a new one.
			</Alert>

			<!-- The one-time key is the exception to the auto-hide rule: it starts revealed and
			     stays that way, because hiding a value nobody has copied yet is theatre. -->
			<SecretField
				value={createdKey}
				visible
				autoHideMs={0}
				label="Key"
				helper="Store it in your password manager or your CI secret store."
			/>
		</div>
	{:else}
		<form id="api-key-form" class="flex flex-col gap-4" onsubmit={create}>
			<Field label="Name" helper="Where the key will live — a machine, a pipeline, a script.">
				<Input bind:value={newName} placeholder="ci-dashboard" required disabled={creating} />
			</Field>

			<Field label="Scope">
				<Select bind:value={newScopes} disabled={creating}>
					<option value="read">Read — fetch analytics</option>
					<option value="read,write">Read and write — fetch and change data</option>
				</Select>
			</Field>

			{#if createError}
				<Alert tone="danger" title="Not created">{createError}</Alert>
			{/if}
		</form>
	{/if}

	{#snippet footer()}
		{#if createdKey}
			<Button size="lg" class="w-full" onclick={() => (createOpen = false)}>Done</Button>
		{:else}
			<div class="flex gap-2">
				<Button variant="ghost" size="lg" class="flex-1" disabled={creating} onclick={() => (createOpen = false)}>
					Cancel
				</Button>
				<Button
					type="submit"
					form="api-key-form"
					size="lg"
					icon={icons.key}
					class="flex-1"
					disabled={creating || !newName.trim()}
				>
					{#if creating}<Spinner size="sm" />{/if}
					{creating ? 'Creating…' : 'Create key'}
				</Button>
			</div>
		{/if}
	{/snippet}
</Drawer>

<ConfirmModal
	bind:open={revokeOpen}
	tone="danger"
	title="Revoke this key?"
	description={`"${revokeTarget?.name ?? ''}" stops working immediately, and anything still using it starts failing. It cannot be un-revoked — the row stays listed so the audit trail still names it.`}
	confirmLabel="Revoke key"
	cancelLabel="Keep it"
	onConfirm={revoke}
	onCancel={() => (revokeTarget = null)}
/>
