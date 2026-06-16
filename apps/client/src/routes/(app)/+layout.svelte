<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api, isAuthenticated, clearToken } from '$lib';
	import Icon from '@iconify/svelte';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import { userStore } from '$lib/stores/user.svelte';
	import MobileNav from '$lib/components/MobileNav.svelte';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import Globe from '@lucide/svelte/icons/globe';
	import Settings from '@lucide/svelte/icons/settings';
	import Users from '@lucide/svelte/icons/users';
	import LogOut from '@lucide/svelte/icons/log-out';

	let { children } = $props();

	function getInitials(name: string): string {
		return name
			.split(' ')
			.map((w) => w[0])
			.filter(Boolean)
			.slice(0, 2)
			.join('')
			.toUpperCase();
	}

	onMount(async () => {
		if (!isAuthenticated()) {
			goto('/login');
			return;
		}
		try {
			userStore.value = await api.auth.me();
		} catch {}
		try {
			const ws = await api.workspaces.list();
			workspaceStore.all = ws;
			if (ws.length > 0) workspaceStore.current = ws[0];
		} catch {}
		api.auth.syncProfile().then(async () => {
			try {
				userStore.value = await api.auth.me();
			} catch {}
		}).catch(() => {});
	});

	function logout() {
		clearToken();
		goto('/login');
	}

	const navLinks = [
		{ href: '/sites', label: 'Sites', icon: Globe },
		{ href: '/team', label: 'Team', icon: Users },
		{ href: '/settings', label: 'Settings', icon: Settings }
	];
</script>

<div class="flex h-screen w-full overflow-hidden">
	<aside class="sticky top-0 hidden h-screen w-60 flex-col border-r bg-background md:flex">
		<div class="flex items-center gap-3 px-5 pt-8 pb-6">
			<Icon icon="solar:panorama-bold-duotone" class="w-7 h-7" />
			<span class="text-2xl font-bold tracking-tight">Vision</span>
		</div>

		{#if workspaceStore.current}
			<div class="mx-3 mb-3 rounded-md border border-border/60 bg-muted/40 px-3 py-2">
				<p class="truncate text-xs font-medium text-muted-foreground">Workspace</p>
				<p class="truncate text-sm font-semibold">{workspaceStore.current.name}</p>
			</div>
		{/if}

		<nav class="flex flex-1 flex-col gap-1 px-3">
			{#each navLinks as link}
				{@const active = page.url.pathname === link.href || page.url.pathname.startsWith(link.href + '/')}
				<a
					href={link.href}
					class="flex items-center gap-3 rounded-md px-3 py-2.5 text-sm transition-colors {active
						? 'bg-foreground text-background font-medium'
						: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
				>
					<link.icon class="h-4 w-4 shrink-0" />
					{link.label}
				</a>
			{/each}
		</nav>

		<Separator />

		<div class="flex flex-col gap-2 p-4">
			<a
				href="/profile"
				class="flex items-center gap-3 rounded-lg border border-border/70 bg-muted/40 p-2.5 transition-colors hover:bg-muted"
			>
				<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-border bg-foreground text-xs font-semibold text-background overflow-hidden">
					{#if userStore.value?.avatar_url}
						<img
							src="/api{userStore.value.avatar_url}"
							alt={userStore.value.name || userStore.value.email}
							class="h-full w-full object-cover"
						/>
					{:else}
						{userStore.value ? getInitials(userStore.value.name || userStore.value.email) : '..'}
					{/if}
				</div>
				<div class="min-w-0 flex-1">
					<p class="truncate text-sm font-medium">{userStore.value?.name || 'Set your profile'}</p>
					<p class="truncate text-xs text-muted-foreground">{userStore.value?.email ?? ''}</p>
				</div>
			</a>
			<Button
				variant="ghost"
				size="sm"
				class="w-full justify-start gap-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
				onclick={logout}
			>
				<LogOut class="h-4 w-4" />
				Logout
			</Button>
		</div>
	</aside>

	<main class="flex-1 overflow-auto p-4 pb-24 sm:p-6 md:p-8 md:pb-8">
		{@render children()}
	</main>
</div>

<MobileNav user={userStore.value} />
