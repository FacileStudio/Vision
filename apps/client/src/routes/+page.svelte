<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { isAuthenticated, setToken } from '$lib';
	import { Card, Divider, icons } from '@facile/muse';

	let redirecting = $state(true);
	let ssoOnly = $state(false);

	/*
	 * muse's Button renders a <button>, and a landing page's calls to action have to be
	 * real anchors — crawlable, middle-clickable. These mirror Button's primary and outline
	 * variants on an <a>; they are the only place in the app that restates them.
	 */
	const primaryLink =
		'inline-flex h-11 shrink-0 items-center justify-center gap-2 whitespace-nowrap rounded-fc-pill bg-fc-accent px-6 text-fc-md font-medium text-fc-accent-fg transition-opacity hover:opacity-90 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring';
	const outlineLink =
		'inline-flex h-11 shrink-0 items-center justify-center gap-2 whitespace-nowrap rounded-fc-pill border border-fc-border px-6 text-fc-md font-medium text-fc-fg transition-colors hover:bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring';

	const features = [
		{
			icon: icons.code,
			title: 'One script tag',
			body: 'Drop a single <script> tag on your site. No API keys, no config, nothing to maintain.'
		},
		{
			icon: icons.dashboard,
			title: 'Real-time analytics',
			body: 'Pageviews, visitors, entry and exit pages, referrers, devices and countries — one dashboard.'
		},
		{
			icon: icons.shield,
			title: 'Domain-locked',
			body: 'Events are verified against the domains you registered, so nobody can inflate your numbers.'
		}
	];

	function readSsoToken(): string | null {
		const fragment = location.hash.startsWith('#') ? location.hash.slice(1) : location.hash;
		return new URLSearchParams(fragment).get('token');
	}

	onMount(async () => {
		const token = readSsoToken();
		if (token) {
			setToken(token);
			history.replaceState(null, '', page.url.pathname);
			goto('/sites');
			return;
		}
		if (isAuthenticated()) {
			goto('/sites');
			return;
		}

		try {
			const cfg = await fetch('/api/auth/config').then((r) => r.json());
			ssoOnly = cfg.sso_only ?? false;
		} catch {}

		redirecting = false;
	});

	const startHref = $derived(ssoOnly ? '/login' : '/login?tab=register');
</script>

<svelte:head>
	<title>Vision — Web Analytics</title>
	<meta name="description" content="Simple, privacy-friendly web analytics for your projects." />
</svelte:head>

{#if !redirecting}
	<div class="min-h-dvh bg-fc-page text-fc-fg">
		<header class="border-b border-fc-border">
			<div class="mx-auto flex max-w-fc-xl items-center justify-between gap-4 px-6 py-4">
				<span class="flex items-center gap-3">
					<iconify-icon icon="solar:panorama-bold-duotone" width="24" height="24" class="block"
					></iconify-icon>
					<span class="text-fc-xl font-semibold tracking-tight">Vision</span>
				</span>
				<div class="flex items-center gap-2">
					<a href="/login" class="{outlineLink} h-9 px-4 text-fc-sm">Log in</a>
					<a href={startHref} class="{primaryLink} h-9 px-4 text-fc-sm">
						{ssoOnly ? 'Continue with SSO' : 'Get started'}
					</a>
				</div>
			</div>
		</header>

		<main>
			<section class="mx-auto max-w-fc-lg px-6 py-20 text-center sm:py-24">
				<h1 class="text-fc-3xl font-semibold tracking-tight">
					See your traffic.<br />Know your audience.
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-fc-md text-fc-fg-muted">
					Vision is a lightweight, self-hosted analytics tracker for your projects. One script
					tag, zero config, full visibility.
				</p>
				<div class="mt-10 flex flex-col justify-center gap-3 sm:flex-row">
					<a href={startHref} class={primaryLink}>
						{ssoOnly ? 'Continue with SSO' : 'Start tracking'}
					</a>
					<a href="/login" class={outlineLink}>Log in</a>
				</div>
			</section>

			<Divider class="my-0" />

			<section class="mx-auto max-w-fc-xl px-6 py-20">
				<div class="grid gap-4 md:grid-cols-3">
					{#each features as feature (feature.title)}
						<Card class="flex flex-col gap-4">
							<span
								class="flex size-10 items-center justify-center rounded-fc-md bg-fc-surface text-fc-fg"
							>
								<iconify-icon icon={feature.icon} width="20" height="20" class="block"
								></iconify-icon>
							</span>
							<div class="flex flex-col gap-1">
								<h2 class="text-fc-lg font-semibold">{feature.title}</h2>
								<p class="text-fc-sm text-fc-fg-muted">{feature.body}</p>
							</div>
						</Card>
					{/each}
				</div>
			</section>

			<Divider class="my-0" />

			<section class="mx-auto max-w-fc-lg px-6 py-20 text-center">
				<h2 class="text-fc-2xl font-semibold tracking-tight">
					{ssoOnly ? 'Ready to sign in?' : 'Ready to start?'}
				</h2>
				<p class="mt-4 text-fc-sm text-fc-fg-muted">
					{ssoOnly
						? 'Use your organisation SSO to reach Vision.'
						: 'Free to use. Self-hosted. No credit card, no third party.'}
				</p>
				<a href={startHref} class="{primaryLink} mt-8">
					{ssoOnly ? 'Continue with SSO' : 'Create an account'}
				</a>
			</section>
		</main>

		<footer class="border-t border-fc-border">
			<div class="mx-auto max-w-fc-xl px-6 py-6 text-center text-fc-sm text-fc-fg-muted">
				© {new Date().getFullYear()} Vision by
				<a href="https://facile.studio" class="font-semibold text-fc-fg underline">Facile.</a>
			</div>
		</footer>
	</div>
{/if}
