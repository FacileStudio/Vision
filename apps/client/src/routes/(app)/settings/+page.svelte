<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Webhook } from '$lib';
	import Icon from '@iconify/svelte';
	import * as ToggleGroup from '$lib/components/ui/toggle-group/index.js';

	let webhooks = $state<Webhook[]>([]);
	let showWebhookForm = $state(false);
	let webhookUrl = $state('');
	let webhookSecret = $state('');
	let webhookIntervalHours = $state(24);
	let webhookCustomHours = $state('');
	let webhookFreqMode = $state<'preset' | 'custom'>('preset');
	let editingWebhookId = $state<number | null>(null);
	let webhookSaving = $state(false);
	let deletingWebhookId = $state<number | null>(null);

	const frequencyPresets = [
		{ hours: 1, label: 'Hourly' },
		{ hours: 6, label: '6h' },
		{ hours: 12, label: '12h' },
		{ hours: 24, label: 'Daily' },
		{ hours: 168, label: 'Weekly' },
		{ hours: 720, label: 'Monthly' }
	];

	const presetHoursSet = new Set(frequencyPresets.map((p) => p.hours));

	function frequencyLabel(hours: number): string {
		const preset = frequencyPresets.find((p) => p.hours === hours);
		if (preset) return preset.label;
		if (hours < 24) return `Every ${hours}h`;
		const days = Math.floor(hours / 24);
		if (hours % 24 === 0) return days === 1 ? 'Daily' : `Every ${days}d`;
		return `Every ${hours}h`;
	}

	function setPresetFrequency(hours: string) {
		webhookFreqMode = 'preset';
		webhookIntervalHours = Number(hours);
		webhookCustomHours = '';
	}

	function setCustomFrequency() {
		webhookFreqMode = 'custom';
		webhookCustomHours = String(webhookIntervalHours);
	}

	let effectiveIntervalHours = $derived(
		webhookFreqMode === 'custom' ? Number(webhookCustomHours) || 1 : webhookIntervalHours
	);

	async function loadWebhooks() {
		try {
			webhooks = await api.webhooks.list();
		} catch {}
	}

	function resetWebhookForm() {
		webhookUrl = '';
		webhookSecret = '';
		webhookIntervalHours = 24;
		webhookCustomHours = '';
		webhookFreqMode = 'preset';
		editingWebhookId = null;
		showWebhookForm = false;
	}

	function startEditWebhook(wh: Webhook) {
		webhookUrl = wh.url;
		webhookSecret = '';
		const hours = wh.interval_hours;
		if (presetHoursSet.has(hours)) {
			webhookFreqMode = 'preset';
			webhookIntervalHours = hours;
			webhookCustomHours = '';
		} else {
			webhookFreqMode = 'custom';
			webhookIntervalHours = hours;
			webhookCustomHours = String(hours);
		}
		editingWebhookId = wh.id;
		showWebhookForm = true;
	}

	async function saveWebhook() {
		webhookSaving = true;
		try {
			if (editingWebhookId) {
				const existing = webhooks.find((w) => w.id === editingWebhookId);
				await api.webhooks.update(editingWebhookId, {
					url: webhookUrl,
					secret: webhookSecret,
					interval_hours: effectiveIntervalHours,
					enabled: existing?.enabled ?? true
				});
			} else {
				await api.webhooks.create({
					url: webhookUrl,
					secret: webhookSecret,
					interval_hours: effectiveIntervalHours
				});
			}
			resetWebhookForm();
			await loadWebhooks();
		} catch {}
		webhookSaving = false;
	}

	async function toggleWebhookEnabled(wh: Webhook) {
		try {
			await api.webhooks.update(wh.id, {
				url: wh.url,
				secret: '',
				interval_hours: wh.interval_hours,
				enabled: !wh.enabled
			});
			await loadWebhooks();
		} catch {}
	}

	async function deleteWebhook(id: number) {
		try {
			await api.webhooks.delete(id);
			deletingWebhookId = null;
			await loadWebhooks();
		} catch {}
	}

	async function testWebhook(id: number) {
		try {
			await api.webhooks.test(id);
		} catch {}
	}

	onMount(async () => {
		await loadWebhooks();
	});
</script>

<svelte:head><title>Settings — Vision</title></svelte:head>

<h1 class="mb-6 text-2xl font-bold">Settings</h1>

