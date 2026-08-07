<script lang="ts">
	import { page } from '$app/state';
	import { Divider, PageTransition, Tabs, icons } from '@facile/muse';

	let { children } = $props();

	/*
	 * The section lives in the URL, so /settings/api opens on API, reload keeps you there
	 * and browser-back walks the sections. That is why Tabs items carry `href`.
	 */
	const sections = [
		{ id: 'profile', label: 'Profile', icon: icons.userCircle, href: '/settings' },
		{ id: 'appearance', label: 'Appearance', icon: icons.palette, href: '/settings/appearance' },
		{
			id: 'notifications',
			label: 'Notifications',
			icon: icons.notification,
			href: '/settings/notifications'
		},
		{ id: 'api', label: 'API', icon: icons.key, href: '/settings/api' }
	];

	const active = $derived(
		sections.find((s) => s.href !== '/settings' && page.url.pathname.startsWith(s.href))?.id ??
			'profile'
	);
</script>

<svelte:head><title>Settings — Vision</title></svelte:head>

<div class="flex flex-col gap-8">
	<div class="flex flex-col gap-2">
		<h1 class="text-fc-2xl font-semibold text-fc-fg">Settings</h1>
		<p class="text-fc-sm text-fc-fg-muted">Your account, this browser, and everything wired to your spaces.</p>
	</div>

	<!-- The rule needs air: pulled tight under the strip it reads as an underline welded to
	     the active pill. -->
	<div class="flex flex-col gap-4">
		<Tabs items={sections} value={active} label="Settings sections" />
		<Divider class="my-0" />
	</div>

	<PageTransition key={active} distance={8} duration={0.25}>
		{@render children()}
	</PageTransition>
</div>
