<script lang="ts">
	import { page } from '$app/state';
	import { Divider, PageTransition, Tabs, icons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let { children } = $props();

	/*
	 * The section lives in the URL, so /settings/api opens on API, reload keeps you there
	 * and browser-back walks the sections. That is why Tabs items carry `href`.
	 */
	const sections = $derived([
		{ id: 'profile', label: t('settings.profile'), icon: icons.userCircle, href: '/settings' },
		{
			id: 'appearance',
			label: t('settings.appearance'),
			icon: icons.palette,
			href: '/settings/appearance'
		},
		{
			id: 'notifications',
			label: t('settings.notifications'),
			icon: icons.notification,
			href: '/settings/notifications'
		},
		{ id: 'api', label: t('settings.api'), icon: icons.key, href: '/settings/api' }
	]);

	const active = $derived(
		sections.find((s) => s.href !== '/settings' && page.url.pathname.startsWith(s.href))?.id ??
			'profile'
	);
</script>

<svelte:head><title>{t('settings.pageTitle')}</title></svelte:head>

<div class="flex flex-col gap-8">
	<div class="flex flex-col gap-2">
		<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('settings.heading')}</h1>
		<p class="text-fc-sm text-fc-fg-muted">{t('settings.subtitle')}</p>
	</div>

	<!-- The rule needs air: pulled tight under the strip it reads as an underline welded to
	     the active pill. -->
	<div class="flex flex-col gap-4">
		<Tabs items={sections} value={active} label={t('settings.sectionsLabel')} />
		<Divider class="my-0" />
	</div>

	<PageTransition key={active} distance={8} duration={0.25}>
		{@render children()}
	</PageTransition>
</div>
