<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { AnalyticsOverview } from '$lib';
	import {
		BarChart,
		Card,
		DonutChart,
		LineChart,
		Skeleton,
		StatCard,
		StatusDot,
		Tabs,
		icons
	} from '@facile/muse';
	import WorldMap from '$lib/components/map/world-map.svelte';
	import BarList from '$lib/components/bar-list.svelte';
	import {
		type RangeKey,
		ranges as allRanges,
		rangeDates,
		formatDateRange,
		trendPercent,
		fmt,
		classifyReferrer,
		defaultGranularity,
		formatChartDate,
		mergeByLabel
	} from '$lib/utils/analytics';

	let siteName = $state('');
	let siteDomain = $state('');
	let overview = $state<AnalyticsOverview | null>(null);
	let realtimeCount = $state(0);
	let loading = $state(true);
	let notFound = $state(false);
	let pollTimer: ReturnType<typeof setInterval> | null = null;
	let realtimeTimer: ReturnType<typeof setInterval> | null = null;

	let selectedRange = $state<RangeKey>('30d');
	/* No custom range on a public dashboard: it is a report, not a query tool. */
	const ranges = allRanges.filter((r) => r.key !== 'custom');

	const token = $derived((page.params as Record<string, string>).token);
	const granularity = $derived(defaultGranularity(selectedRange));
	const dateRangeLabel = $derived.by(() => {
		const { from, to } = rangeDates(selectedRange);
		return formatDateRange(from, to);
	});

	const viewsPerVisitor = $derived(
		overview && overview.unique_visitors > 0
			? overview.total_pageviews / overview.unique_visitors
			: 0
	);
	const prevViewsPerVisitor = $derived(
		overview && overview.prev_unique_visitors > 0
			? overview.prev_total_pageviews / overview.prev_unique_visitors
			: 0
	);

	const pageviewsTrend = $derived(
		trendPercent(overview?.total_pageviews ?? 0, overview?.prev_total_pageviews ?? 0)
	);
	const visitorsTrend = $derived(
		trendPercent(overview?.unique_visitors ?? 0, overview?.prev_unique_visitors ?? 0)
	);
	const vpvTrend = $derived(trendPercent(viewsPerVisitor, prevViewsPerVisitor));
	const bounceTrend = $derived(
		trendPercent(overview?.bounce_rate ?? 0, overview?.prev_bounce_rate ?? 0)
	);

	const chartLabels = $derived(
		(overview?.pageviews_per_day ?? []).map((d) => formatChartDate(d.date, granularity))
	);
	const trafficSeries = $derived([
		{ name: 'Pageviews', data: (overview?.pageviews_per_day ?? []).map((d) => d.count) },
		{
			name: 'Visitors',
			data: (overview?.pageviews_per_day ?? []).map(
				(d) => overview?.unique_visitors_per_day?.find((v) => v.date === d.date)?.count ?? 0
			)
		}
	]);

	const hourlySeries = $derived([
		{ name: 'Pageviews', data: (overview?.hourly_distribution ?? []).map((h) => h.count) }
	]);
	const hourlyLabels = $derived(
		(overview?.hourly_distribution ?? []).map((h) => (h.hour % 3 === 0 ? `${h.hour}h` : ''))
	);

	const trafficSources = $derived.by(() => {
		if (!overview) return [];
		let search = 0;
		let social = 0;
		let other = 0;
		let refTotal = 0;

		for (const r of overview.top_referrers ?? []) {
			refTotal += r.count;
			const cat = classifyReferrer(r.referrer);
			if (cat === 'search') search += r.count;
			else if (cat === 'social') social += r.count;
			else other += r.count;
		}

		const direct = Math.max(0, overview.total_pageviews - refTotal);
		return [
			{ label: 'Direct', value: direct },
			{ label: 'Search', value: search },
			{ label: 'Social', value: social },
			{ label: 'Other', value: other }
		].filter((d) => d.value > 0);
	});

	/* Merged before the slice, like the site dashboard: these lists key their `{#each}` on the
	   label, so a repeated one is a fatal each_key_duplicate on a page anyone can open. */
	function top<T extends { count: number }>(list: T[] | undefined, label: (d: T) => string) {
		return mergeByLabel((list ?? []).map((d) => ({ label: label(d), count: d.count }))).slice(0, 8);
	}

	const topPages = $derived(top(overview?.top_pages, (d) => d.path));
	const topReferrers = $derived(top(overview?.top_referrers, (d) => d.referrer));
	const browsers = $derived(top(overview?.top_browsers, (d) => d.browser));
	const operatingSystems = $derived(top(overview?.top_os, (d) => d.os));
	const devices = $derived(
		mergeByLabel((overview?.top_devices ?? []).map((d) => ({ label: d.device, count: d.count }))).map(
			(d) => ({ label: d.label, value: d.count })
		)
	);
	const screens = $derived(
		mergeByLabel((overview?.top_screens ?? []).map((d) => ({ label: d.screen, count: d.count }))).map(
			(d) => ({ label: d.label, value: d.count })
		)
	);

	async function refresh() {
		try {
			const { from, to } = rangeDates(selectedRange);
			const data = await api.share.overview(token, from, to, granularity);
			siteName = data.site.name;
			siteDomain = data.site.domain;
			overview = data.overview;
			loading = false;
		} catch {
			notFound = true;
			loading = false;
		}
	}

	async function fetchRealtime() {
		try {
			realtimeCount = (await api.share.realtime(token)).visitors;
		} catch {}
	}

	onMount(() => {
		(async () => {
			await Promise.all([refresh(), fetchRealtime()]);
			pollTimer = setInterval(refresh, 30000);
			realtimeTimer = setInterval(fetchRealtime, 10000);
		})();

		return () => {
			if (pollTimer) clearInterval(pollTimer);
			if (realtimeTimer) clearInterval(realtimeTimer);
		};
	});

	$effect(() => {
		selectedRange;
		if (!loading && !notFound) refresh();
	});
