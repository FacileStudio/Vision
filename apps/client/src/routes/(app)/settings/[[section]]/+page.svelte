<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { PreferencesSection, ProfileSection, SettingsLayout, icons, settingsRoute } from '@facile/muse';
	import { userStore } from '$lib/stores/user.svelte';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import { theme } from '$lib/theme.svelte';
	import NotificationsPanel from '$lib/components/settings/NotificationsPanel.svelte';
	import TokensPanel from '$lib/components/settings/TokensPanel.svelte';

	const LEGACY: Record<string, string> = { appearance: 'preferences', api: 'tokens' };

	const SECTIONS = [
		{ id: 'profile', icon: icons.userCircle },
		{ id: 'preferences', icon: icons.palette },
		{ id: 'notifications', icon: icons.notification },
		{ id: 'tokens', icon: icons.key }
	] as const;

	const sections = $derived(
		SECTIONS.map((s) => ({ ...s, label: t(`settings.${s.id}`), href: `/settings/${s.id}` }))
	);

	const requested = $derived(page.params.section);
	const user = $derived(userStore.value);

	$effect(() => {
		const legacy = requested ? LEGACY[requested] : undefined;
		if (legacy) void goto(`/settings/${legacy}`, { replaceState: true });
		else if (requested && !settingsRoute(requested, sections).known) void goto('/settings', { replaceState: true });
	});
</script>

<svelte:head><title>{t('settings.pageTitle')}</title></svelte:head>

<SettingsLayout
	{sections}
	active={requested}
	backHref="/settings"
	title={t('settings.heading')}
	description={t('settings.subtitle')}
	navLabel={t('settings.sectionsLabel')}
>
	{#snippet children(id)}
		{#if id === 'profile'}
			<ProfileSection
				name={user?.name?.trim() || user?.email || t('nav.account')}
				email={user?.email}
				avatar={user?.avatar_url || undefined}
			/>
		{:else if id === 'preferences'}
			<PreferencesSection theme={theme.mode} onThemeChange={(mode) => theme.set(mode)} onLocaleChange={rememberLocale} />
		{:else if id === 'notifications'}
			<NotificationsPanel />
		{:else}
			<TokensPanel />
		{/if}
	{/snippet}
</SettingsLayout>
