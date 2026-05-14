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

	let site = $state<Site | null>(null);
	let overview = $state<AnalyticsOverview | null>(null);
	let live = $state(false);
	let copied = $state(false);
	let pollTimer: ReturnType<typeof setInterval> | null = null;

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

	async function refresh(siteId: number) {
		try {
			overview = await api.analytics.overview(siteId);
			live = true;
		} catch {
			live = false;
		}
	}

	onMount(() => {
		const id = Number(page.params.id);

		(async () => {
			site = await api.sites.get(id);
			await refresh(id);
			pollTimer = setInterval(() => refresh(id), 5000);
		})();

		return () => {
			if (pollTimer) clearInterval(pollTimer);
		};
	});

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

		{#if overview.top_countries?.length > 0}
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
{/if}
