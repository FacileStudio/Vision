<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { Alert, Button, Card, EmptyState, Skeleton, icons, toast } from '@facile/muse';
	import { ApiError, api, notOwnSession, type PendingAction } from '$lib/backend';
	import PendingCard from '$lib/components/pending-card.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let actions = $state<PendingAction[]>([]);
	let loading = $state(true);
	let busy = $state('');
	let shared = $state('');
	let signInAgain = $state(false);

	const focused = $derived(page.url.searchParams.get('action') ?? '');
	const ordered = $derived([
		...actions.filter((action) => action.id === focused),
		...actions.filter((action) => action.id !== focused)
	]);

	onMount(load);

	/** Keeps this page for the tab so the layout reopens it once the new session is back. */
	function keepReturn() {
		try {
			sessionStorage.setItem('vision.return', page.url.pathname + page.url.search);
		} catch {}
	}

	/** Reports whether the API refused this session for confirming, and shows the sign-in prompt if so. */
	function refusedSession(err: unknown): boolean {
		signInAgain = notOwnSession(err);
		return signInAgain;
	}

	/** The server's own words for any other refusal, else the fallback text. */
	function reason(err: unknown, fallback: string): string {
		return err instanceof ApiError && err.status === 403 ? err.message : fallback;
	}

	/** Loads the open actions. */
	async function load() {
		loading = true;
		try {
			actions = (await api.pending.list()).actions;
		} catch (err) {
			if (!refusedSession(err)) toast.danger(reason(err, t('pending.loadFailed')));
		}
		loading = false;
	}

	/** Runs one action once, then drops it from the list; a public link is shown to copy. */
	async function confirm(action: PendingAction) {
		busy = action.id;
		try {
			const result = await api.pending.confirm(action.id);
			actions = actions.filter((item) => item.id !== action.id);
			if (result.url) shared = result.url;
			else toast.success(t('pending.done.deleted'));
		} catch (err) {
			if (!refusedSession(err)) {
				toast.danger(err instanceof Error ? err.message : t('pending.failed'));
				await load();
			}
		}
		busy = '';
	}

	/** Closes one action without running it. */
	async function dismiss(action: PendingAction) {
		busy = action.id;
		try {
			await api.pending.dismiss(action.id);
			toast.success(t('pending.dismissed'));
			actions = actions.filter((item) => item.id !== action.id);
		} catch (err) {
			if (!refusedSession(err)) toast.danger(err instanceof Error ? err.message : t('pending.failed'));
		}
		busy = '';
	}

	/** Copies the public link the person just opened. */
	async function copyShared() {
		try {
			await navigator.clipboard.writeText(shared);
			toast.success(t('pending.copied'));
		} catch {
			toast.danger(t('pending.failed'));
		}
	}
</script>

<svelte:head>
	<title>{t('pending.title')}</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-2xl flex-col gap-6">
	<div class="flex flex-col gap-2">
		<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('pending.heading')}</h1>
		<p class="text-fc-sm text-fc-fg-muted">{t('pending.intro')}</p>
	</div>

	{#if shared}
		<Card class="flex flex-col gap-3">
			<p class="text-fc-sm font-medium text-fc-fg">{t('pending.done.shared')}</p>
			<p class="break-all rounded-fc-md bg-fc-surface p-3 font-mono text-fc-xs text-fc-fg">{shared}</p>
			<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
				<Button variant="ghost" class="w-full sm:w-auto [@media(pointer:coarse)]:h-11" onclick={() => (shared = '')}>
					{t('pending.close')}
				</Button>
				<Button class="w-full sm:w-auto [@media(pointer:coarse)]:h-11" icon={icons.copy} onclick={copyShared}>
					{t('pending.copy')}
				</Button>
			</div>
		</Card>
	{/if}

	{#if signInAgain}
		<Alert tone="warning" title={t('pending.signInAgain.title')}>
			<p>{t('pending.signInAgain.body')}</p>
			<Button href="/api/auth/oidc" onclick={keepReturn} class="mt-3 w-full sm:w-auto [@media(pointer:coarse)]:h-11">
				{t('pending.signInAgain.action')}
			</Button>
		</Alert>
	{:else if loading}
		{#each [0, 1] as row (row)}
			<Skeleton class="h-40 w-full rounded-fc-md" />
		{/each}
	{:else if actions.length === 0}
		<EmptyState icon={icons.check} title={t('pending.none')} description={t('pending.noneBody')} />
	{:else}
		<section class="flex flex-col gap-4">
			{#each ordered as action (action.id)}
				<PendingCard {action} busy={busy === action.id} onconfirm={confirm} ondismiss={dismiss} />
			{/each}
		</section>
	{/if}
</div>
