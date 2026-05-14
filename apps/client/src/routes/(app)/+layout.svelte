<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { isAuthenticated, clearToken } from '$lib';

	let { children } = $props();

	onMount(() => {
		if (!isAuthenticated()) {
			goto('/login');
		}
	});

	function logout() {
		clearToken();
		goto('/login');
	}
</script>

<div class="flex min-h-screen">
	<nav class="w-56 border-r bg-sidebar p-4 space-y-2">
		<h2 class="text-lg font-bold mb-6">Vision</h2>
		<a href="/dashboard" class="block rounded-md px-3 py-2 text-sm hover:bg-sidebar-accent">Dashboard</a>
		<a href="/sites" class="block rounded-md px-3 py-2 text-sm hover:bg-sidebar-accent">Sites</a>
		<a href="/settings" class="block rounded-md px-3 py-2 text-sm hover:bg-sidebar-accent">Settings</a>
		<button
			onclick={logout}
			class="block w-full text-left rounded-md px-3 py-2 text-sm text-muted-foreground hover:bg-sidebar-accent mt-auto"
		>
			Logout
		</button>
	</nav>
	<main class="flex-1 p-8">
		{@render children()}
	</main>
</div>
