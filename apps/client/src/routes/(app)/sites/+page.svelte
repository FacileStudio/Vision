<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { SiteSummary } from '$lib';
	import { Button, ConfirmModal, EmptyState, Skeleton, StatCard, StatusDot, Tabs, icons, toast } from '@facile/muse';
	import AddSiteDrawer from '$lib/components/add-site-drawer.svelte';
	import SiteSummaryRow from '$lib/components/site-summary-row.svelte';
	import SpaceFilter from '$lib/components/space-filter.svelte';
	import { spaceFilter, workspaceStore } from '$lib/stores/workspace.svelte';
	import { periodFilter, periodWindow, periods, type Period } from '$lib/utils/period';
	import { trendPercent } from '$lib/utils/analytics';
	import { fmt } from '$lib/i18n/format';
	import { locale, t } from '$lib/i18n/index.svelte';

	let sites = $state<SiteSummary[]>([]);
	let loading = $state(true);
	let drawerOpen = $state(false);
	let pendingDelete = $state<SiteSummary | null>(null);
	let confirmOpen = $state(false);
	let request = 0;

	const space = $derived(spaceFilter(page.url));
	const period = $derived(periodFilter(page.url));
	const periodTabs = $derived(periods.map((id) => ({ id, label: t(`range.${id}`) })));
	const showSpace = $derived(space === 'all' && workspaceStore.all.length > 1);

	const totals = $derived(
		sites.reduce(
			(sum, s) => ({
				visitors: sum.visitors + s.unique_visitors,
				pageviews: sum.pageviews + s.pageviews,
				prevVisitors: sum.prevVisitors + s.prev_unique_visitors,
				prevPageviews: sum.prevPageviews + s.prev_pageviews,
				live: sum.live + s.live
			}),
			{ visitors: 0, pageviews: 0, prevVisitors: 0, prevPageviews: 0, live: 0 }
		)
	);
	const visitorsTrend = $derived(trendPercent(totals.visitors, totals.prevVisitors, locale()));
	const pageviewsTrend = $derived(trendPercent(totals.pageviews, totals.prevPageviews, locale()));

	/**
	 * Loads the summary for one space and period. A slower answer to an older filter never
	 * overwrites a newer one.
	 */
	async function load(scope: string, range: Period) {
		const mine = ++request;
		loading = true;
		const { from, to } = periodWindow(range);
		try {
			const answer = await api.analytics.summary(scope, from, to);
			if (mine === request) sites = answer.sites;
		} catch (e) {
			if (mine === request) toast.danger(e instanceof Error ? e.message : t('sites.loadFailed'));
		} finally {
			if (mine === request) loading = false;
		}
	}

	$effect(() => {
		load(space, period);
	});

	/** Writes the chosen period into the URL, dropping it when it is the default. */
	function choosePeriod(id: string) {
		const url = new URL(page.url);
		if (id === '7d') url.searchParams.delete('period');
		else url.searchParams.set('period', id);
		void goto(url, { replaceState: true, keepFocus: true, noScroll: true });
	}

	async function deleteSite() {
		const site = pendingDelete;
		if (!site) return;
		await api.sites.delete(site.site_id);
		sites = sites.filter((s) => s.site_id !== site.site_id);
		pendingDelete = null;
		toast.neutral(t('sites.deleted', { name: site.name }));
	}
</script>

<svelte:head><title>{t('sites.pageTitle')}</title></svelte:head>

<AddSiteDrawer
	bind:open={drawerOpen}
	space={space === 'all' ? '' : space}
	onCreated={(site) => {
		toast.success(t('sites.tracked', { name: site.name }));
		load(space, period);
	}}
/>

<div class="flex flex-col gap-8">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex min-w-0 flex-col gap-2">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('sites.heading')}</h1>
			<p class="text-fc-sm text-fc-fg-muted">{t('sites.subtitle')}</p>
		</div>
		<Button icon={icons.plus} onclick={() => (drawerOpen = true)}>{t('sites.add')}</Button>
	</div>

	<div class="flex flex-wrap items-center gap-3">
		{#if workspaceStore.all.length > 1}<SpaceFilter />{/if}
		<div class="min-w-0">
			<Tabs items={periodTabs} value={period} onChange={choosePeriod} label={t('range.label')} />
		</div>
	</div>

	{#if loading && sites.length === 0}
		<Skeleton class="h-24 w-full rounded-fc-md" />
		{#each [0, 1, 2] as row (row)}
			<Skeleton class="h-16 w-full rounded-fc-md" />
		{/each}
	{:else if sites.length === 0}
		<EmptyState icon={icons.globe} title={t('sites.empty.title')} description={t('sites.empty.text')}>
			<Button size="lg" icon={icons.plus} onclick={() => (drawerOpen = true)}>{t('sites.add')}</Button>
		</EmptyState>
	{:else}
		<div class="grid grid-cols-2 gap-4 md:grid-cols-3">
			<StatCard label={t('stat.visitors')} value={fmt(totals.visitors)}>
				<span class="text-fc-xs {visitorsTrend.color}" title={t('sites.vsPrevious')}>{visitorsTrend.text}</span>
			</StatCard>
			<StatCard label={t('stat.pageviews')} value={fmt(totals.pageviews)}>
				<span class="text-fc-xs {pageviewsTrend.color}" title={t('sites.vsPrevious')}>{pageviewsTrend.text}</span>
			</StatCard>
			<StatCard label={t('stat.activeNow')} value={fmt(totals.live)} class="col-span-2 md:col-span-1">
				<StatusDot tone="success" label={t('stat.rightNow')} pulse class="text-fc-xs" />
			</StatCard>
		</div>

		<section class="flex flex-col gap-3">
			{#each sites as site (site.site_id)}
				<SiteSummaryRow
					{site}
					{showSpace}
					onDelete={(s) => {
						pendingDelete = s;
						confirmOpen = true;
					}}
				/>
			{/each}
		</section>
	{/if}
</div>

<ConfirmModal
	bind:open={confirmOpen}
	tone="danger"
	title={t('sites.deleteTitle', { name: pendingDelete?.name ?? t('sites.thisSite') })}
	description={t('sites.deleteBody', { domain: pendingDelete?.domain ?? t('sites.thisDomain') })}
	confirmLabel={t('sites.deleteConfirm')}
	cancelLabel={t('common.keepIt')}
	onConfirm={deleteSite}
	onCancel={() => (pendingDelete = null)}
/>
