<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { Site, AnalyticsOverview } from '$lib';
	import { AreaChart, BarChart, PieChart } from 'layerchart';
	import { scaleBand } from 'd3-scale';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { Copy, Check, Sun, Moon } from '@lucide/svelte';
	import WorldMap from '$lib/components/map/world-map.svelte';

	let site = $state<Site | null>(null);
	let overview = $state<AnalyticsOverview | null>(null);
	let live = $state(false);
	let copied = $state(false);
	let realtimeCount = $state(0);
	let pollTimer: ReturnType<typeof setInterval> | null = null;
	let realtimeTimer: ReturnType<typeof setInterval> | null = null;

	type RangeKey = 'today' | 'this_week' | '7d' | 'this_month' | '30d' | '90d' | 'this_year' | 'custom';
	let selectedRange = $state<RangeKey>('30d');
	let customFrom = $state('');
	let customTo = $state('');

	let darkMode = $state(false);

	const ranges: { key: RangeKey; label: string }[] = [
		{ key: 'today', label: 'Today' },
		{ key: 'this_week', label: 'This Week' },
		{ key: '7d', label: 'Last 7 Days' },
		{ key: 'this_month', label: 'This Month' },
		{ key: '30d', label: 'Last 30 Days' },
		{ key: '90d', label: 'Last 90 Days' },
		{ key: 'this_year', label: 'This Year' },
		{ key: 'custom', label: 'Custom' }
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

		if (key === 'custom') {
			return { from: customFrom || toStr, to: customTo || toStr };
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

	function trackingSnippet(): string {
		return `<script defer src="${page.url.origin}/s.js?v=4"><\/script>`;
	}

	async function copySnippet() {
		await navigator.clipboard.writeText(trackingSnippet());
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}

	async function refresh(siteId: number) {
		try {
			const { from, to } = rangeDates(selectedRange);
			overview = await api.analytics.overview(siteId, from, to);
			live = true;
		} catch {
			live = false;
		}
	}

	async function fetchRealtime(siteId: number) {
		try {
			const res = await api.analytics.realtime.visitors(siteId);
			realtimeCount = res.visitors;
		} catch {}
	}

	function toggleDarkMode() {
		darkMode = !darkMode;
		document.documentElement.classList.toggle('dark', darkMode);
		localStorage.setItem('theme', darkMode ? 'dark' : 'light');
	}

	onMount(() => {
		const savedTheme = localStorage.getItem('theme');
		if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
			darkMode = true;
			document.documentElement.classList.add('dark');
		}

		const id = Number(page.params.id);

		(async () => {
			site = await api.sites.get(id);
			await Promise.all([refresh(id), fetchRealtime(id)]);
			pollTimer = setInterval(() => refresh(id), 5000);
			realtimeTimer = setInterval(() => fetchRealtime(id), 10000);
		})();

		return () => {
			if (pollTimer) clearInterval(pollTimer);
			if (realtimeTimer) clearInterval(realtimeTimer);
		};
	});

	$effect(() => {
		selectedRange;
		customFrom;
		customTo;
		const id = Number(page.params.id);
		if (id && site) refresh(id);
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
</script>

{#if site}
	<div class="mb-8 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold">{site.name}</h1>
			<p class="text-muted-foreground">{site.domain}</p>
		</div>
		<div class="flex items-center gap-3 text-sm">
			<button
				onclick={toggleDarkMode}
				class="rounded-md p-2 text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
				aria-label="Toggle dark mode"
			>
				{#if darkMode}
					<Sun class="h-4 w-4" />
				{:else}
					<Moon class="h-4 w-4" />
				{/if}
			</button>
			{#if live}
				<span class="relative flex h-2.5 w-2.5">
					<span
						class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75"
					></span>
					<span class="relative inline-flex h-2.5 w-2.5 rounded-full bg-green-500"></span>
				</span>
				<span class="text-green-600">Live</span>
			{:else}
				<span class="h-2.5 w-2.5 rounded-full bg-gray-300"></span>
				<span class="text-muted-foreground">Connecting…</span>
			{/if}
		</div>
	</div>

	<div class="mb-8 rounded-lg border p-4">
		<h2 class="font-semibold mb-2">Tracking Script</h2>
		<p class="text-sm text-muted-foreground mb-2">Add this to your website's &lt;head&gt;:</p>
		<div class="relative group">
			<pre
				class="rounded bg-muted p-3 pr-12 text-xs overflow-x-auto"
				>&lt;script defer src="{page.url.origin}/s.js?v=4"&gt;&lt;/script&gt;</pre
			>
			<button
				onclick={copySnippet}
				class="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1.5 text-muted-foreground hover:text-foreground hover:bg-background/80 transition-colors"
				aria-label="Copy tracking script"
			>
				{#if copied}
					<Check class="h-4 w-4 text-green-500" />
				{:else}
					<Copy class="h-4 w-4" />
				{/if}
			</button>
		</div>
	</div>

	{#if overview}
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

		{#if selectedRange === 'custom'}
			<div class="mb-4 flex items-center gap-2">
				<input
					type="date"
					bind:value={customFrom}
					class="rounded-md border bg-background px-3 py-1.5 text-sm"
				/>
				<span class="text-muted-foreground text-sm">to</span>
				<input
					type="date"
					bind:value={customTo}
					class="rounded-md border bg-background px-3 py-1.5 text-sm"
				/>
			</div>
		{/if}

		<p class="mb-6 text-sm text-muted-foreground">{dateRangeLabel()}</p>

		<div class="grid gap-4 grid-cols-1 md:grid-cols-3 lg:grid-cols-5 mb-8">
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Total Pageviews</p>
				<p class="text-3xl font-bold">{overview.total_pageviews.toLocaleString()}</p>
				<p class="text-xs {pageviewsTrend.color}">{pageviewsTrend.text}</p>
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Unique Visitors</p>
				<p class="text-3xl font-bold">{overview.unique_visitors.toLocaleString()}</p>
				<p class="text-xs {visitorsTrend.color}">{visitorsTrend.text}</p>
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Views / Visitor</p>
				<p class="text-3xl font-bold">{viewsPerVisitor().toFixed(1)}</p>
				<p class="text-xs {vpvTrend.color}">{vpvTrend.text}</p>
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Bounce Rate</p>
				<p class="text-3xl font-bold">—</p>
				<p class="text-xs text-muted-foreground">Requires session tracking</p>
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Active Now</p>
				<div class="flex items-center gap-2">
					<span class="relative flex h-2 w-2">
						<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75"></span>
						<span class="relative inline-flex h-2 w-2 rounded-full bg-green-500"></span>
					</span>
					<p class="text-3xl font-bold">{realtimeCount.toLocaleString()}</p>
				</div>
			</div>
		</div>

		{#if chartData.length > 1}
			<Card.Root class="mb-8">
				<Card.Header>
					<Card.Title>Traffic Over Time</Card.Title>
					<Card.Description>Pageviews and unique visitors</Card.Description>
				</Card.Header>
				<Card.Content>
					<Chart.Container config={chartConfig} class="min-h-[300px] w-full">
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
						<div class="flex items-center gap-6">
							<div class="min-h-[200px] w-[200px] shrink-0">
								<Chart.Container config={trafficConfig} class="h-[200px] w-[200px]">
									<PieChart
										data={trafficSourcesData()}
										key="key"
										label="label"
										value="value"
										innerRadius={0.6}
									/>
								</Chart.Container>
							</div>
							<div class="space-y-2">
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

		<div class="grid gap-4 md:grid-cols-2 mb-8">
			{#if topPagesData.length > 0}
				<Card.Root>
					<Card.Header>
						<Card.Title>Top Pages</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="space-y-1">
							{#each topPagesData as item}
								<div class="relative">
									<div
										class="absolute inset-y-0 left-0 rounded bg-muted"
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

			{#if topReferrersData.length > 0}
				<Card.Root>
					<Card.Header>
						<Card.Title>Top Referrers</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="space-y-1">
							{#each topReferrersData as item}
								<div class="relative">
									<div
										class="absolute inset-y-0 left-0 rounded bg-muted"
										style="width: {(item.count / maxReferrerCount) * 100}%"
									></div>
									<div class="relative flex items-center justify-between px-3 py-1.5 text-sm">
										<span class="flex items-center gap-2 truncate mr-2">
											<img src="https://www.google.com/s2/favicons?domain={item.referrer}&sz=16" alt="" class="h-4 w-4 shrink-0" />
											{item.referrer}
										</span>
										<span class="text-muted-foreground tabular-nums shrink-0">{item.count}</span>
									</div>
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
			{/if}
		</div>

		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
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
										class="absolute inset-y-0 left-0 rounded bg-muted"
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
				<Card.Root>
					<Card.Header>
						<Card.Title>Operating Systems</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="space-y-1">
							{#each osData as item}
								<div class="relative">
									<div
										class="absolute inset-y-0 left-0 rounded bg-muted"
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

	{/if}
{/if}
