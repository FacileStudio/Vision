<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { BulleButton, BullePanel, Drawer, Skeleton, Alert } from '@facile/muse';
	import { bulleApi, upgradeLink, type Anomaly, type BulleProfile, type Explanation } from '$lib/bulle';
	import { bulleErrorKey, planRefusal } from '$lib/bulle-errors';
	import { AskRun } from '$lib/ask.svelte';
	import { locale, t } from '$lib/i18n/index.svelte';
	import type { AppKey } from '$lib/i18n/messages';
	import AnomalyCard from './anomaly-card.svelte';
	import AskData from './ask-data.svelte';

	let { site }: { site: { id: number; name: string; workspace_id: number } } = $props();

	const LANGUAGES = ['en', 'fr', 'es', 'de'];

	let profile = $state<BulleProfile | null>(null);
	let open = $state(false);
	let anomalies = $state<Anomaly[] | null>(null);
	let loadFailed = $state(false);
	let loadedFor = 0;
	let explanations = $state<Record<number, Explanation>>({});
	let explaining = $state<number[]>([]);
	let explainErrors = $state<Record<number, AppKey>>({});
	let upgradeUrl = $state<string | null>(null);
	const ask = new AskRun();

	const name = $derived(profile?.name?.trim() || 'Bulle');
	const gated = $derived(upgradeUrl !== null || ask.upgradeUrl !== null);
	const upgradeHref = $derived(gated ? upgradeLink(upgradeUrl || ask.upgradeUrl) : undefined);
	const busy = $derived(ask.busy || explaining.length > 0);
	const mood = $derived(
		busy ? 'thinking' : ask.error ? 'unavailable' : ask.run?.state === 'done' ? 'done' : 'idle'
	);

	onMount(async () => {
		try {
			profile = await bulleApi.profile();
		} catch {
			return;
		}
		try {
			const page = await bulleApi.ideas.list(site.workspace_id);
			if (!page.ai_access) upgradeUrl = page.upgrade_url ?? '';
		} catch {}
	});

	onDestroy(() => ask.stop());

	$effect(() => {
		if (open && loadedFor !== site.id) void loadAnomalies();
	});

	async function loadAnomalies() {
		loadedFor = site.id;
		loadFailed = false;
		try {
			anomalies = (await bulleApi.anomalies(site.id)).anomalies;
		} catch {
			loadFailed = true;
			loadedFor = 0;
		}
	}

	async function explain(anomaly: Anomaly) {
		explaining = [...explaining, anomaly.id];
		const { [anomaly.id]: _, ...rest } = explainErrors;
		explainErrors = rest;
		try {
			const lang = LANGUAGES.includes(locale()) ? locale() : 'en';
			explanations = { ...explanations, [anomaly.id]: await bulleApi.explain(anomaly.id, lang) };
		} catch (error) {
			const refusal = planRefusal(error);
			if (refusal) upgradeUrl = refusal.upgradeUrl;
			else explainErrors = { ...explainErrors, [anomaly.id]: bulleErrorKey(error) };
		} finally {
			explaining = explaining.filter((id) => id !== anomaly.id);
		}
	}

	function submit(question: string) {
		void ask.start({
			spaceId: site.workspace_id,
			siteId: site.id,
			title: `${site.name}`,
			question,
			locale: locale(),
			href: window.location.href
		});
	}
</script>

{#if profile}
	<BulleButton {name} {mood} label={t('bulle.label')} onclick={() => (open = true)} />

	<Drawer bind:open title={name} showClose>
		<div class="flex min-w-0 flex-col gap-8">
			<section class="flex min-w-0 flex-col gap-3" aria-labelledby="bulle-anomalies">
				<h3 id="bulle-anomalies" class="m-0 text-fc-lg font-medium text-fc-fg">{t('bulle.anomalies.heading')}</h3>
				<p class="m-0 text-fc-sm text-fc-fg-muted">{t('bulle.anomalies.intro')}</p>
				{#if loadFailed}
					<Alert tone="danger">{t('bulle.error.load')}</Alert>
				{:else if anomalies === null}
					<Skeleton class="h-24 w-full rounded-fc-md" />
				{:else if anomalies.length === 0}
					<p class="m-0 text-fc-sm text-fc-fg-muted">{t('bulle.anomalies.none')}</p>
				{:else}
					<ul class="m-0 flex list-none flex-col gap-3 p-0">
						{#each anomalies as anomaly (anomaly.id)}
							<AnomalyCard
								{anomaly}
								explanation={explanations[anomaly.id] ?? null}
								explaining={explaining.includes(anomaly.id)}
								error={explainErrors[anomaly.id] ?? null}
								canExplain={!gated}
								onExplain={() => void explain(anomaly)}
							/>
						{/each}
					</ul>
				{/if}
			</section>

			<BullePanel
				{name}
				title={t('bulle.ask.heading')}
				{busy}
				{mood}
				{upgradeHref}
				upsellText={t('bulle.upsell')}
				class="p-0 bg-transparent"
			>
				<AskData {ask} onAsk={submit} />
			</BullePanel>
		</div>
	</Drawer>
{/if}
