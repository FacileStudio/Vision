<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Webhook } from '$lib';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import {
		Alert,
		Button,
		ConfirmModal,
		Drawer,
		Field,
		IconButton,
		Input,
		SecretField,
		Select,
		SettingsRow,
		SettingsSection,
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

	const presets = [
		{ hours: 1, label: 'Hourly' },
		{ hours: 6, label: 'Every 6 hours' },
		{ hours: 12, label: 'Every 12 hours' },
		{ hours: 24, label: 'Daily' },
		{ hours: 168, label: 'Weekly' },
		{ hours: 720, label: 'Monthly' }
	];

	const presetHours = new Set(presets.map((p) => String(p.hours)));

	const intervalHours = $derived(
		preset === 'custom' ? Math.max(1, Number(customHours) || 1) : Number(preset)
	);

	function frequencyLabel(hours: number): string {
		const match = presets.find((p) => p.hours === hours);
		if (match) return match.label;
		if (hours < 24) return `Every ${hours} hours`;
		const days = Math.round(hours / 24);
		return hours % 24 === 0 ? `Every ${days} days` : `Every ${hours} hours`;
	}

	async function load() {
		loading = true;
		try {
			webhooks = await api.webhooks.list(workspaceStore.current?.id);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Could not load your report webhooks.');
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
			toast.success(editingId ? 'Webhook updated.' : 'Webhook added.');
			await load();
		} catch (e) {
			formError = e instanceof Error ? e.message : 'Could not save the webhook.';
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
			toast.danger(e instanceof Error ? e.message : 'Could not change that webhook.');
			await load();
		}
	}

	async function test(wh: Webhook) {
		try {
			await api.webhooks.test(wh.id);
			toast.success('Test report sent.');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'The test delivery failed.');
		}
	}

	async function remove() {
		const target = deleteTarget;
		if (!target) return;
		await api.webhooks.delete(target.id);
		deleteTarget = null;
		toast.neutral('Webhook deleted.');
		await load();
	}
</script>

<div class="flex flex-col gap-10">
	<SettingsSection
		title="Report webhooks"
		description="Vision posts a signed analytics summary to each URL on its own schedule. Nothing else is emailed or sent."
	>
		{#snippet actions()}
			<Button icon={icons.plus} onclick={openCreate}>New webhook</Button>
		{/snippet}

		{#if loading}
			<p class="text-fc-sm text-fc-fg-muted">Loading…</p>
		{:else if webhooks.length === 0}
			<Alert tone="info">
				No webhooks yet. Add one to receive a periodic summary of every site in this space.
			</Alert>
		{:else}
			{#each webhooks as wh (wh.id)}
				<SettingsRow
					label={wh.url}
					description="{frequencyLabel(wh.interval_hours)} · {wh.last_sent_at
						? `last sent ${new Date(wh.last_sent_at).toLocaleDateString()}`
						: 'never sent'}"
				>
					<StatusDot
						tone={wh.enabled ? 'success' : 'neutral'}
						label={wh.enabled ? 'Active' : 'Paused'}
					/>
					<Switch
						checked={wh.enabled}
						aria-label="Enable {wh.url}"
						onchange={() => toggleEnabled(wh)}
					/>
					<IconButton variant="ghost" aria-label="Send a test report to {wh.url}" onclick={() => test(wh)}>
						<iconify-icon icon={icons.mail} width="18" height="18" class="block"></iconify-icon>
					</IconButton>
					<IconButton variant="ghost" aria-label="Edit {wh.url}" onclick={() => openEdit(wh)}>
						<iconify-icon icon={icons.edit} width="18" height="18" class="block"></iconify-icon>
					</IconButton>
					<Button
						variant="ghost-danger"
						icon={icons.remove}
						aria-label="Delete {wh.url}"
						onclick={() => {
							deleteTarget = wh;
							deleteOpen = true;
						}}
					>
						Delete
					</Button>
				</SettingsRow>
			{/each}
		{/if}
	</SettingsSection>
</div>

<Drawer
	bind:open={formOpen}
	title={editingId ? 'Edit webhook' : 'New webhook'}
	description="Reports are signed with the shared secret so the receiver can verify them."
	showClose
>
	<form id="webhook-form" class="flex flex-col gap-4" onsubmit={save}>
		<Field label="URL" helper="Where the report is POSTed.">
			<Input bind:value={url} type="url" placeholder="https://example.com/hooks/vision" required disabled={saving} />
		</Field>

		<SecretField
			bind:value={secret}
			editable
			label="Shared secret"
			helper={editingId
				? 'Leave empty to keep the secret you already set.'
				: 'Used to sign every delivery. Store it wherever the receiver reads it from.'}
			disabled={saving}
		/>

		<Field label="Frequency">
			<Select bind:value={preset} disabled={saving}>
				{#each presets as p (p.hours)}
					<option value={String(p.hours)}>{p.label}</option>
				{/each}
				<option value="custom">Custom…</option>
			</Select>
		</Field>

		{#if preset === 'custom'}
			<Field label="Every, in hours" helper={frequencyLabel(intervalHours)}>
				<Input bind:value={customHours} type="number" min="1" max="8760" disabled={saving} />
			</Field>
		{/if}

		{#if formError}
			<Alert tone="danger" title="Not saved">{formError}</Alert>
		{/if}
	</form>

	{#snippet footer()}
		<div class="flex gap-2">
			<Button variant="ghost" size="lg" class="flex-1" disabled={saving} onclick={() => (formOpen = false)}>
				Cancel
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
				{saving ? 'Saving…' : editingId ? 'Update' : 'Add webhook'}
			</Button>
		</div>
	{/snippet}
</Drawer>

<ConfirmModal
	bind:open={deleteOpen}
	tone="danger"
	title="Delete this webhook?"
	description={`Nothing is posted to ${deleteTarget?.url ?? 'it'} again, and the schedule is not recoverable. Pause it instead if you only want the reports to stop for a while.`}
	confirmLabel="Delete webhook"
	cancelLabel="Keep it"
	onConfirm={remove}
	onCancel={() => (deleteTarget = null)}
/>
