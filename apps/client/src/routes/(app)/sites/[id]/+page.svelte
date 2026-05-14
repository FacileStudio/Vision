<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { Site, AnalyticsOverview } from '$lib';
	import { LineChart } from 'layerchart';
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

	type RangeKey = 'today' | '7d' | '30d' | '90d';
	let selectedRange = $state<RangeKey>('30d');

	let darkMode = $state(false);

	const ranges: { key: RangeKey; label: string; days: number }[] = [
		{ key: 'today', label: 'Today', days: 0 },
		{ key: '7d', label: '7d', days: 7 },
		{ key: '30d', label: '30d', days: 30 },
		{ key: '90d', label: '90d', days: 90 }
	];

	function rangeDates(key: RangeKey): { from: string; to: string } {
		const to = new Date();
		const toStr = to.toISOString().slice(0, 10);
		if (key === 'today') return { from: toStr, to: toStr };
		const daysMap: Record<RangeKey, number> = { today: 0, '7d': 7, '30d': 30, '90d': 90 };
		const from = new Date(to);
		from.setDate(from.getDate() - daysMap[key]);
		return { from: from.toISOString().slice(0, 10), to: toStr };
	}

	const pageviewsConfig = {
		pageviews: { label: 'Pageviews', color: 'var(--foreground)' }
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
		const id = Number(page.params.id);
		if (id && site) refresh(id);
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
	let screensData = $derived(overview?.top_screens ?? []);

	let maxBrowserCount = $derived(Math.max(...browsersData.map((d) => d.count), 1));
	let maxOsCount = $derived(Math.max(...osData.map((d) => d.count), 1));
	let maxDeviceCount = $derived(Math.max(...devicesData.map((d) => d.count), 1));
	let maxScreenCount = $derived(Math.max(...screensData.map((d) => d.count), 1));
	let maxPageCount = $derived(Math.max(...topPagesData.map((d) => d.count), 1));
	let maxReferrerCount = $derived(Math.max(...topReferrersData.map((d) => d.count), 1));
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
		<div class="mb-6 flex gap-2">
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

		<div class="grid gap-4 grid-cols-1 md:grid-cols-3 mb-8">
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Total Pageviews</p>
				<p class="text-3xl font-bold">{overview.total_pageviews.toLocaleString()}</p>
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Unique Visitors</p>
				<p class="text-3xl font-bold">{overview.unique_visitors.toLocaleString()}</p>
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Active Now</p>
				<p class="text-3xl font-bold">{realtimeCount.toLocaleString()}</p>
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

		<div class="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-4 mb-8">
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

			{#if screensData.length > 0}
				<Card.Root>
					<Card.Header>
						<Card.Title>Screens</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="space-y-1">
							{#each screensData as item}
								<div class="relative">
									<div
										class="absolute inset-y-0 left-0 rounded bg-muted"
										style="width: {(item.count / maxScreenCount) * 100}%"
									></div>
									<div class="relative flex justify-between px-3 py-1.5 text-sm">
										<span>{item.screen}</span>
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

	{/if}
{/if}
