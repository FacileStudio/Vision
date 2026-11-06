<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Alert, AuthScreen } from '@facile/muse';
	import { currentUser } from '$lib';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';

	let failed = $state(false);
	let reason = $state('');

	onMount(async () => {
		failed = page.url.searchParams.has('error');
		reason = page.url.searchParams.get('error') ?? '';
		if (await currentUser().catch(() => null)) goto('/sites');
	});
</script>

<svelte:head>
	<title>{t('login.title')}</title>
</svelte:head>

<AuthScreen appName="Vision" href="/api/auth/oidc" onLocaleChange={rememberLocale}>
	{#snippet logo()}
		<iconify-icon icon="solar:panorama-bold-duotone" width="40" height="40" class="block text-fc-fg"
		></iconify-icon>
	{/snippet}
	{#snippet tagline()}
		{t('login.tagline')}
	{/snippet}
	{#snippet notice()}
		{#if failed}
			<Alert tone="danger" title={t('login.failedTitle')}>{reason || t('login.failed')}</Alert>
		{/if}
	{/snippet}
</AuthScreen>
