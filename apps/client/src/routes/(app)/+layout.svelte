<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api, currentUser } from '$lib';
	import {
		MobileNav,
		PageTransition,
		SideBar,
		SpaceSwitcher,
		Topbar,
		icons
	} from '@facile/muse';
	import { userStore } from '$lib/stores/user.svelte';
	import { workspaceStore } from '$lib/stores/workspace.svelte';

	let { children } = $props();

	let collapsed = $state(false);
	let scroller: HTMLElement | null = $state(null);

	onMount(async () => {
		/* The API is the only thing that knows whether this browser has a session:
		   a local login leaves a bearer token, an SSO login leaves a cookie the
		   client cannot read. A thrown error is a bad round-trip, not a logout —
		   bouncing to /login on one would sign people out on a hiccup. */
		let me;
		try {
			me = await currentUser();
		} catch {
			return;
		}
		if (!me) {
			goto('/login');
			return;
		}
		userStore.value = me;
		try {
			workspaceStore.hydrate(await api.workspaces.list());
		} catch {}
		api.auth
			.syncProfile()
			.then(async () => {
				try {
					userStore.value = await api.auth.me();
				} catch {}
			})
			.catch(() => {});
	});

	/* <main> is the scroll container and sits outside PageTransition, so its scrollTop
	   survives a route change unless someone puts it back. */
	$effect(() => {
		if (page.url.pathname) scroller?.scrollTo({ top: 0 });
	});

	/*
	 * No Settings row here, by design — the user card at the bottom of the rail is the only
	 * way in, and the avatar on MobileNav is its phone counterpart. See CHARTE §14.
	 */
	const links = [
		{ href: '/sites', label: 'Sites', icon: icons.globe },
		{ href: '/team', label: 'Teams', icon: icons.usersGroup }
	];

	function isActive(href: string) {
		return page.url.pathname === href || page.url.pathname.startsWith(href + '/');
	}

	const navPages = $derived(links.map((l) => ({ ...l, active: isActive(l.href) })));
	const onSettings = $derived(isActive('/settings'));

	const user = $derived({
		name: userStore.value?.name?.trim() || userStore.value?.email || 'Account',
		avatar: userStore.value?.avatar_url || undefined
	});

	const spaces = $derived(workspaceStore.all.map((w) => ({ id: String(w.id), name: w.name })));
	const activeSpaceId = $derived(
		workspaceStore.current ? String(workspaceStore.current.id) : null
	);

	/* Vision has no personal scope: sites.workspace_id is NOT NULL and the migration mints a
	   real per-user workspace instead, so `null` is not a space this app can be in. Both
	   switchers are therefore passed a null label, which removes the row rather than leaving
	   one whose click this function silently drops. */
	function selectSpace(id: string | null) {
		const next = workspaceStore.all.find((w) => String(w.id) === id);
		if (next) workspaceStore.switchTo(next);
	}
</script>

<div class="flex h-dvh w-full overflow-hidden bg-fc-page">
	<div class="hidden h-full shrink-0 p-3 md:block">
		<SideBar
			icon="solar:panorama-bold-duotone"
			title="Vision"
			bind:collapsed
			pages={navPages}
			{spaces}
			{activeSpaceId}
			onSpaceSelect={selectSpace}
			personalSpaceLabel={null}
			manageSpacesHref="/team"
			{user}
			userHref="/settings"
			userActive={onSettings}
			class="h-full"
		/>
	</div>

	<!-- `overscroll-contain`: <main> is the only scroller, so a flick past either end has
	     nowhere useful to chain to and would otherwise rubber-band the whole shell.
	     `min-w-0` is what lets it shrink below its content's intrinsic width. -->
	<main
		bind:this={scroller}
		class="min-w-0 flex-1 overflow-auto overscroll-contain pb-28 md:pb-0"
	>
		<!-- Spaces live in the rail, and the rail is desktop-only — without this header there
		     is no way to switch space on a phone at all. -->
		<Topbar class="md:hidden">
			<span class="text-fc-md font-semibold text-fc-fg">Vision</span>
			<div class="min-w-0 max-w-56 flex-1">
				<SpaceSwitcher
					{spaces}
					activeId={activeSpaceId}
					onSelect={selectSpace}
					personalLabel={null}
					manageHref="/team"
				/>
			</div>
		</Topbar>

		<div class="mx-auto flex max-w-fc-xl flex-col gap-8 px-4 py-8 sm:px-6 md:px-10 md:py-10">
			<PageTransition key={page.url.pathname}>
				{@render children()}
			</PageTransition>
		</div>
	</main>

	<MobileNav items={navPages} {user} profileHref="/settings" profileActive={onSettings} />
</div>