</script>

<svelte:head><title>{siteName ? `${siteName} — Vision` : 'Vision'}</title></svelte:head>

<div class="min-h-dvh bg-fc-page">
	<div class="mx-auto flex max-w-fc-xl flex-col gap-10 px-4 py-8 sm:px-6 md:px-10 md:py-10">
		{#if loading}
			<Skeleton class="h-16 w-full rounded-fc-md" />
			<Skeleton class="h-64 w-full rounded-fc-md" />
		{:else if notFound}
			<div class="flex min-h-[60dvh] flex-col items-center justify-center gap-3 text-center">
				<iconify-icon icon={icons.eyeClosed} width="40" height="40" class="block text-fc-fg-muted"
				></iconify-icon>
				<h1 class="text-fc-xl font-semibold text-fc-fg">Dashboard not found</h1>
				<p class="max-w-sm text-fc-sm text-fc-fg-muted">
					This link has been revoked, or it never existed. Ask whoever shared it for a new one.
				</p>
			</div>
		{:else if overview}
			<div class="flex flex-wrap items-start justify-between gap-4">
				<div class="flex min-w-0 flex-col gap-2">
					<h1 class="truncate text-fc-2xl font-semibold text-fc-fg">{siteName}</h1>
					<p class="truncate text-fc-sm text-fc-fg-muted">{siteDomain}</p>
				</div>
				<StatusDot tone="success" label="Live" pulse />
			</div>

			<section class="flex flex-col gap-4">
				<div class="flex flex-wrap items-center gap-3">
					<div class="min-w-0 flex-1">
						<Tabs
							items={ranges.map((r) => ({ id: r.key, label: r.label }))}
							value={selectedRange}
							onChange={(id) => (selectedRange = id as RangeKey)}
							label="Date range"
						/>
					</div>
					<span class="text-fc-xs text-fc-fg-muted">{dateRangeLabel}</span>
				</div>

				<div class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-5">
					<StatCard label="Pageviews" value={fmt(overview.total_pageviews)}>
						<span class="text-fc-xs {pageviewsTrend.color}">{pageviewsTrend.text}</span>
					</StatCard>
					<StatCard label="Visitors" value={fmt(overview.unique_visitors)}>
						<span class="text-fc-xs {visitorsTrend.color}">{visitorsTrend.text}</span>
					</StatCard>
					<StatCard label="Views / visitor" value={viewsPerVisitor.toFixed(1)}>
						<span class="text-fc-xs {vpvTrend.color}">{vpvTrend.text}</span>
					</StatCard>
					<StatCard label="Bounce rate" value="{overview.bounce_rate.toFixed(1)}%">
						<span class="text-fc-xs {bounceTrend.color}">{bounceTrend.text}</span>
					</StatCard>
					<StatCard label="Active now" value={fmt(realtimeCount)}>
						<StatusDot tone="success" label="right now" pulse class="text-fc-xs" />
					</StatCard>
				</div>
			</section>

			<section class="flex flex-col gap-4">
				<div class="flex flex-col gap-1">
					<h2 class="text-fc-lg font-semibold text-fc-fg">Traffic over time</h2>
					<p class="text-fc-sm text-fc-fg-muted">Pageviews and unique visitors.</p>
				</div>
				<Card class="flex flex-col gap-4">
					<LineChart series={trafficSeries} labels={chartLabels} area height={280} />
				</Card>

				<div class="grid gap-4 lg:grid-cols-2">
					{#if trafficSources.length > 0}
						<Card class="flex flex-col gap-4">
							<p class="text-fc-sm font-medium text-fc-fg">Traffic sources</p>
							<DonutChart
								data={trafficSources}
								centerLabel="pageviews"
								centerValue={fmt(overview.total_pageviews)}
								class="flex-1"
							/>
						</Card>
					{/if}
					{#if hourlySeries[0].data.length > 0}
						<Card class="flex flex-col gap-4">
							<p class="text-fc-sm font-medium text-fc-fg">By hour of day</p>
							<BarChart series={hourlySeries} labels={hourlyLabels} height={220} />
						</Card>
					{/if}
				</div>
			</section>

			{#if (overview.top_countries?.length ?? 0) > 0}
				<section class="flex flex-col gap-4">
					<h2 class="text-fc-lg font-semibold text-fc-fg">Where they are</h2>
					<Card>
						<WorldMap countries={overview.top_countries} />
					</Card>
				</section>
			{/if}

			<section class="flex flex-col gap-4">
				<h2 class="text-fc-lg font-semibold text-fc-fg">Content</h2>
				<div class="grid gap-4 lg:grid-cols-3">
					<BarList title="Top pages" items={topPages} class="lg:col-span-2" />
					{#if devices.length > 0}
						<Card class="flex flex-col gap-4">
							<p class="text-fc-sm font-medium text-fc-fg">Devices</p>
							<DonutChart data={devices} class="flex-1" />
						</Card>
					{/if}

					<BarList title="Top referrers" items={topReferrers} class="lg:col-span-2" />
					{#if screens.length > 0}
						<Card class="flex flex-col gap-4">
							<p class="text-fc-sm font-medium text-fc-fg">Screens</p>
							<DonutChart data={screens} class="flex-1" />
						</Card>
					{/if}

					<BarList title="Browsers" items={browsers} />
					<BarList title="Operating systems" items={operatingSystems} class="lg:col-span-2" />
				</div>
			</section>

			<p class="flex items-center justify-center gap-1.5 py-6 text-fc-xs text-fc-fg-muted">
				<iconify-icon icon={icons.dashboard} width="14" height="14" class="block"></iconify-icon>
				Powered by Vision
			</p>
		{/if}
	</div>
</div>
