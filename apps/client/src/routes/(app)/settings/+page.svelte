<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { UserProfile, Webhook } from '$lib';
	import { Trash2, Pencil, Plus, Send } from '@lucide/svelte';

	let profile = $state<UserProfile | null>(null);
	let name = $state('');
	let email = $state('');
	let profileError = $state('');
	let profileSuccess = $state('');
	let profileLoading = $state(false);

	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordError = $state('');
	let passwordSuccess = $state('');
	let passwordLoading = $state(false);

	let webhooks = $state<Webhook[]>([]);
	let showWebhookForm = $state(false);
	let webhookUrl = $state('');
	let webhookSecret = $state('');
	let webhookPeriod = $state('daily');
	let editingWebhookId = $state<number | null>(null);
	let webhookSaving = $state(false);
	let deletingWebhookId = $state<number | null>(null);

	const webhookPeriods = ['hourly', 'daily', 'weekly', 'monthly'] as const;

	async function loadWebhooks() {
		try {
			webhooks = await api.webhooks.list();
		} catch {}
	}

	function resetWebhookForm() {
		webhookUrl = '';
		webhookSecret = '';
		webhookPeriod = 'daily';
		editingWebhookId = null;
		showWebhookForm = false;
	}

	function startEditWebhook(wh: Webhook) {
		webhookUrl = wh.url;
		webhookSecret = '';
		webhookPeriod = wh.period;
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
					period: webhookPeriod,
					enabled: existing?.enabled ?? true
				});
			} else {
				await api.webhooks.create({
					url: webhookUrl,
					secret: webhookSecret,
					period: webhookPeriod
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
				period: wh.period,
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
		try {
			profile = await api.auth.me();
			name = profile.name ?? '';
			email = profile.email;
		} catch (e: any) {
			profileError = e.message;
		}
		await loadWebhooks();
	});

	async function updateProfile() {
		profileError = '';
		profileSuccess = '';
		profileLoading = true;
		try {
			profile = await api.auth.updateProfile(name, email);
			name = profile.name ?? '';
			email = profile.email;
			profileSuccess = 'Profile updated.';
		} catch (e: any) {
			profileError = e.message;
		} finally {
			profileLoading = false;
		}
	}

	async function changePassword() {
		passwordError = '';
		passwordSuccess = '';

		if (newPassword !== confirmPassword) {
			passwordError = 'Passwords do not match.';
			return;
		}

		passwordLoading = true;
		try {
			await api.auth.changePassword(currentPassword, newPassword);
			passwordSuccess = 'Password changed.';
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
		} catch (e: any) {
			passwordError = e.message;
		} finally {
			passwordLoading = false;
		}
	}
</script>

<h1 class="text-2xl font-bold mb-6">Settings</h1>

<div class="space-y-8 max-w-lg">
	<div class="rounded-xl border bg-card p-4 backdrop-blur-sm">
		<h2 class="text-lg font-semibold mb-4">Profile</h2>

		{#if profile}
			<form onsubmit={updateProfile} class="space-y-3">
				<div>
					<label for="name" class="block text-sm font-medium mb-1">Name</label>
					<input
						id="name"
						bind:value={name}
						placeholder="Your name"
						class="w-full rounded-md border bg-background px-3 py-2 text-sm"
					/>
				</div>
				<div>
					<label for="email" class="block text-sm font-medium mb-1">Email</label>
					<input
						id="email"
						type="email"
						bind:value={email}
						placeholder="you@example.com"
						class="w-full rounded-md border bg-background px-3 py-2 text-sm"
						required
					/>
				</div>

				{#if profileError}
					<p class="text-destructive text-sm">{profileError}</p>
				{/if}
				{#if profileSuccess}
					<p class="text-sm text-green-600">{profileSuccess}</p>
				{/if}

				<button
					type="submit"
					disabled={profileLoading}
					class="rounded-md bg-primary px-4 py-2 text-sm text-primary-foreground disabled:opacity-50"
				>
					{profileLoading ? 'Saving...' : 'Save'}
				</button>
			</form>

			<p class="text-xs text-muted-foreground mt-4">
				Member since {new Date(profile.created_at).toLocaleDateString()}
			</p>
		{:else if !profileError}
			<p class="text-sm text-muted-foreground">Loading...</p>
		{/if}
	</div>

	<div class="rounded-xl border bg-card p-4 backdrop-blur-sm">
		<h2 class="text-lg font-semibold mb-4">Change Password</h2>

		<form onsubmit={changePassword} class="space-y-3">
			<div>
				<label for="current-password" class="block text-sm font-medium mb-1">Current Password</label>
				<input
					id="current-password"
					type="password"
					bind:value={currentPassword}
					class="w-full rounded-md border bg-background px-3 py-2 text-sm"
					required
				/>
			</div>
			<div>
				<label for="new-password" class="block text-sm font-medium mb-1">New Password</label>
				<input
					id="new-password"
					type="password"
					bind:value={newPassword}
					class="w-full rounded-md border bg-background px-3 py-2 text-sm"
					required
					minlength="8"
				/>
			</div>
			<div>
				<label for="confirm-password" class="block text-sm font-medium mb-1">Confirm New Password</label>
				<input
					id="confirm-password"
					type="password"
					bind:value={confirmPassword}
					class="w-full rounded-md border bg-background px-3 py-2 text-sm"
					required
					minlength="8"
				/>
			</div>

			{#if passwordError}
				<p class="text-destructive text-sm">{passwordError}</p>
			{/if}
			{#if passwordSuccess}
				<p class="text-sm text-green-600">{passwordSuccess}</p>
			{/if}

			<button
				type="submit"
				disabled={passwordLoading}
				class="rounded-md bg-primary px-4 py-2 text-sm text-primary-foreground disabled:opacity-50"
			>
				{passwordLoading ? 'Changing...' : 'Change Password'}
			</button>
		</form>
	</div>

	<div class="rounded-xl border bg-card p-6 backdrop-blur-sm">
		<div class="flex items-center justify-between mb-1">
			<h2 class="text-lg font-semibold">Webhook Reports</h2>
			{#if !showWebhookForm && webhooks.length > 0}
				<button
					onclick={() => { resetWebhookForm(); showWebhookForm = true; }}
					class="flex items-center gap-1.5 rounded-full bg-foreground px-3 py-1 text-sm font-medium text-background transition-colors hover:bg-foreground/90"
				>
					<Plus class="h-3.5 w-3.5" />
					Add
				</button>
			{/if}
		</div>
		<p class="text-sm text-muted-foreground mb-4">Send periodic analytics reports for all your sites to external services</p>

		{#if webhooks.length === 0 && !showWebhookForm}
			<button
				onclick={() => (showWebhookForm = true)}
				class="flex items-center gap-2 rounded-lg border border-dashed px-4 py-3 text-sm text-muted-foreground transition-colors hover:border-foreground/30 hover:text-foreground w-full justify-center"
			>
				<Plus class="h-4 w-4" />
				Add Webhook
			</button>
		{/if}

		{#if webhooks.length > 0}
			<div class="space-y-3 mb-4">
				{#each webhooks as wh}
					<div class="rounded-xl border bg-card p-4 backdrop-blur-sm">
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
										<span class="rounded-full px-2.5 py-0.5 text-xs font-medium {wh.enabled ? 'bg-foreground text-background' : 'bg-muted text-muted-foreground'}">
											{wh.period}
										</span>
										<span class="text-xs text-muted-foreground">
											{wh.last_sent_at ? `Last sent ${new Date(wh.last_sent_at).toLocaleDateString()}` : 'Never sent'}
										</span>
									</div>
								</div>
								<div class="flex items-center gap-1.5 shrink-0">
									<button
										onclick={() => toggleWebhookEnabled(wh)}
										class="relative h-6 w-10 rounded-full transition-colors {wh.enabled ? 'bg-green-500' : 'bg-muted'}"
										aria-label="{wh.enabled ? 'Disable' : 'Enable'} webhook"
									>
										<span class="absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform {wh.enabled ? 'left-[18px]' : 'left-0.5'}"></span>
									</button>
									<button
										onclick={() => testWebhook(wh.id)}
										class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-foreground hover:bg-muted"
										aria-label="Test webhook"
									>
										<Send class="h-3.5 w-3.5" />
									</button>
									<button
										onclick={() => startEditWebhook(wh)}
										class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-foreground hover:bg-muted"
										aria-label="Edit webhook"
									>
										<Pencil class="h-3.5 w-3.5" />
									</button>
									<button
										onclick={() => (deletingWebhookId = wh.id)}
										class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-red-500 hover:bg-muted"
										aria-label="Delete webhook"
									>
										<Trash2 class="h-3.5 w-3.5" />
									</button>
								</div>
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}

		{#if showWebhookForm}
			<div class="rounded-xl border bg-card p-4 backdrop-blur-sm space-y-4">
				<div>
					<label for="webhook-url" class="block text-sm font-medium mb-1">URL</label>
					<input
						id="webhook-url"
						type="text"
						bind:value={webhookUrl}
						placeholder="https://nook.example.com/webhook/vision"
						class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>
				<div>
					<label for="webhook-secret" class="block text-sm font-medium mb-1">Secret</label>
					<input
						id="webhook-secret"
						type="password"
						bind:value={webhookSecret}
						placeholder="Shared HMAC secret"
						class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium mb-1.5">Period</label>
					<div class="flex gap-2">
						{#each webhookPeriods as p}
							<button
								onclick={() => (webhookPeriod = p)}
								class="rounded-full px-3 py-1 text-sm font-medium transition-colors {webhookPeriod === p
									? 'bg-foreground text-background'
									: 'bg-muted text-muted-foreground hover:text-foreground'}"
							>
								{p[0].toUpperCase() + p.slice(1)}
							</button>
						{/each}
					</div>
				</div>
				<div class="flex gap-2 pt-1">
					<button
						onclick={saveWebhook}
						disabled={webhookSaving || !webhookUrl}
						class="rounded-full bg-foreground px-4 py-1.5 text-sm font-medium text-background transition-colors hover:bg-foreground/90 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						{webhookSaving ? 'Saving…' : editingWebhookId ? 'Update' : 'Save'}
					</button>
					<button
						onclick={resetWebhookForm}
						class="rounded-full bg-muted px-4 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
					>
						Cancel
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