<div class="max-w-lg space-y-8">
	<div class="rounded-lg border p-6">
		<div class="mb-1 flex items-center justify-between">
			<h2 class="text-lg font-semibold">Webhook Reports</h2>
			{#if !showWebhookForm && webhooks.length > 0}
				<button
					onclick={() => {
						resetWebhookForm();
						showWebhookForm = true;
					}}
					class="flex items-center gap-1.5 rounded-full bg-foreground px-3 py-1 text-sm font-medium text-background transition-colors hover:bg-foreground/90"
				>
					<Icon icon="solar:add-circle-linear" class="h-4 w-4" />
					Add
				</button>
			{/if}
		</div>
		<p class="mb-4 text-sm text-muted-foreground">
			Send periodic analytics reports for all your sites to external services
		</p>

		{#if webhooks.length === 0 && !showWebhookForm}
			<button
				onclick={() => (showWebhookForm = true)}
				class="flex w-full items-center justify-center gap-2 rounded-lg border border-dashed px-4 py-3 text-sm text-muted-foreground transition-colors hover:border-foreground/30 hover:text-foreground"
			>
				<Icon icon="solar:add-circle-linear" class="h-4 w-4" />
				Add Webhook
			</button>
		{/if}

		{#if webhooks.length > 0}
			<div class="mb-4 space-y-3">
				{#each webhooks as wh}
					<div class="rounded-lg border p-4">
						{#if deletingWebhookId === wh.id}
							<div class="flex items-center justify-between">
								<p class="text-sm">Delete this webhook?</p>
								<div class="flex gap-2">
									<button
										onclick={() => deleteWebhook(wh.id)}
										class="rounded-full bg-red-500 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-red-600"
									>
										Delete
									</button>
									<button
										onclick={() => (deletingWebhookId = null)}
										class="rounded-full bg-muted px-3 py-1 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
									>
										Cancel
									</button>
								</div>
							</div>
						{:else}
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-medium">{wh.url}</p>
									<div class="mt-1.5 flex flex-wrap items-center gap-2">
										<span
											class="rounded-full px-2.5 py-0.5 text-xs font-medium {wh.enabled
												? 'bg-foreground text-background'
												: 'bg-muted text-muted-foreground'}"
										>
											{wh.period}
										</span>
										<span class="text-xs text-muted-foreground">
											{wh.last_sent_at
												? `Last sent ${new Date(wh.last_sent_at).toLocaleDateString()}`
												: 'Never sent'}
										</span>
									</div>
								</div>
								<div class="flex shrink-0 items-center gap-1.5">
									<button
										onclick={() => toggleWebhookEnabled(wh)}
										class="relative h-6 w-10 rounded-full transition-colors {wh.enabled
											? 'bg-green-500'
											: 'bg-muted'}"
										aria-label="{wh.enabled ? 'Disable' : 'Enable'} webhook"
									>
										<span
											class="absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform {wh.enabled
												? 'left-[18px]'
												: 'left-0.5'}"
										></span>
									</button>
									<button
										onclick={() => testWebhook(wh.id)}
										class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
										aria-label="Test webhook"
									>
										<Icon icon="solar:plain-linear" class="h-3.5 w-3.5" />
									</button>
									<button
										onclick={() => startEditWebhook(wh)}
										class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
										aria-label="Edit webhook"
									>
										<Icon icon="solar:pen-linear" class="h-3.5 w-3.5" />
									</button>
									<button
										onclick={() => (deletingWebhookId = wh.id)}
										class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-red-500"
										aria-label="Delete webhook"
									>
										<Icon icon="solar:trash-bin-trash-linear" class="h-3.5 w-3.5" />
									</button>
								</div>
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}

		{#if showWebhookForm}
			<div class="space-y-4 rounded-lg border p-4">
				<div>
					<label for="webhook-url" class="mb-1 block text-sm font-medium">URL</label>
					<input
						id="webhook-url"
						type="text"
						bind:value={webhookUrl}
						placeholder="https://example.com/webhook/vision"
						class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>
				<div>
					<label for="webhook-secret" class="mb-1 block text-sm font-medium">Secret</label>
					<input
						id="webhook-secret"
						type="password"
						bind:value={webhookSecret}
						placeholder="Shared HMAC secret"
						class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>
				<div>
					<span class="mb-1.5 block text-sm font-medium">Frequency</span>
					<div class="flex flex-wrap gap-2">
						{#each frequencyPresets as p}
							<button
								onclick={() => setPresetFrequency(String(p.hours))}
								class="rounded-full px-3 py-1 text-sm font-medium transition-colors {webhookFreqMode ===
									'preset' && webhookIntervalHours === p.hours
									? 'bg-foreground text-background'
									: 'bg-muted text-muted-foreground hover:text-foreground'}"
							>
								{p.label}
							</button>
						{/each}
						<button
							onclick={setCustomFrequency}
							class="rounded-full px-3 py-1 text-sm font-medium transition-colors {webhookFreqMode ===
							'custom'
								? 'bg-foreground text-background'
								: 'bg-muted text-muted-foreground hover:text-foreground'}"
						>
							Custom
						</button>
					</div>
					{#if webhookFreqMode === 'custom'}
						<div class="mt-3 flex items-center gap-2">
							<span class="text-sm text-muted-foreground">Every</span>
							<input
								type="number"
								min="1"
								max="8760"
								bind:value={webhookCustomHours}
								class="w-20 rounded-md border bg-background px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
							/>
							<span class="text-sm text-muted-foreground">hours</span>
							{#if Number(webhookCustomHours) >= 24}
								<span class="text-xs text-muted-foreground">
									({frequencyLabel(Number(webhookCustomHours))})
								</span>
							{/if}
						</div>
					{/if}
				</div>
				<div class="flex gap-2 pt-1">
					<button
						onclick={saveWebhook}
						disabled={webhookSaving || !webhookUrl || effectiveIntervalHours < 1}
						class="flex items-center gap-1.5 rounded-full bg-foreground px-4 py-1.5 text-sm font-medium text-background transition-colors hover:bg-foreground/90 disabled:cursor-not-allowed disabled:opacity-50"
					>
						<Icon icon="solar:diskette-linear" class="h-4 w-4" />
						{webhookSaving ? 'Saving…' : editingWebhookId ? 'Update' : 'Save'}
					</button>
					<button
						onclick={resetWebhookForm}
						class="flex items-center gap-1.5 rounded-full bg-muted px-4 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
					>
						<Icon icon="solar:close-circle-linear" class="h-4 w-4" />
						Cancel
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
