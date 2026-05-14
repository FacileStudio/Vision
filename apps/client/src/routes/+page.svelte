<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { isAuthenticated, setToken } from '$lib';
	import Icon from '@iconify/svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Separator } from '$lib/components/ui/separator';
	import BarChart2 from '@lucide/svelte/icons/bar-chart-2';
	import Globe from '@lucide/svelte/icons/globe';
	import ArrowRight from '@lucide/svelte/icons/arrow-right';

	let redirecting = $state(true);
	let ssoOnly = $state(false);

	onMount(async () => {
		const token = page.url.searchParams.get('token');
		if (token) {
			setToken(token);
			goto('/dashboard');
			return;
		}
		if (isAuthenticated()) {
			goto('/dashboard');
			return;
		}

		try {
			const cfg = await fetch('/api/auth/config').then((r) => r.json());
			ssoOnly = cfg.sso_only ?? false;
		} catch {}

		redirecting = false;
	});
</script>

<svelte:head>
	<title>Vision — Web Analytics</title>
	<meta name="description" content="Simple, privacy-friendly web analytics for your projects." />
</svelte:head>

{#if !redirecting}
<div class="min-h-screen bg-background text-foreground">
	<header class="border-b border-border">
		<div class="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
			<div class="flex h-14 items-center gap-3">
				<Icon icon="solar:panorama-bold-duotone" class="w-7 h-7" />
				<span class="text-2xl font-bold tracking-tight">Vision</span>
			</div>
			<div class="flex items-center gap-2">
				<Button variant="ghost" href="/login">Log in</Button>
				<Button href={ssoOnly ? '/login' : '/login?tab=register'}>
					{ssoOnly ? 'Continue with SSO' : 'Get started'}
				</Button>
			</div>
		</div>
	</header>

	<main>
		<section class="mx-auto max-w-5xl px-6 py-24 text-center">
			<h1 class="text-5xl font-bold tracking-tight">
				See your traffic.<br />Know your audience.
			</h1>
			<p class="mx-auto mt-6 max-w-xl text-lg text-muted-foreground">
				Vision is a lightweight web analytics tracker for your projects.
				One script tag, zero config, full visibility.
			</p>
			<div class="mt-10 flex justify-center gap-3">
				<Button size="lg" href={ssoOnly ? '/login' : '/login?tab=register'}>
					{ssoOnly ? 'Continue with SSO' : 'Start tracking'}
					<ArrowRight class="ml-2 size-4" />
				</Button>
				<Button size="lg" variant="outline" href="/login">Log in</Button>
			</div>
		</section>

		<Separator />

		<section class="mx-auto max-w-5xl px-6 py-20">
			<div class="grid gap-6 md:grid-cols-3">
				<Card.Root class="border border-border">
					<Card.Header>
						<div class="mb-2 flex size-10 items-center justify-center rounded-md border border-border">
							<Icon icon="solar:code-bold-duotone" class="size-5" />
						</div>
						<Card.Title>One script tag</Card.Title>
						<Card.Description>
							Drop a single &lt;script&gt; tag on your site. No API keys, no config. It just works.
						</Card.Description>
					</Card.Header>
				</Card.Root>

				<Card.Root class="border border-border">
					<Card.Header>
						<div class="mb-2 flex size-10 items-center justify-center rounded-md border border-border">
							<BarChart2 class="size-5" />
						</div>
						<Card.Title>Real-time analytics</Card.Title>
						<Card.Description>
							Pageviews, unique visitors, top pages, referrers, and countries — all in one dashboard.
						</Card.Description>
					</Card.Header>
				</Card.Root>

				<Card.Root class="border border-border">
					<Card.Header>
						<div class="mb-2 flex size-10 items-center justify-center rounded-md border border-border">
							<Globe class="size-5" />
						</div>
						<Card.Title>Domain-locked</Card.Title>
						<Card.Description>
							Events are verified against registered domains. No one can spoof your data.
						</Card.Description>
					</Card.Header>
				</Card.Root>
			</div>
		</section>

		<Separator />

		<section class="mx-auto max-w-5xl px-6 py-20 text-center">
			<h2 class="text-3xl font-bold tracking-tight">
				{ssoOnly ? 'Ready to sign in?' : 'Ready to start?'}
			</h2>
			<p class="mt-4 text-muted-foreground">
				{ssoOnly ? 'Use your organization SSO to access Vision.' : 'Free to use. Self-hosted. No credit card required.'}
			</p>
			<Button class="mt-8" size="lg" href={ssoOnly ? '/login' : '/login?tab=register'}>
				{ssoOnly ? 'Continue with SSO' : 'Create an account'}
			</Button>
		</section>
	</main>

	<footer class="border-t border-border text-center text-muted">
		<div class="mx-auto max-w-5xl px-6 py-6 text-sm text-muted-foreground">
			© {new Date().getFullYear()} Vision by <a href="https://facile.studio" class="underline hover:cursor-pointer font-semibold">Facile.</a>
		</div>
	</footer>
</div>
{/if}
