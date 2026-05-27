<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { AnalyticsOverview } from '$lib';
	import { AreaChart, BarChart, PieChart } from 'layerchart';
	import { scaleBand } from 'd3-scale';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import Icon from '@iconify/svelte';
	import WorldMap from '$lib/components/map/world-map.svelte';

	function fmt(n: number): string {
		return n.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
	}

	let siteName = $state('');
	let siteDomain = $state('');
	let overview = $state<AnalyticsOverview | null>(null);
	let realtimeCount = $state(0);
	let loading = $state(true);
	let notFound = $state(false);
	let pollTimer: ReturnType<typeof setInterval> | null = null;
	let realtimeTimer: ReturnType<typeof setInterval> | null = null;

	type RangeKey = 'today' | 'this_week' | '7d' | 'this_month' | '30d' | '90d' | 'this_year';
	let selectedRange = $state<RangeKey>('30d');

	const ranges: { key: RangeKey; label: string }[] = [
		{ key: 'today', label: 'Today' },
		{ key: 'this_week', label: 'This Week' },
		{ key: '7d', label: 'Last 7 Days' },
		{ key: 'this_month', label: 'This Month' },
		{ key: '30d', label: 'Last 30 Days' },
		{ key: '90d', label: 'Last 90 Days' },
		{ key: 'this_year', label: 'This Year' }
	];

	function rangeDates(key: RangeKey): { from: string; to: string } {
		const now = new Date();
		const toStr = now.toISOString().slice(0, 10);

		if (key === 'today') return { from: toStr, to: toStr };

		if (key === 'this_week') {
			const day = now.getDay();
			const diff = day === 0 ? 6 : day - 1;
			const monday = new Date(now);
			monday.setDate(now.getDate() - diff);
			return { from: monday.toISOString().slice(0, 10), to: toStr };
		}

		if (key === 'this_month') {
			const first = new Date(now.getFullYear(), now.getMonth(), 1);
			return { from: first.toISOString().slice(0, 10), to: toStr };
		}

		if (key === 'this_year') {
			return { from: `${now.getFullYear()}-01-01`, to: toStr };
		}

		const daysMap: Record<string, number> = { '7d': 7, '30d': 30, '90d': 90 };
		const from = new Date(now);
		from.setDate(from.getDate() - (daysMap[key] ?? 30));
		return { from: from.toISOString().slice(0, 10), to: toStr };
	}

	function formatDateRange(from: string, to: string): string {
		const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric', year: 'numeric' };
		const f = new Date(from + 'T00:00:00');
		const t = new Date(to + 'T00:00:00');
		if (from === to) return f.toLocaleDateString('en-US', opts);
		return `${f.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })} – ${t.toLocaleDateString('en-US', opts)}`;
	}

	let dateRangeLabel = $derived(() => {
		const { from, to } = rangeDates(selectedRange);
		return formatDateRange(from, to);
	});

	function trendPercent(current: number, prev: number): { text: string; color: string } {
		if (prev === 0) return { text: '—', color: 'text-muted-foreground' };
		const pct = ((current - prev) / prev) * 100;
		if (pct > 0) return { text: `↑ ${pct.toFixed(1)}%`, color: 'text-green-500' };
		if (pct < 0) return { text: `↓ ${Math.abs(pct).toFixed(1)}%`, color: 'text-red-500' };
		return { text: '—', color: 'text-muted-foreground' };
	}

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

	const CHART_COLORS = ['var(--chart-1)', 'var(--chart-2)', 'var(--chart-3)', 'var(--chart-4)', 'var(--chart-5)'];

	const SEARCH_ENGINES = ['google', 'bing', 'duckduckgo', 'yahoo', 'baidu', 'yandex'];
	const SOCIAL_NETWORKS = ['twitter', 'x.com', 'facebook', 'instagram', 'linkedin', 'reddit', 'tiktok', 'youtube'];

	function classifyReferrer(ref: string): 'search' | 'social' | 'other' {
		const lower = ref.toLowerCase();
		if (SEARCH_ENGINES.some((s) => lower.includes(s))) return 'search';
		if (SOCIAL_NETWORKS.some((s) => lower.includes(s))) return 'social';
		return 'other';
	}

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

	const trafficConfig = {
		direct: { label: 'Direct', color: 'var(--chart-1)' },
		search: { label: 'Search', color: 'var(--chart-2)' },
		social: { label: 'Social', color: 'var(--chart-3)' },
		other: { label: 'Other', color: 'var(--chart-4)' }
	} satisfies Chart.ChartConfig;

	const hourlyConfig = {
		count: { label: 'Pageviews', color: 'var(--chart-1)' }
	} satisfies Chart.ChartConfig;

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

	let topPagesData = $derived((overview?.top_pages ?? []).slice(0, 8));
	let topReferrersData = $derived((overview?.top_referrers ?? []).slice(0, 8));

	let browsersData = $derived(overview?.top_browsers ?? []);
	let osData = $derived(overview?.top_os ?? []);
	let devicesData = $derived(overview?.top_devices ?? []);
	let screensData = $derived(overview?.top_screens ?? []);

	let devicesPieData = $derived(
		devicesData.map((d, i) => ({
			key: d.device,
			label: d.device,
			value: d.count,
			color: CHART_COLORS[i % CHART_COLORS.length]
		}))
	);

	let screensPieData = $derived(
		screensData.map((d, i) => ({
			key: d.screen,
			label: d.screen,
			value: d.count,
			color: CHART_COLORS[i % CHART_COLORS.length]
		}))
	);

	let maxBrowserCount = $derived(Math.max(...browsersData.map((d) => d.count), 1));
	let maxOsCount = $derived(Math.max(...osData.map((d) => d.count), 1));
	let maxPageCount = $derived(Math.max(...topPagesData.map((d) => d.count), 1));
	let maxReferrerCount = $derived(Math.max(...topReferrersData.map((d) => d.count), 1));

	let pageviewsTrend = $derived(overview ? trendPercent(overview.total_pageviews, overview.prev_total_pageviews) : { text: '—', color: 'text-muted-foreground' });
	let visitorsTrend = $derived(overview ? trendPercent(overview.unique_visitors, overview.prev_unique_visitors) : { text: '—', color: 'text-muted-foreground' });
	let vpvTrend = $derived(trendPercent(viewsPerVisitor(), prevViewsPerVisitor()));
	let bounceTrend = $derived(overview ? trendPercent(overview.bounce_rate, overview.prev_bounce_rate) : { text: '—', color: 'text-muted-foreground' });
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
					<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75"></span>
					<span class="relative inline-flex h-2.5 w-2.5 rounded-full bg-green-500"></span>
				</span>
				<span class="text-green-600">Live</span>
			</div>
		</div>

		<div class="mb-2 flex flex-wrap gap-2">
			{#each ranges as r}
				<button
					onclick={() => (selectedRange = r.key)}
					class="rounded-full px-3 py-1 text-sm font-medium transition-colors {selectedRange === r.key
						? 'bg-foreground text-background'
						: 'bg-muted text-muted-foreground hover:text-foreground'}"
				>
					{r.label}
				</button>
			{/each}
		</div>

		<p class="mb-6 text-sm text-muted-foreground">{dateRangeLabel()}</p>

		<div class="grid gap-4 grid-cols-1 md:grid-cols-3 lg:grid-cols-5 mb-8">
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Total Pageviews</p>
				<p class="text-3xl font-bold">{fmt(overview.total_pageviews)}</p>
				{#if pageviewsTrend.text !== '—'}<p class="text-xs {pageviewsTrend.color}">{pageviewsTrend.text}</p>{/if}
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Unique Visitors</p>
				<p class="text-3xl font-bold">{fmt(overview.unique_visitors)}</p>
				{#if visitorsTrend.text !== '—'}<p class="text-xs {visitorsTrend.color}">{visitorsTrend.text}</p>{/if}
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Views / Visitor</p>
				<p class="text-3xl font-bold">{viewsPerVisitor().toFixed(1)}</p>
				{#if vpvTrend.text !== '—'}<p class="text-xs {vpvTrend.color}">{vpvTrend.text}</p>{/if}
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Bounce Rate</p>
				<p class="text-3xl font-bold">{overview.bounce_rate.toFixed(1)}%</p>
				{#if bounceTrend.text !== '—'}<p class="text-xs {bounceTrend.color}">{bounceTrend.text}</p>{/if}
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Active Now</p>
				<div class="flex items-center gap-2">
					<span class="relative flex h-2 w-2">
						<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75"></span>
						<span class="relative inline-flex h-2 w-2 rounded-full bg-green-500"></span>
					</span>
					<p class="text-3xl font-bold">{fmt(realtimeCount)}</p>
				</div>
			</div>
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
								{
									key: 'pageviews',
									label: 'Pageviews',
									color: 'var(--chart-1)'
								},
								{
									key: 'visitors',
									label: 'Visitors',
									color: 'var(--chart-2)'
								}
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

		<div class="grid md:grid-cols-2 gap-4 mb-8">
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
							<div class="space-y-2 shrink-0">
								{#each trafficSourcesData() as source}
									{@const pct = trafficSourcesTotal() > 0 ? ((source.value / trafficSourcesTotal()) * 100).toFixed(1) : '0'}
									<div class="flex items-center gap-2 text-sm">
										<span class="h-3 w-3 rounded-full shrink-0" style="background: {source.color}"></span>
										<span>{source.label}</span>
										<span class="text-muted-foreground tabular-nums">{pct}%</span>
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
								series={[
									{
										key: 'count',
										label: 'Pageviews',
										color: 'var(--chart-1)'
									}
								]}
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

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-8">
			{#if topPagesData.length > 0}
				<Card.Root class="lg:col-span-2">
					<Card.Header>
						<Card.Title>Top Pages</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="space-y-1">
							{#each topPagesData as item}
								<div class="relative">
									<div
										class="absolute inset-y-0 left-0 rounded bg-muted/50"
										style="width: {(item.count / maxPageCount) * 100}%"
									></div>
									<div class="relative flex justify-between px-3 py-1.5 text-sm">
										<span class="truncate mr-2">{item.path}</span>
										<span class="text-muted-foreground tabular-nums shrink-0">{item.count}</span>
									</div>
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
			{/if}

			{#if devicesData.length > 0}
				<Card.Root>
					<Card.Header>
						<Card.Title>Devices</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="min-h-[180px]">
							<Chart.Container config={{ devices: { label: 'Devices', color: 'var(--chart-1)' } }} class="h-[180px] w-full">
								<PieChart
									data={devicesPieData}
									key="key"
									label="label"
									value="value"
									innerRadius={0.5}
								/>
							</Chart.Container>
						</div>
						<div class="mt-3 space-y-1">
							{#each devicesPieData as item}
								<div class="flex items-center gap-2 text-xs">
									<span class="h-2.5 w-2.5 rounded-full shrink-0" style="background: {item.color}"></span>
									<span class="truncate">{item.label}</span>
									<span class="text-muted-foreground tabular-nums ml-auto">{item.value}</span>
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
			{/if}

			{#if topReferrersData.length > 0}
				<Card.Root class="lg:col-span-2">
					<Card.Header>
						<Card.Title>Top Referrers</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="space-y-1">
							{#each topReferrersData as item}
								<div class="relative">
									<div
										class="absolute inset-y-0 left-0 rounded bg-muted/50"
										style="width: {(item.count / maxReferrerCount) * 100}%"
									></div>
									<div class="relative flex items-center justify-between px-3 py-1.5 text-sm">
										<a href="{item.referrer.includes('://') ? '' : 'https://'}{item.referrer}" target="_blank" rel="noopener noreferrer" class="flex items-center gap-2 truncate mr-2 hover:underline">
											<img src="https://www.google.com/s2/favicons?domain={item.referrer}&sz=16" alt="" class="h-4 w-4 shrink-0" />
											{item.referrer}
										</a>
										<span class="text-muted-foreground tabular-nums shrink-0">{item.count}</span>
									</div>
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
			{/if}

			{#if screensData.length > 0}
				<Card.Root>
					<Card.Header>
						<Card.Title>Screens</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="min-h-[180px]">
							<Chart.Container config={{ screens: { label: 'Screens', color: 'var(--chart-2)' } }} class="h-[180px] w-full">
								<PieChart
									data={screensPieData}
									key="key"
									label="label"
									value="value"
									innerRadius={0.5}
								/>
							</Chart.Container>
						</div>
						<div class="mt-3 space-y-1">
							{#each screensPieData as item}
								<div class="flex items-center gap-2 text-xs">
									<span class="h-2.5 w-2.5 rounded-full shrink-0" style="background: {item.color}"></span>
									<span class="truncate">{item.label}</span>
									<span class="text-muted-foreground tabular-nums ml-auto">{item.value}</span>
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
			{/if}

			{#if browsersData.length > 0}
				<Card.Root>
					<Card.Header>
						<Card.Title>Browsers</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="space-y-1">
							{#each browsersData as item}
								<div class="relative">
									<div
										class="absolute inset-y-0 left-0 rounded bg-muted/50"
										style="width: {(item.count / maxBrowserCount) * 100}%"
									></div>
									<div class="relative flex justify-between px-3 py-1.5 text-sm">
										<span>{item.browser}</span>
										<span class="text-muted-foreground tabular-nums">{item.count}</span>
									</div>
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
			{/if}

			{#if osData.length > 0}
				<Card.Root class="lg:col-span-2">
					<Card.Header>
						<Card.Title>Operating Systems</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="space-y-1">
							{#each osData as item}
								<div class="relative">
									<div
										class="absolute inset-y-0 left-0 rounded bg-muted/50"
										style="width: {(item.count / maxOsCount) * 100}%"
									></div>
									<div class="relative flex justify-between px-3 py-1.5 text-sm">
										<span>{item.os}</span>
										<span class="text-muted-foreground tabular-nums">{item.count}</span>
									</div>
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
			{/if}
		</div>

		<div class="flex items-center justify-center py-6 text-xs text-muted-foreground">
			<Icon icon="solar:chart-square-linear" class="mr-1.5 h-3.5 w-3.5" />
			Powered by Vision
		</div>
	</div>
{/if}
