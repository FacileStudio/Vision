<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api, getToken } from '$lib';
	import type { Site, AnalyticsOverview, GoalConversionsResponse, Goal } from '$lib';
	import { AreaChart, BarChart, PieChart } from 'layerchart';
	import { scaleBand } from 'd3-scale';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import * as ToggleGroup from '$lib/components/ui/toggle-group/index.js';
	import { Copy, Check } from '@lucide/svelte';
	import Icon from '@iconify/svelte';
	import WorldMap from '$lib/components/map/world-map.svelte';
	import StatsDrawer from '$lib/components/stats-drawer.svelte';
	import SiteFavicon from '$lib/components/site-favicon.svelte';
	import StatCard from '$lib/components/stat-card.svelte';
	import BarListCard from '$lib/components/bar-list-card.svelte';
	import PieLegendCard from '$lib/components/pie-legend-card.svelte';
	import PerformanceCard from '$lib/components/performance-card.svelte';
	import {
		type RangeKey,
		type Granularity,
		ranges,
		rangeDates,
		formatDateRange,
		trendPercent,
		fmt,
		classifyReferrer,
		CHART_COLORS,
		allowedGranularities,
		defaultGranularity,
		formatChartDate
	} from '$lib/utils/analytics';

	let site = $state<Site | null>(null);
	let overview = $state<AnalyticsOverview | null>(null);
	let goalConversions = $state<GoalConversionsResponse | null>(null);
	let siteGoals = $state<Goal[]>([]);
	let showGoalForm = $state(false);
	let goalName = $state('');
	let goalType = $state<'pageview' | 'event'>('pageview');
	let goalPagePath = $state('');
	let goalEventName = $state('');
	let goalMatchType = $state('exact');
	let goalSaving = $state(false);
	let live = $state(false);
	let copied = $state(false);
	let realtimeCount = $state(0);
	let pollTimer: ReturnType<typeof setInterval> | null = null;
	let realtimeTimer: ReturnType<typeof setInterval> | null = null;
	let shareCopied = $state(false);
	let shareUrl = $derived(
		site?.share_token ? `${page.url.origin}/share/${site.share_token}` : ''
	);

	let selectedRange = $state<RangeKey>('30d');
	let customFrom = $state('');
	let customTo = $state('');
	let selectedGranularity = $state<Granularity>('day');

	let activeFilters = $state<Record<string, string>>({});
	let showFilters = $state(false);
	let activeDrawer = $state<string | null>(null);

	const filterFields = [
		{ key: 'country', label: 'Country', placeholder: 'e.g. US' },
		{ key: 'browser', label: 'Browser', placeholder: 'e.g. Chrome' },
		{ key: 'os', label: 'OS', placeholder: 'e.g. macOS' },
		{ key: 'device', label: 'Device', placeholder: 'e.g. Desktop' },
		{ key: 'path', label: 'Path', placeholder: 'e.g. /blog' },
		{ key: 'referrer', label: 'Referrer', placeholder: 'e.g. google.com' }
	];

	let activeFilterCount = $derived(Object.values(activeFilters).filter(Boolean).length);

	function clearFilters() {
		activeFilters = {};
	}

	let visibleGranularities = $derived(allowedGranularities(selectedRange));

	function selectRange(v: string) {
		const range = v as RangeKey;
		selectedRange = range;
		selectedGranularity = defaultGranularity(range);
	}

	function setFilter(key: string) {
		return (label: string) => {
			activeFilters = { ...activeFilters, [key]: label };
			showFilters = true;
		};
	}

	let dateRangeLabel = $derived(() => {
		const { from, to } = rangeDates(selectedRange, customFrom, customTo);
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
		visitors: { label: 'Visitors', color: 'var(--chart-2)' },
		prev_pageviews: { label: 'Prev Pageviews', color: 'var(--chart-1)' },
		prev_visitors: { label: 'Prev Visitors', color: 'var(--chart-2)' }
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

	function trackingSnippet(): string {
		return `<script defer src="${page.url.origin}/s.js"><\/script>`;
	}

	async function generateShare() {
		const id = Number(page.params.id);
		site = await api.sites.share(id);
	}

	async function revokeShare() {
		const id = Number(page.params.id);
		await api.sites.revokeShare(id);
		site = await api.sites.get(id);
	}

	async function copyShareUrl() {
		await navigator.clipboard.writeText(shareUrl);
		shareCopied = true;
		setTimeout(() => (shareCopied = false), 2000);
	}

	async function copySnippet() {
		await navigator.clipboard.writeText(trackingSnippet());
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}

	async function loadGoals(siteId: number) {
		try {
			siteGoals = await api.goals.list(siteId);
		} catch {}
	}

	async function refreshGoalConversions(siteId: number) {
		try {
			const { from, to } = rangeDates(selectedRange, customFrom, customTo);
			goalConversions = await api.goals.conversions(siteId, from, to);
		} catch {}
	}

	function resetGoalForm() {
		goalName = '';
		goalType = 'pageview';
		goalPagePath = '';
		goalEventName = '';
		goalMatchType = 'exact';
		showGoalForm = false;
	}

	async function saveGoal() {
		goalSaving = true;
		try {
			const id = Number(page.params.id);
			await api.goals.create({
				site_id: id,
				name: goalName,
				goal_type: goalType,
				event_name: goalType === 'event' ? goalEventName : undefined,
				page_path: goalType === 'pageview' ? goalPagePath : undefined,
				match_type: goalType === 'pageview' ? goalMatchType : undefined
			});
			resetGoalForm();
			await Promise.all([loadGoals(id), refreshGoalConversions(id)]);
		} catch {}
		goalSaving = false;
	}

	async function deleteGoal(goalId: number) {
		try {
			const id = Number(page.params.id);
			await api.goals.delete(goalId);
			await Promise.all([loadGoals(id), refreshGoalConversions(id)]);
		} catch {}
	}

	async function refresh(siteId: number) {
		try {
			const { from, to } = rangeDates(selectedRange, customFrom, customTo);
			overview = await api.analytics.overview(siteId, from, to, selectedGranularity, activeFilters);
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

	async function exportCSV() {
		const { from, to } = rangeDates(selectedRange, customFrom, customTo);
		const token = getToken();
		const res = await fetch(
			`/api/analytics/${page.params.id}/export?from=${from}&to=${to}&format=csv`,
			{ headers: { Authorization: `Bearer ${token}` } }
		);
		if (!res.ok) return;
		const blob = await res.blob();
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = 'vision-export.csv';
		a.click();
		URL.revokeObjectURL(url);
	}

	onMount(() => {
		const id = Number(page.params.id);

		(async () => {
			site = await api.sites.get(id);
			await Promise.all([refresh(id), fetchRealtime(id), loadGoals(id), refreshGoalConversions(id)]);
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
		selectedGranularity;
		activeFilters;
		const id = Number(page.params.id);
		if (id && site) {
			refresh(id);
			refreshGoalConversions(id);
		}
	});

	let chartData = $derived(
		(overview?.pageviews_per_day ?? []).map((d, i) => ({
			date: d.date,
			pageviews: d.count,
			visitors: overview?.unique_visitors_per_day?.find((v) => v.date === d.date)?.count ?? 0,
			prev_pageviews: overview?.prev_pageviews_per_day?.[i]?.count ?? 0,
			prev_visitors: overview?.prev_unique_visitors_per_day?.[i]?.count ?? 0
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
	let entryPagesData = $derived(
		(overview?.top_entry_pages ?? []).slice(0, 8).map((d) => ({ label: d.path, count: d.count }))
	);
	let exitPagesData = $derived(
		(overview?.top_exit_pages ?? []).slice(0, 8).map((d) => ({ label: d.path, count: d.count }))
	);
	let utmSourcesData = $derived(
		(overview?.top_utm_sources ?? []).slice(0, 8).map((d) => ({ label: d.value, count: d.count }))
	);
	let utmMediumsData = $derived(
		(overview?.top_utm_mediums ?? []).slice(0, 8).map((d) => ({ label: d.value, count: d.count }))
	);
	let utmCampaignsData = $derived(
		(overview?.top_utm_campaigns ?? []).slice(0, 8).map((d) => ({ label: d.value, count: d.count }))
	);
	let eventsData = $derived(
		(overview?.top_events ?? []).slice(0, 8).map((d) => ({ label: d.name, count: d.count }))
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

	let drawerConfig = $derived(() => {
		if (!overview || !activeDrawer)
			return { title: '', items: [] as { label: string; count: number }[], filterKey: '' };
		const configs: Record<
			string,
			{ title: string; items: { label: string; count: number }[]; filterKey: string }
		> = {
			pages: {
				title: 'Top Pages',
				items: (overview.top_pages ?? []).map((d) => ({ label: d.path, count: d.count })),
				filterKey: 'path'
			},
			referrers: {
				title: 'Top Referrers',
				items: (overview.top_referrers ?? []).map((d) => ({ label: d.referrer, count: d.count })),
				filterKey: 'referrer'
			},
			browsers: {
				title: 'Browsers',
				items: (overview.top_browsers ?? []).map((d) => ({ label: d.browser, count: d.count })),
				filterKey: 'browser'
			},
			os: {
				title: 'Operating Systems',
				items: (overview.top_os ?? []).map((d) => ({ label: d.os, count: d.count })),
				filterKey: 'os'
			},
			entry: {
				title: 'Entry Pages',
				items: (overview.top_entry_pages ?? []).map((d) => ({ label: d.path, count: d.count })),
				filterKey: 'path'
			},
			exit: {
				title: 'Exit Pages',
				items: (overview.top_exit_pages ?? []).map((d) => ({ label: d.path, count: d.count })),
				filterKey: 'path'
			},
			countries: {
				title: 'Countries',
				items: (overview.top_countries ?? []).map((d) => ({ label: d.country, count: d.count })),
				filterKey: 'country'
			},
			utm_sources: {
				title: 'UTM Sources',
				items: (overview.top_utm_sources ?? []).map((d) => ({ label: d.value, count: d.count })),
				filterKey: ''
			},
			utm_mediums: {
				title: 'UTM Mediums',
				items: (overview.top_utm_mediums ?? []).map((d) => ({ label: d.value, count: d.count })),
				filterKey: ''
			},
			utm_campaigns: {
				title: 'UTM Campaigns',
				items: (overview.top_utm_campaigns ?? []).map((d) => ({ label: d.value, count: d.count })),
				filterKey: ''
			},
			events: {
				title: 'Custom Events',
				items: (overview.top_events ?? []).map((d) => ({ label: d.name, count: d.count })),
				filterKey: ''
			}
		};
		return configs[activeDrawer] ?? { title: '', items: [], filterKey: '' };
	});

	function applyDrawerFilter(value: string) {
		const key = drawerConfig().filterKey;
		if (key) {
			activeFilters = { ...activeFilters, [key]: value };
			showFilters = true;
		}
		activeDrawer = null;
	}
</script>

<svelte:head><title>{site ? `${site.name} — Vision` : 'Vision'}</title></svelte:head>

{#if site}
	<div class="mb-8 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
		<div class="min-w-0">
			<div class="flex items-center gap-3">
				<SiteFavicon domain={site.domain} name={site.name} class="h-6 w-6" />
				<h1 class="truncate text-2xl font-bold">{site.name}</h1>
			</div>
			<p class="truncate text-muted-foreground">{site.domain}</p>
			<div class="mt-2 flex flex-wrap items-center gap-2">
				{#if site.share_token}
					<div class="flex min-w-0 max-w-full items-center gap-2 rounded-lg bg-muted/50 px-3 py-1.5">
						<Icon icon="solar:link-linear" class="h-4 w-4 shrink-0 text-muted-foreground" />
						<code class="truncate text-xs text-muted-foreground">{shareUrl}</code>
						<button
							onclick={copyShareUrl}
							class="shrink-0 rounded p-1 text-muted-foreground transition-colors hover:text-foreground"
							aria-label="Copy share link"
						>
							{#if shareCopied}
								<Icon icon="solar:check-circle-linear" class="h-4 w-4 text-green-500" />
							{:else}
								<Icon icon="solar:copy-linear" class="h-4 w-4" />
							{/if}
						</button>
						<button
							onclick={revokeShare}
							class="shrink-0 rounded p-1 text-red-500 transition-colors hover:bg-red-500/10"
							aria-label="Revoke share link"
						>
							<Icon icon="solar:trash-bin-trash-linear" class="h-4 w-4" />
						</button>
					</div>
				{:else}
					<button
						onclick={generateShare}
						class="flex items-center gap-1.5 rounded-full bg-muted px-3 py-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
					>
						<Icon icon="solar:share-linear" class="h-3.5 w-3.5" />
						Create public link
					</button>
				{/if}
			</div>
		</div>
		<div class="flex items-center gap-3 text-sm">
			<button
				onclick={exportCSV}
				class="flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
				aria-label="Export CSV"
			>
				<Icon icon="solar:download-linear" class="h-4 w-4" />
				Export
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

	<Card.Root class="mb-8">
		<Card.Header>
			<Card.Title>Tracking Script</Card.Title>
			<Card.Description>Add this to your website's &lt;head&gt;</Card.Description>
		</Card.Header>
		<Card.Content>
			<div class="group relative">
				<pre
					class="overflow-x-auto rounded bg-muted p-3 pr-12 text-xs"
				>&lt;script defer src="{page.url.origin}/s.js"&gt;&lt;/script&gt;</pre>
				<button
					onclick={copySnippet}
					class="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1.5 text-muted-foreground transition-colors hover:bg-background/80 hover:text-foreground"
					aria-label="Copy tracking script"
				>
					{#if copied}
						<Check class="h-4 w-4 text-green-500" />
					{:else}
						<Copy class="h-4 w-4" />
					{/if}
				</button>
			</div>
		</Card.Content>
	</Card.Root>

	{#if overview}
		<div class="mb-6 flex flex-wrap items-center gap-3">
			<ToggleGroup.Root
				type="single"
				variant="outline"
				size="sm"
				value={selectedRange}
				onValueChange={(v) => {
					if (v) selectRange(v);
				}}
			>
				{#each ranges as r (r.key)}
					<ToggleGroup.Item value={r.key}>{r.label}</ToggleGroup.Item>
				{/each}
			</ToggleGroup.Root>

			<button
				onclick={() => (showFilters = !showFilters)}
				class="flex items-center gap-1.5 rounded-md border px-2.5 py-1.5 text-xs font-medium transition-colors {activeFilterCount >
				0
					? 'border-foreground bg-foreground text-background'
					: 'text-muted-foreground hover:text-foreground'}"
			>
				<Icon icon="solar:filter-linear" class="h-3.5 w-3.5" />
				Filters
				{#if activeFilterCount > 0}
					<span class="rounded-full bg-background px-1.5 text-[10px] leading-4 text-foreground"
						>{activeFilterCount}</span
					>
				{/if}
			</button>
			{#if activeFilterCount > 0}
				<button
					onclick={clearFilters}
					class="text-xs text-muted-foreground transition-colors hover:text-foreground"
				>
					Clear
				</button>
			{/if}

			<span class="ml-auto text-xs text-muted-foreground">{dateRangeLabel()}</span>
		</div>

		{#if selectedRange === 'custom'}
			<div class="mb-4 flex items-center gap-2">
				<input
					type="date"
					bind:value={customFrom}
					class="rounded-md border bg-background px-3 py-1.5 text-sm"
				/>
				<span class="text-sm text-muted-foreground">to</span>
				<input
					type="date"
					bind:value={customTo}
					class="rounded-md border bg-background px-3 py-1.5 text-sm"
				/>
			</div>
		{/if}

		{#if showFilters}
			<div class="mb-6 grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-6">
				{#each filterFields as f}
					<div>
						<label for="filter-{f.key}" class="mb-1 block text-xs text-muted-foreground"
							>{f.label}</label
						>
						<input
							id="filter-{f.key}"
							type="text"
							value={activeFilters[f.key] ?? ''}
							oninput={(e) => {
								activeFilters = { ...activeFilters, [f.key]: e.currentTarget.value };
							}}
							placeholder={f.placeholder}
							class="w-full rounded-md border bg-background px-2.5 py-1.5 text-sm placeholder:text-muted-foreground"
						/>
					</div>
				{/each}
			</div>
		{/if}

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

		<Card.Root class="mb-8">
			<Card.Header>
				<div class="flex items-center justify-between">
					<Card.Title>Goals</Card.Title>
					{#if !showGoalForm}
						<button
							onclick={() => (showGoalForm = true)}
							class="flex items-center gap-1.5 rounded-full bg-foreground px-3 py-1 text-sm font-medium text-background transition-colors hover:bg-foreground/90"
						>
							<Icon icon="mdi:plus" class="h-4 w-4" />
							Add Goal
						</button>
					{/if}
				</div>
			</Card.Header>
			<Card.Content>
				{#if showGoalForm}
					<div class="mb-4 space-y-3 rounded-lg border p-4">
						<div>
							<label for="goal-name" class="mb-1 block text-sm font-medium">Name</label>
							<input
								id="goal-name"
								type="text"
								bind:value={goalName}
								placeholder="e.g. Signup, Pricing page visit"
								class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
							/>
						</div>
						<div>
							<span class="mb-1.5 block text-sm font-medium">Type</span>
							<div class="flex gap-2">
								<button
									onclick={() => (goalType = 'pageview')}
									class="rounded-full px-3 py-1 text-sm font-medium transition-colors {goalType === 'pageview' ? 'bg-foreground text-background' : 'bg-muted text-muted-foreground hover:text-foreground'}"
								>
									Pageview
								</button>
								<button
									onclick={() => (goalType = 'event')}
									class="rounded-full px-3 py-1 text-sm font-medium transition-colors {goalType === 'event' ? 'bg-foreground text-background' : 'bg-muted text-muted-foreground hover:text-foreground'}"
								>
									Custom Event
								</button>
							</div>
						</div>
						{#if goalType === 'pageview'}
							<div class="flex gap-2">
								<div class="flex-1">
									<label for="goal-path" class="mb-1 block text-sm font-medium">Path</label>
									<input
										id="goal-path"
										type="text"
										bind:value={goalPagePath}
										placeholder="/pricing"
										class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
									/>
								</div>
								<div class="w-36">
									<label for="goal-match" class="mb-1 block text-sm font-medium">Match</label>
									<select
										id="goal-match"
										bind:value={goalMatchType}
										class="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
									>
										<option value="exact">Exact</option>
										<option value="starts_with">Starts with</option>
										<option value="contains">Contains</option>
									</select>
								</div>
							</div>
						{:else}
							<div>
								<label for="goal-event" class="mb-1 block text-sm font-medium">Event Name</label>
								<input
									id="goal-event"
									type="text"
									bind:value={goalEventName}
									placeholder="signup"
									class="w-full rounded-md border bg-background px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
								/>
							</div>
						{/if}
						<div class="flex gap-2 pt-1">
							<button
								onclick={saveGoal}
								disabled={goalSaving || !goalName || (goalType === 'pageview' ? !goalPagePath : !goalEventName)}
								class="flex items-center gap-1.5 rounded-full bg-foreground px-4 py-1.5 text-sm font-medium text-background transition-colors hover:bg-foreground/90 disabled:cursor-not-allowed disabled:opacity-50"
							>
								{goalSaving ? 'Saving…' : 'Save'}
							</button>
							<button
								onclick={resetGoalForm}
								class="rounded-full bg-muted px-4 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
							>
								Cancel
							</button>
						</div>
					</div>
				{/if}

				{#if goalConversions && goalConversions.goals.length > 0}
					{@const maxConv = Math.max(...goalConversions.goals.map((g: { conversions: number }) => g.conversions), 1)}
					<div class="space-y-1">
						{#each goalConversions.goals as goal (goal.id)}
							<div class="group relative rounded transition-colors hover:bg-muted/30">
								<div
									class="absolute inset-y-0 left-0 rounded bg-muted/50"
									style="width: {(goal.conversions / maxConv) * 100}%"
								></div>
								<div class="relative flex items-center justify-between px-3 py-1.5 text-sm">
									<div class="flex items-center gap-2">
										<Icon
											icon={goal.goal_type === 'event' ? 'solar:bolt-linear' : 'solar:link-linear'}
											class="h-3.5 w-3.5 text-muted-foreground"
										/>
										<span>{goal.name}</span>
									</div>
									<div class="flex items-center gap-3">
										<span class="tabular-nums text-muted-foreground">{goal.conversions}</span>
										<span class="w-14 text-right tabular-nums text-xs text-muted-foreground">
											{goal.conversion_rate.toFixed(1)}%
										</span>
										<button
											onclick={() => deleteGoal(goal.id)}
											class="rounded-md p-1 text-muted-foreground opacity-0 transition-all hover:text-red-500 group-hover:opacity-100"
											aria-label="Delete goal"
										>
											<Icon icon="solar:trash-bin-trash-linear" class="h-3.5 w-3.5" />
										</button>
									</div>
								</div>
							</div>
						{/each}
					</div>
					<p class="mt-3 text-xs text-muted-foreground">
						Based on {goalConversions.total_visitors} unique visitors
					</p>
				{:else if siteGoals.length === 0 && !showGoalForm}
					<p class="text-sm text-muted-foreground">
						No goals configured. Add a goal to track conversion rates.
					</p>
				{:else if siteGoals.length > 0 && goalConversions?.goals.length === 0}
					<p class="text-sm text-muted-foreground">
						No conversions yet for the selected period.
					</p>
				{/if}
			</Card.Content>
		</Card.Root>

		{#if overview.performance && overview.performance.sample_count > 0}
			<PerformanceCard performance={overview.performance} />
		{/if}

		{#if chartData.length > 0}
			<Card.Root class="mb-8">
				<Card.Header>
					<div class="flex items-center justify-between">
						<div>
							<Card.Title>Traffic Over Time</Card.Title>
							<Card.Description>Pageviews and unique visitors</Card.Description>
						</div>
						<ToggleGroup.Root
							type="single"
							variant="outline"
							size="sm"
							value={selectedGranularity}
							onValueChange={(v) => {
								if (v) selectedGranularity = v as Granularity;
							}}
						>
							{#each visibleGranularities as g (g.key)}
								<ToggleGroup.Item value={g.key}>{g.label}</ToggleGroup.Item>
							{/each}
						</ToggleGroup.Root>
					</div>
				</Card.Header>
				<Card.Content class="px-0">
					<Chart.Container
						config={chartConfig}
						class="aspect-auto h-[300px] w-full [&_.lc-area-path:nth-child(n+3)]:opacity-20"
					>
						<AreaChart
							data={chartData}
							x="date"
							xScale={scaleBand().padding(0.25)}
							axis="x"
							series={[
								{ key: 'pageviews', label: 'Pageviews', color: 'var(--chart-1)' },
								{ key: 'visitors', label: 'Visitors', color: 'var(--chart-2)' },
								{ key: 'prev_pageviews', label: 'Prev Pageviews', color: 'var(--chart-1)' },
								{ key: 'prev_visitors', label: 'Prev Visitors', color: 'var(--chart-2)' }
							]}
							props={{
								xAxis: {
									format: (d: string) => formatChartDate(d, selectedGranularity)
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
						<div class="flex flex-col items-center justify-center gap-6 sm:flex-row">
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

		{#if (overview?.top_utm_sources?.length ?? 0) > 0 || (overview?.top_utm_mediums?.length ?? 0) > 0 || (overview?.top_utm_campaigns?.length ?? 0) > 0}
			<div class="mb-8 grid grid-cols-1 gap-4 md:grid-cols-3">
				{#if (overview?.top_utm_sources?.length ?? 0) > 0}
					<BarListCard
						title="UTM Sources"
						items={utmSourcesData}
						showSeeAll={(overview?.top_utm_sources?.length ?? 0) > 8}
						onSeeAll={() => (activeDrawer = 'utm_sources')}
					/>
				{/if}
				{#if (overview?.top_utm_mediums?.length ?? 0) > 0}
					<BarListCard
						title="UTM Mediums"
						items={utmMediumsData}
						showSeeAll={(overview?.top_utm_mediums?.length ?? 0) > 8}
						onSeeAll={() => (activeDrawer = 'utm_mediums')}
					/>
				{/if}
				{#if (overview?.top_utm_campaigns?.length ?? 0) > 0}
					<BarListCard
						title="UTM Campaigns"
						items={utmCampaignsData}
						showSeeAll={(overview?.top_utm_campaigns?.length ?? 0) > 8}
						onSeeAll={() => (activeDrawer = 'utm_campaigns')}
					/>
				{/if}
			</div>
		{/if}

		{#if (overview?.top_events?.length ?? 0) > 0}
			<BarListCard
				title="Custom Events"
				items={eventsData}
				showSeeAll={(overview?.top_events?.length ?? 0) > 8}
				onSeeAll={() => (activeDrawer = 'events')}
				formatCount
				class="mb-8"
			>
				{#snippet label(item)}
					<span class="font-mono">{item.label}</span>
				{/snippet}
			</BarListCard>
		{/if}

		{#if overview.top_countries?.length > 0}
			<Card.Root class="mb-8">
				<Card.Header>
					<div class="flex items-center justify-between">
						<Card.Title>Visitors</Card.Title>
						{#if (overview?.top_countries?.length ?? 0) > 8}
							<button
								onclick={() => (activeDrawer = 'countries')}
								class="text-xs text-muted-foreground transition-colors hover:text-foreground"
								>See all →</button
							>
						{/if}
					</div>
				</Card.Header>
				<Card.Content>
					<WorldMap countries={overview.top_countries} />
				</Card.Content>
			</Card.Root>
		{/if}

		{#if (overview?.top_entry_pages?.length ?? 0) > 0 || (overview?.top_exit_pages?.length ?? 0) > 0}
			<div class="mb-8 grid gap-4 md:grid-cols-2">
				{#if (overview?.top_entry_pages?.length ?? 0) > 0}
					<BarListCard
						title="Entry Pages"
						items={entryPagesData}
						showSeeAll={(overview?.top_entry_pages?.length ?? 0) > 8}
						onSeeAll={() => (activeDrawer = 'entry')}
						onItemClick={setFilter('path')}
					/>
				{/if}
				{#if (overview?.top_exit_pages?.length ?? 0) > 0}
					<BarListCard
						title="Exit Pages"
						items={exitPagesData}
						showSeeAll={(overview?.top_exit_pages?.length ?? 0) > 8}
						onSeeAll={() => (activeDrawer = 'exit')}
						onItemClick={setFilter('path')}
					/>
				{/if}
			</div>
		{/if}

		<div class="mb-8 grid grid-cols-1 gap-4 lg:grid-cols-3">
			{#if topPagesData.length > 0}
				<BarListCard
					title="Top Pages"
					items={topPagesData}
					showSeeAll={(overview?.top_pages?.length ?? 0) > 8}
					onSeeAll={() => (activeDrawer = 'pages')}
					onItemClick={setFilter('path')}
					class="lg:col-span-2"
				/>
			{/if}

			{#if devicesPieData.length > 0}
				<PieLegendCard
					title="Devices"
					data={devicesPieData}
					configKey="devices"
					onItemClick={setFilter('device')}
				/>
			{/if}

			{#if topReferrersData.length > 0}
				<BarListCard
					title="Top Referrers"
					items={topReferrersData}
					showSeeAll={(overview?.top_referrers?.length ?? 0) > 8}
					onSeeAll={() => (activeDrawer = 'referrers')}
					onItemClick={setFilter('referrer')}
					class="lg:col-span-2"
				>
					{#snippet label(item)}
						<a
							href="{item.label.includes('://') ? '' : 'https://'}{item.label}"
							target="_blank"
							rel="noopener noreferrer"
							class="flex items-center gap-2 truncate mr-2 hover:underline"
							onclick={(e) => e.stopPropagation()}
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
				<BarListCard
					title="Browsers"
					items={browsersData}
					showSeeAll={(overview?.top_browsers?.length ?? 0) > 8}
					onSeeAll={() => (activeDrawer = 'browsers')}
					onItemClick={setFilter('browser')}
				/>
			{/if}

			{#if osData.length > 0}
				<BarListCard
					title="Operating Systems"
					items={osData}
					showSeeAll={(overview?.top_os?.length ?? 0) > 8}
					onSeeAll={() => (activeDrawer = 'os')}
					onItemClick={setFilter('os')}
					class="lg:col-span-2"
				/>
			{/if}
		</div>

		<StatsDrawer
			open={activeDrawer !== null}
			onclose={() => (activeDrawer = null)}
			title={drawerConfig().title}
			items={drawerConfig().items}
			onFilter={drawerConfig().filterKey ? applyDrawerFilter : undefined}
		/>
	{/if}
{/if}
