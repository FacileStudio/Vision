<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { currentUser } from '$lib';
	import {
		FeatureSplit,
		ForDevelopers,
		LANDING_HEADING,
		LandingCta,
		LandingFooter,
		LandingHeader,
		PricingPlans,
		SignInWithFacile,
		WorksWith,
		icons,
		type DeveloperRow,
		type FeatureItem,
		type WorksWithApp
	} from '@facile/muse';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import DashboardDemo from '$lib/components/landing/dashboard-demo.svelte';
	import LandingCards from '$lib/components/landing/landing-cards.svelte';

	const logo = 'solar:panorama-bold-duotone';
	const signInHref = '/api/auth/oidc';
	const repo = 'https://github.com/FacileStudio/Vision';
	const width = 'mx-auto w-full max-w-5xl px-4 sm:px-6';

	let redirecting = $state(true);

	const features: FeatureItem[] = $derived([
		{ icon: icons.eye, title: t('landing.feature.live.title'), text: t('landing.feature.live.text') },
		{ icon: icons.globe, title: t('landing.feature.places.title'), text: t('landing.feature.places.text') },
		{ icon: icons.external, title: t('landing.feature.share.title'), text: t('landing.feature.share.text') },
		{ icon: icons.usersGroup, title: t('landing.feature.team.title'), text: t('landing.feature.team.text') },
		{ icon: icons.shield, title: t('landing.feature.clean.title'), text: t('landing.feature.clean.text') },
		{ icon: icons.download, title: t('landing.feature.export.title'), text: t('landing.feature.export.text') }
	]);

	const apps: WorksWithApp[] = $derived([
		{ name: 'Bulle', icon: 'solar:smile-circle-bold-duotone', line: t('landing.with.bulle'), href: 'https://bulle.facile.studio' },
		{ name: 'Docs', icon: 'solar:documents-bold-duotone', line: t('landing.with.docs'), href: 'https://docs.facile.studio' }
	]);

	const devRows: DeveloperRow[] = $derived([
		{
			icon: icons.code,
			title: t('landing.dev.tag'),
			text: t('landing.dev.tagText'),
			command: `<script defer src="https://vision.facile.studio/s.js"></` + 'script>'
		},
		{ kind: 'api', text: t('landing.dev.api'), href: `${repo}/blob/main/docs/api.md` },
		{ kind: 'mcp', text: t('landing.dev.mcp') },
		{ kind: 'source', text: t('landing.dev.source'), href: repo }
	]);

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

<svelte:head>
	<title>{t('landing.title')}</title>
	<meta name="description" content={t('landing.description')} />
</svelte:head>

{#if !redirecting}
	<div class="min-h-dvh overflow-x-clip bg-fc-page text-fc-fg">
		<LandingHeader appName="Vision" icon={logo} skipHref="#main">
			{#snippet actions()}
				<SignInWithFacile href={signInHref} size="sm" />
			{/snippet}
		</LandingHeader>

		<main id="main">
			<section class="{width} pt-20 pb-10 text-center sm:pt-24">
				<h1 class="font-fc-title text-fc-3xl font-bold tracking-tight">{t('landing.heroTitle')}</h1>
				<p class="mx-auto mt-5 max-w-xl text-fc-md text-fc-fg-muted">{t('landing.heroText')}</p>
				<LandingCta {signInHref} align="center" class="mt-8" />
			</section>

			<section class="{width} pb-8">
				<DashboardDemo />
			</section>

			<LandingCards class="{width} py-16" />

			<FeatureSplit
				id="features"
				class="{width} scroll-mt-20 py-16"
				title={t('landing.features.title')}
				description={t('landing.features.line')}
				{features}
			/>

			<WorksWith class="{width} py-12" {apps} headingClass={LANDING_HEADING} />

			<PricingPlans
				id="pricing"
				class="{width} py-16"
				title={t('landing.pricing')}
				headingClass={LANDING_HEADING}
				highlight={['ai.access']}
				highlightOnly
				notes={{ '*': t('landing.plan.included') }}
				recommended="pro"
				suiteHref="https://suite.facile.studio/#tarifs"
				suiteNote={t('landing.plan.suite')}
			/>

			<section class="{width} py-20 text-center">
				<h2 class={LANDING_HEADING}>{t('landing.ctaTitle')}</h2>
				<p class="mt-3 text-fc-fg-muted">{t('landing.ctaText')}</p>
				<LandingCta {signInHref} align="center" class="mt-8" />
			</section>

			<div class="{width} pb-20">
				<ForDevelopers rows={devRows} description={t('landing.dev.line')} />
			</div>
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
