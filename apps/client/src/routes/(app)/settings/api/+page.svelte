<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { APIKeyItem } from '$lib';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import { formatDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';
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
			toast.danger(e instanceof Error ? e.message : t('apiKeys.loadFailed'));
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
			createError = e instanceof Error ? e.message : t('apiKeys.createFailed');
		} finally {
			creating = false;
		}
	}

	async function revoke() {
		const target = revokeTarget;
		if (!target) return;
		await api.apiKeys.revoke(target.id);
		revokeTarget = null;
		toast.neutral(t('apiKeys.revoked'));
		await load();
	}

	function when(value: string | null) {
		return value ? formatDate(value, { dateStyle: 'medium' }) : t('common.never');
	}
</script>

<div class="flex flex-col gap-10">
	<SettingsSection title={t('apiKeys.endpointTitle')} description={t('apiKeys.endpointText')}>
		<SecretField value={endpoint} sensitive={false} label={t('apiKeys.baseUrl')} />
	</SettingsSection>

	<SettingsSection title={t('apiKeys.title')} description={t('apiKeys.text')} bare>
		{#snippet actions()}
			<Button icon={icons.plus} onclick={openCreate}>{t('apiKeys.new')}</Button>
		{/snippet}

		{#if loading}
			<p class="text-fc-sm text-fc-fg-muted">{t('common.loading')}</p>
		{:else if keys.length === 0}
			<Alert tone="info">{t('apiKeys.empty')}</Alert>
		{:else}
			<Table>
				<thead>
					<tr>
						<th scope="col">{t('apiKeys.colName')}</th>
						<th scope="col">{t('apiKeys.colKey')}</th>
						<th scope="col">{t('apiKeys.colScope')}</th>
						<th scope="col">{t('apiKeys.colLastUsed')}</th>
						<th scope="col" class="text-right">{t('apiKeys.colActions')}</th>
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
									{key.scopes.includes('write') ? t('apiKeys.scopeWrite') : t('apiKeys.scopeRead')}
								</Badge>
							</td>
							<td class="whitespace-nowrap text-fc-fg-muted">{when(key.last_used_at)}</td>
							<td class="text-right">
								{#if key.is_active}
									<Button
										variant="ghost-danger"
										size="sm"
										icon={icons.revoke}
										aria-label={t('apiKeys.revokeAria', { name: key.name })}
										onclick={() => {
											revokeTarget = key;
											revokeOpen = true;
										}}
									>
										{t('apiKeys.revoke')}
									</Button>
								{:else}
									<Badge>{t('apiKeys.revokedBadge')}</Badge>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</Table>
		{/if}
	</SettingsSection>
</div>

<Drawer bind:open={createOpen} title={t('apiKeys.drawerTitle')} showClose>
	{#if createdKey}
		<div class="flex flex-col gap-4">
			<Alert tone="warning" title={t('apiKeys.copyTitle')}>{t('apiKeys.copyText')}</Alert>

			<!-- The one-time key is the exception to the auto-hide rule: it starts revealed and
			     stays that way, because hiding a value nobody has copied yet is theatre. -->
			<SecretField
				value={createdKey}
				visible
				autoHideMs={0}
				label={t('apiKeys.colKey')}
				helper={t('apiKeys.secretHelper')}
			/>
		</div>
	{:else}
		<form id="api-key-form" class="flex flex-col gap-4" onsubmit={create}>
			<Field label={t('common.name')} helper={t('apiKeys.nameHelper')}>
				<Input bind:value={newName} placeholder="ci-dashboard" required disabled={creating} />
			</Field>

			<Field label={t('apiKeys.scope')}>
				<Select bind:value={newScopes} disabled={creating}>
					<option value="read">{t('apiKeys.scopeReadOption')}</option>
					<option value="read,write">{t('apiKeys.scopeWriteOption')}</option>
				</Select>
			</Field>

			{#if createError}
				<Alert tone="danger" title={t('apiKeys.notCreated')}>{createError}</Alert>
			{/if}
		</form>
	{/if}

	{#snippet footer()}
		{#if createdKey}
			<Button size="lg" class="w-full" onclick={() => (createOpen = false)}>{t('common.done')}</Button>
		{:else}
			<div class="flex gap-2">
				<Button variant="ghost" size="lg" class="flex-1" disabled={creating} onclick={() => (createOpen = false)}>
					{t('common.cancel')}
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
					{creating ? t('common.creating') : t('apiKeys.submit')}
				</Button>
			</div>
		{/if}
	{/snippet}
</Drawer>

<ConfirmModal
	bind:open={revokeOpen}
	tone="danger"
	title={t('apiKeys.revokeTitle')}
	description={t('apiKeys.revokeBody', { name: revokeTarget?.name ?? '' })}
	confirmLabel={t('apiKeys.revokeConfirm')}
	cancelLabel={t('common.keepIt')}
	onConfirm={revoke}
	onCancel={() => (revokeTarget = null)}
/>
