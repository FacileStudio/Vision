<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { currentUser } from '$lib';
	import {
		Card,
		Divider,
		LandingFooter,
		LandingHeader,
		PricingPlans,
		SignInWithFacile,
		icons
	} from '@facile/muse';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';

	const logo = 'solar:panorama-bold-duotone';
	const signInHref = '/api/auth/oidc';

	let redirecting = $state(true);

	const heroLines = $derived(t('landing.heroTitle').split('|'));
	const features = [
		{ icon: icons.code, id: 'script' },
		{ icon: icons.dashboard, id: 'realtime' },
		{ icon: icons.shield, id: 'locked' }
	] as const;

	/* This page is where OIDC_SUCCESS_URL lands, so it is the first thing an SSO
	   user sees after the identity provider sends them back. It used to read a
	   token out of the URL fragment; porte stopped putting one there in v0.2.4
	   and issues a session cookie instead, which is why asking the API is now
	   the only way to notice that the person standing here is signed in. */
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
	<div class="min-h-dvh bg-fc-page text-fc-fg">
		<LandingHeader appName="Vision" icon={logo} skipHref="#main">
			{#snippet actions()}
				<SignInWithFacile href={signInHref} />
			{/snippet}
		</LandingHeader>

		<main id="main">
			<section class="mx-auto max-w-fc-lg px-6 py-20 text-center sm:py-24">
				<h1 class="text-fc-3xl font-semibold tracking-tight">
					{heroLines[0]}<br />{heroLines[1]}
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-fc-md text-fc-fg-muted">{t('landing.heroText')}</p>
				<div class="mt-10 flex justify-center">
					<SignInWithFacile href={signInHref} size="lg" />
				</div>
			</section>

			<section class="mx-auto max-w-fc-xl px-6 py-20">
				<div class="grid gap-4 md:grid-cols-3">
					{#each features as feature (feature.id)}
						<Card class="flex flex-col gap-4">
							<span
								class="flex size-10 items-center justify-center rounded-fc-md bg-fc-surface text-fc-fg"
							>
								<iconify-icon icon={feature.icon} width="20" height="20" class="block"
								></iconify-icon>
							</span>
							<div class="flex flex-col gap-1">
								<h2 class="text-fc-lg font-semibold">{t(`landing.feature.${feature.id}.title`)}</h2>
								<p class="text-fc-sm text-fc-fg-muted">{t(`landing.feature.${feature.id}.text`)}</p>
							</div>
						</Card>
					{/each}
				</div>
			</section>

			<PricingPlans
				highlight={[]}
				title={t('landing.pricing')}
				class="mx-auto max-w-fc-xl px-6 py-20"
			/>

			<Divider class="my-0" />

			<section class="mx-auto max-w-fc-lg px-6 py-20 text-center">
				<h2 class="text-fc-2xl font-semibold tracking-tight">{t('landing.ctaTitle')}</h2>
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
