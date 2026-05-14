<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { Site, AnalyticsOverview } from '$lib';
	import { AreaChart, BarChart } from 'layerchart';
	import { scaleBand } from 'd3-scale';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { Copy, Check } from '@lucide/svelte';

	interface LiveEvent {
		site_id: number;
		path: string;
		referrer: string;
		country: string;
		visitor_id: string;
		timestamp: string;
	}

	let site = $state<Site | null>(null);
	let overview = $state<AnalyticsOverview | null>(null);
	let liveConnected = $state(false);
	let recentEvents = $state<LiveEvent[]>([]);
	let eventSource: EventSource | null = null;
	let copied = $state(false);

	const pageviewsConfig = {
		pageviews: { label: 'Pageviews', color: 'var(--chart-1)' }
	} satisfies Chart.ChartConfig;

	const topPagesConfig = {
		count: { label: 'Views', color: 'var(--chart-2)' }
	} satisfies Chart.ChartConfig;

	const referrersConfig = {
		count: { label: 'Referrals', color: 'var(--chart-3)' }
	} satisfies Chart.ChartConfig;

	function trackingSnippet(): string {
		return `<script defer src="${page.url.origin}/t.js"><\/script>`;
	}

	async function copySnippet() {
		await navigator.clipboard.writeText(trackingSnippet());
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}

	function handleLiveEvent(event: LiveEvent) {
		if (!overview) return;

		overview.total_pageviews += 1;

		const knownVisitors = new Set<string>();
		knownVisitors.add(event.visitor_id);
		if (knownVisitors.size === 1) {
			overview.unique_visitors += 1;
		}

		const pageEntry = overview.top_pages.find((p) => p.path === event.path);
		if (pageEntry) {
			pageEntry.count += 1;
			overview.top_pages = [...overview.top_pages].sort((a, b) => b.count - a.count);
		} else {
			overview.top_pages = [...overview.top_pages, { path: event.path, count: 1 }].sort(
				(a, b) => b.count - a.count
			);
		}

		if (event.referrer) {
			const refEntry = overview.top_referrers.find((r) => r.referrer === event.referrer);
			if (refEntry) {
				refEntry.count += 1;
				overview.top_referrers = [...overview.top_referrers].sort((a, b) => b.count - a.count);
			} else {
				overview.top_referrers = [
					...overview.top_referrers,
					{ referrer: event.referrer, count: 1 }
				].sort((a, b) => b.count - a.count);
			}
		}

		if (event.country) {
			const countryEntry = overview.top_countries.find((c) => c.country === event.country);
			if (countryEntry) {
				countryEntry.count += 1;
				overview.top_countries = [...overview.top_countries].sort((a, b) => b.count - a.count);
			} else {
				overview.top_countries = [
					...overview.top_countries,
					{ country: event.country, count: 1 }
				].sort((a, b) => b.count - a.count);
			}
		}

		const today = new Date().toISOString().split('T')[0];
		const todayEntry = overview.pageviews_per_day.find((d) => d.date === today);
		if (todayEntry) {
			todayEntry.count += 1;
			overview.pageviews_per_day = [...overview.pageviews_per_day];
		} else {
			overview.pageviews_per_day = [...overview.pageviews_per_day, { date: today, count: 1 }];
		}

		recentEvents = [event, ...recentEvents].slice(0, 20);
	}

	function connectSSE(siteId: number) {
		const url = api.events.liveUrl(siteId);
		eventSource = new EventSource(url);

		eventSource.onopen = () => {
			liveConnected = true;
		};

		eventSource.onmessage = (e) => {
			try {
				const event: LiveEvent = JSON.parse(e.data);
				handleLiveEvent(event);
			} catch {}
		};

		eventSource.onerror = () => {
			liveConnected = false;
		};
	}

	onMount(() => {
		const id = Number(page.params.id);

		(async () => {
			site = await api.sites.get(id);
			overview = await api.analytics.overview(id);
			connectSSE(id);
		})();

		return () => {
			if (eventSource) {
				eventSource.close();
				eventSource = null;
			}
		};
	});

	function formatTime(timestamp: string): string {
		try {
			return new Date(timestamp).toLocaleTimeString();
		} catch {
			return timestamp;
		}
	}

	let pageviewChartData = $derived(
		(overview?.pageviews_per_day ?? []).map((d) => ({
			date: d.date,
			pageviews: d.count
		}))
	);

	let topPagesData = $derived((overview?.top_pages ?? []).slice(0, 8));
	let topReferrersData = $derived((overview?.top_referrers ?? []).slice(0, 8));
