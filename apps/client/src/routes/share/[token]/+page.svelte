<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { AnalyticsOverview } from '$lib';
	import { AreaChart, BarChart, PieChart } from 'layerchart';
	import { scaleBand } from 'd3-scale';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import * as ToggleGroup from '$lib/components/ui/toggle-group/index.js';
	import Icon from '@iconify/svelte';
	import WorldMap from '$lib/components/map/world-map.svelte';
	import StatCard from '$lib/components/stat-card.svelte';
	import BarListCard from '$lib/components/bar-list-card.svelte';
	import PieLegendCard from '$lib/components/pie-legend-card.svelte';
	import {
		type RangeKey,
		ranges as allRanges,
		rangeDates,
		formatDateRange,
		trendPercent,
		fmt,
		classifyReferrer,
		CHART_COLORS
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
	const ranges = allRanges.filter((r) => r.key !== 'custom');

	let dateRangeLabel = $derived(() => {
		const { from, to } = rangeDates(selectedRange);
		return formatDateRange(from, to);
	});

	let viewsPerVisitor = $derived(() => {
		if (!overview || overview.unique_visitors === 0) return 0;
		return overview.total_pageviews / overview.unique_visitors;
	});

	let prevViewsPerVisitor = $derived(() => {
		if (!overview || overview.prev_unique_visitors === 0) return 0;
		return overview.prev_total_pageviews / overview.prev_unique_visitors;
	});

	const chartConfig = {
		pageviews: { label: 'Pageviews', color: 'var(--chart-1)' },
		visitors: { label: 'Visitors', color: 'var(--chart-2)' }
	} satisfies Chart.ChartConfig;

	const trafficConfig = {
		direct: { label: 'Direct', color: 'var(--chart-1)' },
		search: { label: 'Search', color: 'var(--chart-2)' },
		social: { label: 'Social', color: 'var(--chart-3)' },
		other: { label: 'Other', color: 'var(--chart-4)' }
	} satisfies Chart.ChartConfig;

	const hourlyConfig = {
		count: { label: 'Pageviews', color: 'var(--chart-1)' }
	} satisfies Chart.ChartConfig;

	let trafficSourcesData = $derived(() => {
		if (!overview) return [];
		const refs = overview.top_referrers ?? [];
		let search = 0;
		let social = 0;
		let other = 0;
		let refTotal = 0;

		for (const r of refs) {
			refTotal += r.count;
			const cat = classifyReferrer(r.referrer);
			if (cat === 'search') search += r.count;
			else if (cat === 'social') social += r.count;
			else other += r.count;
		}

		const direct = Math.max(0, overview.total_pageviews - refTotal);

		return [
			{ key: 'direct', label: 'Direct', value: direct, color: 'var(--chart-1)' },
			{ key: 'search', label: 'Search', value: search, color: 'var(--chart-2)' },
			{ key: 'social', label: 'Social', value: social, color: 'var(--chart-3)' },
			{ key: 'other', label: 'Other', value: other, color: 'var(--chart-4)' }
		].filter((d) => d.value > 0);
	});

	let trafficSourcesTotal = $derived(() => {
		return trafficSourcesData().reduce((sum, d) => sum + d.value, 0);
	});

	async function refresh() {
		const token = (page.params as Record<string, string>).token;
		try {
			const { from, to } = rangeDates(selectedRange);
			const data = await api.share.overview(token, from, to);
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
		const token = (page.params as Record<string, string>).token;
		try {
			const res = await api.share.realtime(token);
			realtimeCount = res.visitors;
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

	let chartData = $derived(
		(overview?.pageviews_per_day ?? []).map((d) => ({
			date: d.date,
			pageviews: d.count,
			visitors: overview?.unique_visitors_per_day?.find((v) => v.date === d.date)?.count ?? 0
		}))
	);

	let hourlyData = $derived(overview?.hourly_distribution ?? []);

	let topPagesData = $derived(
		(overview?.top_pages ?? []).slice(0, 8).map((d) => ({ label: d.path, count: d.count }))
	);
	let topReferrersData = $derived(
		(overview?.top_referrers ?? []).slice(0, 8).map((d) => ({ label: d.referrer, count: d.count }))
	);
	let browsersData = $derived(
		(overview?.top_browsers ?? []).slice(0, 8).map((d) => ({ label: d.browser, count: d.count }))
	);
	let osData = $derived(
		(overview?.top_os ?? []).slice(0, 8).map((d) => ({ label: d.os, count: d.count }))
	);

	let devicesPieData = $derived(
		(overview?.top_devices ?? []).map((d, i) => ({
			key: d.device,
			label: d.device,
			value: d.count,
			color: CHART_COLORS[i % CHART_COLORS.length]
		}))
	);

	let screensPieData = $derived(
		(overview?.top_screens ?? []).map((d, i) => ({
			key: d.screen,
			label: d.screen,
			value: d.count,
			color: CHART_COLORS[i % CHART_COLORS.length]
		}))
	);

	let pageviewsTrend = $derived(
		overview
			? trendPercent(overview.total_pageviews, overview.prev_total_pageviews)
			: { text: '—', color: 'text-muted-foreground' }
	);
	let visitorsTrend = $derived(
		overview
			? trendPercent(overview.unique_visitors, overview.prev_unique_visitors)
			: { text: '—', color: 'text-muted-foreground' }
	);
	let vpvTrend = $derived(trendPercent(viewsPerVisitor(), prevViewsPerVisitor()));
	let bounceTrend = $derived(
		overview
			? trendPercent(overview.bounce_rate, overview.prev_bounce_rate)
			: { text: '—', color: 'text-muted-foreground' }
	);
</script>

<svelte:head><title>{siteName ? `${siteName} — Vision` : 'Vision'}</title></svelte:head>

{#if loading}
	<div class="flex min-h-[60dvh] items-center justify-center">
		<Icon icon="solar:spinner-linear" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else if notFound}
	<div class="flex min-h-[60dvh] flex-col items-center justify-center gap-4">
		<Icon icon="solar:eye-closed-linear" class="h-12 w-12 text-muted-foreground" />
		<h1 class="text-xl font-semibold">Dashboard not found</h1>
		<p class="text-muted-foreground">This share link may have been revoked or is invalid.</p>
	</div>
{:else if overview}
	<div class="mx-auto max-w-7xl px-4 py-8">
		<div class="mb-8 flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-bold">{siteName}</h1>
				<p class="text-muted-foreground">{siteDomain}</p>
			</div>
			<div class="flex items-center gap-3 text-sm">
				<span class="relative flex h-2.5 w-2.5">
					<span
						class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75"
					></span>
					<span class="relative inline-flex h-2.5 w-2.5 rounded-full bg-green-500"></span>
				</span>
				<span class="text-green-600">Live</span>
			</div>
		</div>

		<div class="mb-6 flex flex-wrap items-center gap-3">
			<ToggleGroup.Root
				type="single"
				variant="outline"
				size="sm"
				value={selectedRange}
				onValueChange={(v) => {
					if (v) selectedRange = v as RangeKey;
				}}
			>
				{#each ranges as r (r.key)}
					<ToggleGroup.Item value={r.key}>{r.label}</ToggleGroup.Item>
				{/each}
			</ToggleGroup.Root>

			<span class="ml-auto text-xs text-muted-foreground">{dateRangeLabel()}</span>
		</div>

		<div class="mb-8 grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-5">
			<StatCard label="Pageviews" value={fmt(overview.total_pageviews)} trend={pageviewsTrend} />
			<StatCard label="Visitors" value={fmt(overview.unique_visitors)} trend={visitorsTrend} />
			<StatCard label="Views / Visitor" value={viewsPerVisitor().toFixed(1)} trend={vpvTrend} />
			<StatCard
				label="Bounce Rate"
				value="{overview.bounce_rate.toFixed(1)}%"
				trend={bounceTrend}
			/>
			<StatCard label="Active Now" value={fmt(realtimeCount)} pulse />
		</div>

		{#if chartData.length > 0}
			<Card.Root class="mb-8">
				<Card.Header>
					<Card.Title>Traffic Over Time</Card.Title>
					<Card.Description>Pageviews and unique visitors</Card.Description>
				</Card.Header>
				<Card.Content class="px-0">
					<Chart.Container config={chartConfig} class="aspect-auto h-[300px] w-full">
						<AreaChart
							data={chartData}
							x="date"
							xScale={scaleBand().padding(0.25)}
							axis="x"
							series={[
								{ key: 'pageviews', label: 'Pageviews', color: 'var(--chart-1)' },
								{ key: 'visitors', label: 'Visitors', color: 'var(--chart-2)' }
							]}
							props={{
								xAxis: {
									format: (d: string) => {
										const date = new Date(d);
										return `${date.getMonth() + 1}/${date.getDate()}`;
									}
								}
							}}
						>
							{#snippet tooltip()}
								<Chart.Tooltip />
							{/snippet}
						</AreaChart>
					</Chart.Container>
				</Card.Content>
			</Card.Root>
		{/if}

		<div class="mb-8 grid gap-4 md:grid-cols-2">
			{#if trafficSourcesData().length > 0}
				<Card.Root>
					<Card.Header>
						<Card.Title>Traffic Sources</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="flex items-center justify-center gap-6">
							<div class="aspect-square h-[200px]">
								<Chart.Container config={trafficConfig} class="aspect-auto h-full w-full">
									<PieChart
										data={trafficSourcesData()}
										key="key"
										label="label"
										value="value"
										innerRadius={0.6}
									/>
								</Chart.Container>
							</div>
							<div class="shrink-0 space-y-2">
								{#each trafficSourcesData() as source}
									{@const pct =
										trafficSourcesTotal() > 0
											? ((source.value / trafficSourcesTotal()) * 100).toFixed(1)
											: '0'}
									<div class="flex items-center gap-2 text-sm">
										<span
											class="h-3 w-3 shrink-0 rounded-full"
											style="background: {source.color}"
										></span>
										<span>{source.label}</span>
										<span class="tabular-nums text-muted-foreground">{pct}%</span>
									</div>
								{/each}
							</div>
						</div>
					</Card.Content>
				</Card.Root>
			{/if}

			{#if hourlyData.length > 0}
				<Card.Root>
					<Card.Header>
						<Card.Title>Hourly Distribution</Card.Title>
					</Card.Header>
					<Card.Content>
						<Chart.Container config={hourlyConfig} class="min-h-[200px] w-full">
							<BarChart
								data={hourlyData}
								x="hour"
								xScale={scaleBand().padding(0.25)}
								axis="x"
								series={[{ key: 'count', label: 'Pageviews', color: 'var(--chart-1)' }]}
								props={{
									xAxis: {
										format: (d: number) => {
											if (d % 3 === 0) return `${d}h`;
											return '';
										}
									}
								}}
							>
								{#snippet tooltip()}
									<Chart.Tooltip />
								{/snippet}
							</BarChart>
						</Chart.Container>
					</Card.Content>
				</Card.Root>
			{/if}
		</div>

		{#if overview.top_countries?.length > 0}
			<Card.Root class="mb-8">
				<Card.Header>
					<Card.Title>Visitors</Card.Title>
				</Card.Header>
				<Card.Content>
					<WorldMap countries={overview.top_countries} />
				</Card.Content>
			</Card.Root>
		{/if}

		<div class="mb-8 grid grid-cols-1 gap-4 lg:grid-cols-3">
			{#if topPagesData.length > 0}
				<BarListCard title="Top Pages" items={topPagesData} class="lg:col-span-2" />
			{/if}

			{#if devicesPieData.length > 0}
				<PieLegendCard title="Devices" data={devicesPieData} configKey="devices" />
			{/if}

			{#if topReferrersData.length > 0}
				<BarListCard title="Top Referrers" items={topReferrersData} class="lg:col-span-2">
					{#snippet label(item)}
						<a
							href="{item.label.includes('://') ? '' : 'https://'}{item.label}"
							target="_blank"
							rel="noopener noreferrer"
							class="flex items-center gap-2 truncate mr-2 hover:underline"
						>
							<img
								src="https://www.google.com/s2/favicons?domain={item.label}&sz=16"
								alt=""
								class="h-4 w-4 shrink-0"
							/>
							{item.label}
						</a>
					{/snippet}
				</BarListCard>
			{/if}

			{#if screensPieData.length > 0}
				<PieLegendCard
					title="Screens"
					data={screensPieData}
					configKey="screens"
					configColor="var(--chart-2)"
				/>
			{/if}

			{#if browsersData.length > 0}
				<BarListCard title="Browsers" items={browsersData} />
			{/if}

			{#if osData.length > 0}
				<BarListCard title="Operating Systems" items={osData} class="lg:col-span-2" />
			{/if}
		</div>

		<div class="flex items-center justify-center py-6 text-xs text-muted-foreground">
			<Icon icon="solar:chart-square-linear" class="mr-1.5 h-3.5 w-3.5" />
			Powered by Vision
		</div>
	</div>
{/if}
