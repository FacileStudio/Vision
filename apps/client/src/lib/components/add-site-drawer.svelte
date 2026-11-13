<script lang="ts">
	import { api } from '$lib';
	import type { Site } from '$lib';
	import { Alert, Button, Drawer, Field, Input, Spinner, icons } from '@facile/muse';
	import SpaceSelect from '$lib/components/space-select.svelte';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let {
		open = $bindable(false),
		space = '',
		onCreated
	}: {
		open?: boolean;
		space?: string;
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
			workspaceId = '';
		}
	});

	async function addSite(e: Event) {
		e.preventDefault();
		error = '';
		const id = Number(workspaceId);
		if (!id) {
			error = t('addSite.noSpace');
			return;
		}
		submitting = true;
		try {
			const site = await api.sites.create(name.trim(), domain.trim(), id);
			open = false;
			onCreated(site);
		} catch (e) {
			error = e instanceof Error ? e.message : t('addSite.failed');
		} finally {
			submitting = false;
		}
	}
</script>

<Drawer bind:open title={t('addSite.title')} description={t('addSite.description')} showClose>
	<form id="add-site" class="flex flex-col gap-4" onsubmit={addSite}>
		<div class={workspaceStore.writable.length > 1 ? 'contents' : 'hidden'}>
			<Field label={t('addSite.space')} helper={t('addSite.spaceHelper')}>
				<SpaceSelect bind:value={workspaceId} preferred={space} disabled={submitting} />
			</Field>
		</div>

		<Field label={t('common.name')} helper={t('addSite.nameHelper')}>
			<Input bind:value={name} placeholder={t('addSite.namePlaceholder')} required disabled={submitting} />
		</Field>

		<Field label={t('addSite.domain')} helper={t('addSite.domainHelper')}>
			<Input bind:value={domain} placeholder="example.com" required disabled={submitting} />
		</Field>

		{#if error}
			<Alert tone="danger" title={t('addSite.errorTitle')}>{error}</Alert>
		{/if}
	</form>

	{#snippet footer()}
		<div class="flex gap-2">
			<Button variant="ghost" size="lg" class="flex-1" disabled={submitting} onclick={() => (open = false)}>
				{t('common.cancel')}
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
				{submitting ? t('common.adding') : t('sites.add')}
			</Button>
		</div>
	{/snippet}
</Drawer>
