<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { Site, AnalyticsOverview } from '$lib';
	import { LineChart, BarChart } from 'layerchart';
	import { scaleBand } from 'd3-scale';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { Copy, Check } from '@lucide/svelte';
	import WorldMap from '$lib/components/map/world-map.svelte';

	let site = $state<Site | null>(null);
	let overview = $state<AnalyticsOverview | null>(null);
	let live = $state(false);
	let copied = $state(false);
	let pollTimer: ReturnType<typeof setInterval> | null = null;

	const pageviewsConfig = {
		pageviews: { label: 'Pageviews', color: 'var(--foreground)' }
	} satisfies Chart.ChartConfig;

	const topPagesConfig = {
		count: { label: 'Views', color: 'var(--foreground)' }
	} satisfies Chart.ChartConfig;

	const referrersConfig = {
		count: { label: 'Referrals', color: 'var(--foreground)' }
	} satisfies Chart.ChartConfig;

	function trackingSnippet(): string {
		return `<script defer src="${page.url.origin}/s.js"><\/script>`;
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

	let browsersData = $derived(overview?.top_browsers ?? []);
	let osData = $derived(overview?.top_os ?? []);
	let devicesData = $derived(overview?.top_devices ?? []);

	let maxBrowserCount = $derived(Math.max(...browsersData.map((d) => d.count), 1));
	let maxOsCount = $derived(Math.max(...osData.map((d) => d.count), 1));
	let maxDeviceCount = $derived(Math.max(...devicesData.map((d) => d.count), 1));
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
				>&lt;script defer src="{page.url.origin}/s.js"&gt;&lt;/script&gt;</pre
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
						<LineChart
							data={pageviewChartData}
							x="date"
							xScale={scaleBand().padding(0.25)}
							axis="x"
							series={[
								{
									key: 'pageviews',
									label: 'Pageviews',
									color: 'var(--foreground)'
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
						</LineChart>
					</Chart.Container>
				</Card.Content>
			</Card.Root>
		{/if}

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

		<div class="grid gap-4 grid-cols-1 md:grid-cols-3 mb-8">
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

			{#if devicesData.length > 0}
				<Card.Root>
					<Card.Header>
						<Card.Title>Devices</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="space-y-1">
							{#each devicesData as item}
								<div class="relative">
									<div
										class="absolute inset-y-0 left-0 rounded bg-muted"
										style="width: {(item.count / maxDeviceCount) * 100}%"
									></div>
									<div class="relative flex justify-between px-3 py-1.5 text-sm">
										<span>{item.device}</span>
										<span class="text-muted-foreground tabular-nums">{item.count}</span>
									</div>
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
			{/if}
		</div>

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
										color: 'var(--foreground)'
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
										color: 'var(--foreground)'
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

	{/if}
{/if}
