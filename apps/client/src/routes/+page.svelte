<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { currentUser } from '$lib';
	import {
		Card,
		LandingFooter,
		LandingHeader,
		PricingPlans,
		SignInWithFacile,
		icons
	} from '@facile/muse';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import DashboardDemo from '$lib/components/landing/dashboard-demo.svelte';

	const logo = 'solar:panorama-bold-duotone';
	const signInHref = '/api/auth/oidc';

	let redirecting = $state(true);

	const heroLines = $derived(t('landing.heroTitle').split('|'));
	const features = [
		{ icon: icons.code, id: 'script' },
		{ icon: icons.dashboard, id: 'realtime' },
		{ icon: icons.bolt, id: 'goals' },
		{ icon: icons.shield, id: 'locked' }
	] as const;

	/**
	 * This page is where OIDC_SUCCESS_URL lands, so a signed-in visitor is sent on to /sites.
	 * porte issues a session cookie (no token in the fragment since v0.2.4), so asking the API
	 * is the only way to tell who is standing here.
	 */
	onMount(async () => {
		if (await currentUser().catch(() => null)) {
			goto('/sites');
			return;
		}
		redirecting = false;
	});
</script>

{#if !redirecting}
	<div class="min-h-dvh overflow-x-clip bg-fc-page text-fc-fg">
		<LandingHeader appName="Vision" icon={logo} skipHref="#main">
			{#snippet actions()}
				<SignInWithFacile href={signInHref} />
			{/snippet}
		</LandingHeader>

		<main id="main">
			<section class="mx-auto flex max-w-fc-xl flex-col gap-8 px-6 pt-12 pb-8 sm:pt-20">
				<div class="flex max-w-3xl flex-col gap-4">
					<h1 class="text-fc-2xl font-bold tracking-tight sm:text-fc-3xl">
						{heroLines[0]}<br />{heroLines[1]}
					</h1>
					<p class="text-fc-md text-fc-fg-muted sm:text-fc-lg">{t('landing.heroText')}</p>
				</div>
				<div class="flex flex-wrap items-center gap-3">
					<SignInWithFacile href={signInHref} size="lg" />
				</div>
				<DashboardDemo />
			</section>

			<section class="mx-auto max-w-fc-xl px-6 py-16">
				<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
					{#each features as feature (feature.id)}
						<Card class="flex flex-col gap-4">
							<span
								class="flex size-10 items-center justify-center rounded-fc-md bg-fc-surface text-fc-fg"
							>
								<iconify-icon icon={feature.icon} width="20" height="20" class="block"
								></iconify-icon>
							</span>
							<div class="flex flex-col gap-1">
								<h2 class="text-fc-lg font-bold">{t(`landing.feature.${feature.id}.title`)}</h2>
								<p class="text-fc-sm text-fc-fg-muted">{t(`landing.feature.${feature.id}.text`)}</p>
							</div>
						</Card>
					{/each}
				</div>
			</section>

			<PricingPlans
				highlight={[]}
				title={t('landing.pricing')}
				class="mx-auto max-w-fc-xl px-6 py-16"
			/>

			<section class="mx-auto max-w-fc-lg px-6 py-20 text-center">
				<h2 class="text-fc-2xl font-bold tracking-tight">{t('landing.ctaTitle')}</h2>
				<p class="mt-4 text-fc-sm text-fc-fg-muted">{t('landing.ctaText')}</p>
				<div class="mt-8 flex justify-center">
					<SignInWithFacile href={signInHref} size="lg" />
				</div>
			</section>
		</main>

		<LandingFooter
			appName="Vision"
			icon={logo}
			tagline={t('landing.tagline')}
			legal={{ onLocaleChange: rememberLocale }}
			year={new Date().getFullYear()}
		/>
	</div>
{/if}