</script>

{#if site}
	<div class="mb-8 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold">{site.name}</h1>
			<p class="text-muted-foreground">{site.domain}</p>
		</div>
		<div class="flex items-center gap-2 text-sm">
			{#if liveConnected}
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
				>&lt;script defer src="{page.url.origin}/t.js"&gt;&lt;/script&gt;</pre
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
		<div class="grid gap-4 md:grid-cols-2 mb-8">
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Total Pageviews</p>
				<p class="text-3xl font-bold">{overview.total_pageviews.toLocaleString()}</p>
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Unique Visitors</p>
				<p class="text-3xl font-bold">{overview.unique_visitors.toLocaleString()}</p>
			</div>
		</div>

		{#if pageviewChartData.length > 1}
			<Card.Root class="mb-8">
				<Card.Header>
					<Card.Title>Pageviews Over Time</Card.Title>
					<Card.Description>Daily pageview trend</Card.Description>
				</Card.Header>
				<Card.Content>
					<Chart.Container config={pageviewsConfig} class="min-h-[300px] w-full">
						<AreaChart
							data={pageviewChartData}
							x="date"
							xScale={scaleBand().padding(0.25)}
							axis="x"
							series={[
								{
									key: 'pageviews',
									label: 'Pageviews',
									color: 'var(--chart-1)'
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

		<div class="grid gap-4 md:grid-cols-2 mb-8">
			{#if topPagesData.length > 0}
				<Card.Root>
					<Card.Header>
						<Card.Title>Top Pages</Card.Title>
					</Card.Header>
					<Card.Content>
						<Chart.Container config={topPagesConfig} class="min-h-[250px] w-full">
							<BarChart
								data={topPagesData}
								x="count"
								y="path"
								yScale={scaleBand().padding(0.3)}
								axis="y"
								orientation="horizontal"
								series={[
									{
										key: 'count',
										label: 'Views',
										color: 'var(--chart-2)'
									}
								]}
							>
								{#snippet tooltip()}
									<Chart.Tooltip />
								{/snippet}
							</BarChart>
						</Chart.Container>
					</Card.Content>
				</Card.Root>
			{/if}

			{#if topReferrersData.length > 0}
				<Card.Root>
					<Card.Header>
						<Card.Title>Top Referrers</Card.Title>
					</Card.Header>
					<Card.Content>
						<Chart.Container config={referrersConfig} class="min-h-[250px] w-full">
							<BarChart
								data={topReferrersData}
								x="count"
								y="referrer"
								yScale={scaleBand().padding(0.3)}
								axis="y"
								orientation="horizontal"
								series={[
									{
										key: 'count',
										label: 'Referrals',
										color: 'var(--chart-3)'
									}
								]}
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

		{#if overview.top_countries.length > 0}
			<div class="mb-8">
				<h2 class="font-semibold mb-3">Top Countries</h2>
				<div class="space-y-1 rounded-lg border p-4">
					{#each overview.top_countries as entry}
						<div class="flex justify-between rounded px-3 py-2 text-sm hover:bg-muted">
							<span>{entry.country}</span>
							<span class="text-muted-foreground">{entry.count}</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	{/if}

	{#if recentEvents.length > 0}
		<div class="mb-8">
			<h2 class="font-semibold mb-3">Live Feed</h2>
			<div class="space-y-1 rounded-lg border p-4 max-h-80 overflow-y-auto">
				{#each recentEvents as event}
					<div class="flex items-center gap-3 rounded px-3 py-2 text-sm hover:bg-muted">
						<span class="h-1.5 w-1.5 shrink-0 rounded-full bg-green-500"></span>
						<span class="font-mono truncate flex-1">{event.path}</span>
						{#if event.country}
							<span class="text-muted-foreground text-xs">{event.country}</span>
						{/if}
						<span class="text-muted-foreground text-xs shrink-0"
							>{formatTime(event.timestamp)}</span
						>
					</div>
				{/each}
			</div>
		</div>
	{/if}
{/if}
