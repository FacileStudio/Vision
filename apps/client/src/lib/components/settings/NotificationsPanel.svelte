<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Webhook } from '$lib';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import { formatDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';
	import {
		Alert,
		Button,
		ConfirmModal,
		Drawer,
		Field,
		IconButton,
		Input,
		NotificationsSection,
		SecretField,
		Select,
		SettingsRow,
		Spinner,
		StatusDot,
		Switch,
		icons,
		toast
	} from '@facile/muse';

	let webhooks = $state<Webhook[]>([]);
	let loading = $state(true);

	let formOpen = $state(false);
	let editingId = $state<number | null>(null);
	let url = $state('');
	let secret = $state('');
	let preset = $state('24');
	let customHours = $state<string | number>(24);
	let saving = $state(false);
	let formError = $state('');

	let deleteTarget = $state<Webhook | null>(null);
	let deleteOpen = $state(false);

	const presets = $derived([
		{ hours: 1, label: t('webhooks.hourly') },
		{ hours: 6, label: t('webhooks.every6') },
		{ hours: 12, label: t('webhooks.every12') },
		{ hours: 24, label: t('webhooks.daily') },
		{ hours: 168, label: t('webhooks.weekly') },
		{ hours: 720, label: t('webhooks.monthly') }
	]);

	const presetHours = new Set(['1', '6', '12', '24', '168', '720']);

	const intervalHours = $derived(
		preset === 'custom' ? Math.max(1, Number(customHours) || 1) : Number(preset)
	);

	function frequencyLabel(hours: number): string {
		const match = presets.find((p) => p.hours === hours);
		if (match) return match.label;
		if (hours < 24) return t('webhooks.everyHours', { count: hours });
		const days = Math.round(hours / 24);
		return hours % 24 === 0
			? t('webhooks.everyDays', { count: days })
			: t('webhooks.everyHours', { count: hours });
	}

	async function load() {
		loading = true;
		try {
			webhooks = await api.webhooks.list(workspaceStore.current?.id);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('webhooks.loadFailed'));
		} finally {
			loading = false;
		}
	}

	onMount(load);

	$effect(() => {
		workspaceStore.current;
		load();
	});

	function openCreate() {
		editingId = null;
		url = '';
		secret = '';
		preset = '24';
		customHours = 24;
		formError = '';
		formOpen = true;
	}

	function openEdit(wh: Webhook) {
		editingId = wh.id;
		url = wh.url;
		/* The API never returns the stored secret, so the field starts empty and an empty
		   submit means "leave it alone" — same contract the toggle below relies on. */
		secret = '';
		const hours = String(wh.interval_hours);
		preset = presetHours.has(hours) ? hours : 'custom';
		customHours = wh.interval_hours;
		formError = '';
		formOpen = true;
	}

	async function save(e: Event) {
		e.preventDefault();
		saving = true;
		formError = '';
		try {
			if (editingId) {
				const existing = webhooks.find((w) => w.id === editingId);
				await api.webhooks.update(editingId, {
					url: url.trim(),
					secret,
					interval_hours: intervalHours,
					enabled: existing?.enabled ?? true
				});
			} else {
				await api.webhooks.create({
					url: url.trim(),
					secret,
					interval_hours: intervalHours,
					workspace_id: workspaceStore.current?.id
				});
			}
			formOpen = false;
			toast.success(editingId ? t('webhooks.updated') : t('webhooks.added'));
			await load();
		} catch (e) {
			formError = e instanceof Error ? e.message : t('webhooks.saveFailed');
		} finally {
			saving = false;
		}
	}

	async function toggleEnabled(wh: Webhook) {
		try {
			await api.webhooks.update(wh.id, {
				url: wh.url,
				secret: '',
				interval_hours: wh.interval_hours,
				enabled: !wh.enabled
			});
			await load();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('webhooks.toggleFailed'));
			await load();
		}
	}

	async function test(wh: Webhook) {
		try {
			await api.webhooks.test(wh.id);
			toast.success(t('webhooks.testSent'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('webhooks.testFailed'));
		}
	}

	async function remove() {
		const target = deleteTarget;
		if (!target) return;
		await api.webhooks.delete(target.id);
		deleteTarget = null;
		toast.neutral(t('webhooks.deleted'));
		await load();
	}
</script>

<div class="flex flex-col gap-10">
	<NotificationsSection title={t('webhooks.title')} description={t('webhooks.text')}>
		{#snippet actions()}
			<Button icon={icons.plus} onclick={openCreate}>{t('webhooks.new')}</Button>
		{/snippet}

		{#if loading}
			<p class="text-fc-sm text-fc-fg-muted">{t('common.loading')}</p>
		{:else if webhooks.length === 0}
			<Alert tone="info">{t('webhooks.empty')}</Alert>
		{:else}
			{#each webhooks as wh (wh.id)}
				<SettingsRow
					label={wh.url}
					description={wh.last_sent_at
						? t('webhooks.lastSent', {
								frequency: frequencyLabel(wh.interval_hours),
								date: formatDate(wh.last_sent_at, { dateStyle: 'medium' })
							})
						: t('webhooks.neverSent', { frequency: frequencyLabel(wh.interval_hours) })}
				>
					<StatusDot
						tone={wh.enabled ? 'success' : 'neutral'}
						label={wh.enabled ? t('webhooks.active') : t('webhooks.paused')}
					/>
					<Switch
						checked={wh.enabled}
						aria-label={t('webhooks.enableAria', { url: wh.url })}
						onchange={() => toggleEnabled(wh)}
					/>
					<IconButton variant="ghost" aria-label={t('webhooks.testAria', { url: wh.url })} onclick={() => test(wh)}>
						<iconify-icon icon={icons.mail} width="18" height="18" class="block"></iconify-icon>
					</IconButton>
					<IconButton variant="ghost" aria-label={t('webhooks.editAria', { url: wh.url })} onclick={() => openEdit(wh)}>
						<iconify-icon icon={icons.edit} width="18" height="18" class="block"></iconify-icon>
					</IconButton>
					<Button
						variant="ghost-danger"
						icon={icons.remove}
						aria-label={t('webhooks.deleteAria', { url: wh.url })}
						onclick={() => {
							deleteTarget = wh;
							deleteOpen = true;
						}}
					>
						{t('common.delete')}
					</Button>
				</SettingsRow>
			{/each}
		{/if}
	</NotificationsSection>
</div>

<Drawer
	bind:open={formOpen}
	title={editingId ? t('webhooks.editTitle') : t('webhooks.new')}
	description={t('webhooks.drawerText')}
	showClose
>
	<form id="webhook-form" class="flex flex-col gap-4" onsubmit={save}>
		<Field label={t('webhooks.url')} helper={t('webhooks.urlHelper')}>
			<Input bind:value={url} type="url" placeholder="https://example.com/hooks/vision" required disabled={saving} />
		</Field>

		<SecretField
			bind:value={secret}
			editable
			label={t('webhooks.secret')}
			helper={editingId ? t('webhooks.secretKeep') : t('webhooks.secretNew')}
			disabled={saving}
		/>

		<Field label={t('webhooks.frequency')}>
			<Select bind:value={preset} disabled={saving}>
				{#each presets as p (p.hours)}
					<option value={String(p.hours)}>{p.label}</option>
				{/each}
				<option value="custom">{t('webhooks.custom')}</option>
			</Select>
		</Field>

		{#if preset === 'custom'}
			<Field label={t('webhooks.everyInHours')} helper={frequencyLabel(intervalHours)}>
				<Input bind:value={customHours} type="number" min="1" max="8760" disabled={saving} />
			</Field>
		{/if}

		{#if formError}
			<Alert tone="danger" title={t('webhooks.notSaved')}>{formError}</Alert>
		{/if}
	</form>

	{#snippet footer()}
		<div class="flex gap-2">
			<Button variant="ghost" size="lg" class="flex-1" disabled={saving} onclick={() => (formOpen = false)}>
				{t('common.cancel')}
			</Button>
			<Button
				type="submit"
				form="webhook-form"
				size="lg"
				icon={icons.check}
				class="flex-1"
				disabled={saving || !url.trim()}
			>
				{#if saving}<Spinner size="sm" />{/if}
				{saving ? t('common.saving') : editingId ? t('webhooks.update') : t('webhooks.add')}
			</Button>
		</div>
	{/snippet}
</Drawer>

<ConfirmModal
	bind:open={deleteOpen}
	tone="danger"
	title={t('webhooks.deleteTitle')}
	description={t('webhooks.deleteBody', { url: deleteTarget?.url ?? t('webhooks.deleteIt') })}
	confirmLabel={t('webhooks.deleteConfirm')}
	cancelLabel={t('common.keepIt')}
	onConfirm={remove}
	onCancel={() => (deleteTarget = null)}
/>
