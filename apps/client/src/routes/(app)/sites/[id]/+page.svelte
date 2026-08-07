<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api, getToken } from '$lib';
	import type { Site, AnalyticsOverview, GoalConversionsResponse, Goal } from '$lib';
	import {
		Alert,
		Badge,
		BarChart,
		Button,
		Card,
		ConfirmModal,
		DonutChart,
		Drawer,
		Field,
		Input,
		LineChart,
		SecretField,
		Select,
		SettingsRow,
		Skeleton,
		StatCard,
		StatusDot,
		Tabs,
		icons,
		toast
	} from '@facile/muse';
	import WorldMap from '$lib/components/map/world-map.svelte';
	import StatsDrawer from '$lib/components/stats-drawer.svelte';
	import SiteFavicon from '$lib/components/site-favicon.svelte';
	import BarList from '$lib/components/bar-list.svelte';
	import {
		type RangeKey,
		type Granularity,
		ranges,
		rangeDates,
		formatDateRange,
		trendPercent,
		fmt,
		classifyReferrer,
		allowedGranularities,
		defaultGranularity,
		formatChartDate,
		mergeByLabel
	} from '$lib/utils/analytics';

	let site = $state<Site | null>(null);
	let overview = $state<AnalyticsOverview | null>(null);
	let goalConversions = $state<GoalConversionsResponse | null>(null);
	let siteGoals = $state<Goal[]>([]);
	let live = $state(false);
	let realtimeCount = $state(0);
	let pollTimer: ReturnType<typeof setInterval> | null = null;
	let realtimeTimer: ReturnType<typeof setInterval> | null = null;

	let selectedRange = $state<RangeKey>('30d');
	let customFrom = $state('');
	let customTo = $state('');
	let selectedGranularity = $state<Granularity>('day');

	let activeFilters = $state<Record<string, string>>({});
	let filtersOpen = $state(false);

	let drawerKey = $state<string | null>(null);
	let drawerOpen = $state(false);

	let goalOpen = $state(false);
	let goalName = $state('');
	let goalType = $state<'pageview' | 'event'>('pageview');
	let goalPagePath = $state('');
	let goalEventName = $state('');
	let goalMatchType = $state('exact');
	let goalSaving = $state(false);
	let goalError = $state('');

	let goalDeleteTarget = $state<{ id: number; name: string } | null>(null);
	let goalDeleteOpen = $state(false);

	let revokeShareOpen = $state(false);

	const siteId = $derived(Number(page.params.id));

	const shareUrl = $derived(site?.share_token ? `${page.url.origin}/share/${site.share_token}` : '');
	const trackingSnippet = $derived(`<script defer src="${page.url.origin}/s.js"></` + 'script>');

	const filterFields = [
		{ key: 'country', label: 'Country', placeholder: 'US' },
		{ key: 'browser', label: 'Browser', placeholder: 'Chrome' },
		{ key: 'os', label: 'OS', placeholder: 'macOS' },
		{ key: 'device', label: 'Device', placeholder: 'Desktop' },
		{ key: 'path', label: 'Path', placeholder: '/blog' },
		{ key: 'referrer', label: 'Referrer', placeholder: 'google.com' }
	];

	const activeFilterCount = $derived(Object.values(activeFilters).filter(Boolean).length);
	const visibleGranularities = $derived(allowedGranularities(selectedRange));
	const dateRangeLabel = $derived.by(() => {
		const { from, to } = rangeDates(selectedRange, customFrom, customTo);
		return formatDateRange(from, to);
	});

	function selectRange(key: string) {
		selectedRange = key as RangeKey;
		selectedGranularity = defaultGranularity(selectedRange);
	}

	function setFilter(key: string) {
		return (value: string) => {
			activeFilters = { ...activeFilters, [key]: value };
		};
	}

	function clearFilters() {
		activeFilters = {};
	}

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
		(overview?.pageviews_per_day ?? []).map((d) => formatChartDate(d.date, selectedGranularity))
	);

	/*
	 * Two series, not four. The previous period used to ride along as a pair of ghost areas;
	 * muse charts assign colour by series index and give every series equal weight, so the
	 * comparison lives on the stat cards' deltas instead of as half-visible lines nobody
	 * could read the axis against.
	 */
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

	/* Every list on this page is the same shape with a different key for the label. Rows are
	   merged by label *before* the top-8 slice — see mergeByLabel: duplicates are both a
	   wrong reading and a fatal `each_key_duplicate`, since these render keyed by label. */
	function top<T extends { count: number }>(list: T[] | undefined, label: (d: T) => string) {
		return mergeByLabel((list ?? []).map((d) => ({ label: label(d), count: d.count }))).slice(0, 8);
	}

	/* The drawer shows the same dimension unsliced, so it needs the same merge. */
	function all<T extends { count: number }>(list: T[] | undefined, label: (d: T) => string) {
		return mergeByLabel((list ?? []).map((d) => ({ label: label(d), count: d.count })));
	}

	const topPages = $derived(top(overview?.top_pages, (d) => d.path));
	const topReferrers = $derived(top(overview?.top_referrers, (d) => d.referrer));
	const browsers = $derived(top(overview?.top_browsers, (d) => d.browser));
	const operatingSystems = $derived(top(overview?.top_os, (d) => d.os));
	const entryPages = $derived(top(overview?.top_entry_pages, (d) => d.path));
	const exitPages = $derived(top(overview?.top_exit_pages, (d) => d.path));
	const utmSources = $derived(top(overview?.top_utm_sources, (d) => d.value));
	const utmMediums = $derived(top(overview?.top_utm_mediums, (d) => d.value));
	const utmCampaigns = $derived(top(overview?.top_utm_campaigns, (d) => d.value));
	const events = $derived(top(overview?.top_events, (d) => d.name));

	const drawers: Record<string, { title: string; items: () => { label: string; count: number }[]; filterKey: string }> =
		{
			pages: {
				title: 'Top pages',
				items: () => all(overview?.top_pages, (d) => d.path),
				filterKey: 'path'
			},
			referrers: {
				title: 'Top referrers',
				items: () => all(overview?.top_referrers, (d) => d.referrer),
				filterKey: 'referrer'
			},
			browsers: {
				title: 'Browsers',
				items: () => all(overview?.top_browsers, (d) => d.browser),
				filterKey: 'browser'
			},
			os: {
				title: 'Operating systems',
				items: () => all(overview?.top_os, (d) => d.os),
				filterKey: 'os'
			},
			entry: {
				title: 'Entry pages',
				items: () => all(overview?.top_entry_pages, (d) => d.path),
				filterKey: 'path'
			},
			exit: {
				title: 'Exit pages',
				items: () => all(overview?.top_exit_pages, (d) => d.path),
				filterKey: 'path'
			},
			countries: {
				title: 'Countries',
				items: () => all(overview?.top_countries, (d) => d.country),
				filterKey: 'country'
			},
			utm_sources: {
				title: 'UTM sources',
				items: () => all(overview?.top_utm_sources, (d) => d.value),
				filterKey: ''
			},
			utm_mediums: {
				title: 'UTM mediums',
				items: () => all(overview?.top_utm_mediums, (d) => d.value),
				filterKey: ''
			},
			utm_campaigns: {
				title: 'UTM campaigns',
				items: () => all(overview?.top_utm_campaigns, (d) => d.value),
				filterKey: ''
			},
			events: {
				title: 'Custom events',
				items: () => all(overview?.top_events, (d) => d.name),
				filterKey: ''
			}
		};

	const activeDrawer = $derived(drawerKey ? drawers[drawerKey] : null);

	function openDrawer(key: string) {
		drawerKey = key;
		drawerOpen = true;
	}

	async function refresh() {
		try {
			const { from, to } = rangeDates(selectedRange, customFrom, customTo);
			overview = await api.analytics.overview(
				siteId,
				from,
				to,
				selectedGranularity,
				activeFilters
			);
			live = true;
		} catch {
			live = false;
		}
	}

	async function fetchRealtime() {
		try {
			realtimeCount = (await api.analytics.realtime.visitors(siteId)).visitors;
		} catch {}
	}

	async function loadGoals() {
		try {
			siteGoals = await api.goals.list(siteId);
		} catch {}
	}

	async function refreshGoalConversions() {
		try {
			const { from, to } = rangeDates(selectedRange, customFrom, customTo);
			goalConversions = await api.goals.conversions(siteId, from, to);
		} catch {}
	}

	onMount(() => {
		(async () => {
			try {
				site = await api.sites.get(siteId);
			} catch (e) {
				toast.danger(e instanceof Error ? e.message : 'Could not load this site.');
				return;
			}
			await Promise.all([refresh(), fetchRealtime(), loadGoals(), refreshGoalConversions()]);
			pollTimer = setInterval(refresh, 5000);
			realtimeTimer = setInterval(fetchRealtime, 10000);
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
		if (siteId && site) {
			refresh();
			refreshGoalConversions();
		}
	});

	async function generateShare() {
		site = await api.sites.share(siteId);
		toast.success('Public link created.');
	}

	async function revokeShare() {
		await api.sites.revokeShare(siteId);
		site = await api.sites.get(siteId);
		toast.neutral('Public link revoked.');
	}

	async function exportCSV() {
		const { from, to } = rangeDates(selectedRange, customFrom, customTo);
		const res = await fetch(`/api/analytics/${siteId}/export?from=${from}&to=${to}&format=csv`, {
			headers: { Authorization: `Bearer ${getToken()}` }
		});
		if (!res.ok) {
			toast.danger('The export failed. Nothing was downloaded.');
			return;
		}
		const url = URL.createObjectURL(await res.blob());
		const a = document.createElement('a');
		a.href = url;
		a.download = 'vision-export.csv';
		a.click();
		URL.revokeObjectURL(url);
	}

	function openGoal() {
		goalName = '';
		goalType = 'pageview';
		goalPagePath = '';
		goalEventName = '';
		goalMatchType = 'exact';
		goalError = '';
		goalOpen = true;
	}

	async function saveGoal(e: Event) {
		e.preventDefault();
		goalSaving = true;
		goalError = '';
		try {
			await api.goals.create({
				site_id: siteId,
				name: goalName.trim(),
				goal_type: goalType,
				event_name: goalType === 'event' ? goalEventName.trim() : undefined,
				page_path: goalType === 'pageview' ? goalPagePath.trim() : undefined,
				match_type: goalType === 'pageview' ? goalMatchType : undefined
			});
			goalOpen = false;
			toast.success('Goal added.');
			await Promise.all([loadGoals(), refreshGoalConversions()]);
		} catch (e) {
			goalError = e instanceof Error ? e.message : 'Could not save the goal.';
		} finally {
			goalSaving = false;
		}
	}

	async function deleteGoal() {
		const target = goalDeleteTarget;
		if (!target) return;
		await api.goals.delete(target.id);
		goalDeleteTarget = null;
		toast.neutral('Goal deleted.');
		await Promise.all([loadGoals(), refreshGoalConversions()]);
	}

	const hasConversions = $derived((goalConversions?.goals.length ?? 0) > 0);
</script>

<svelte:head><title>{site ? `${site.name} — Vision` : 'Vision'}</title></svelte:head>

{#if !site}
	<div class="flex flex-col gap-4">
		<Skeleton class="h-16 w-full rounded-fc-md" />
		<Skeleton class="h-64 w-full rounded-fc-md" />
	</div>
{:else}
	<div class="flex flex-col gap-10">
		<div class="flex flex-wrap items-start justify-between gap-4">
			<div class="flex min-w-0 flex-col gap-2">
				<div class="flex min-w-0 items-center gap-3">
					<SiteFavicon domain={site.domain} name={site.name} class="size-7" />
					<h1 class="truncate text-fc-2xl font-semibold text-fc-fg">{site.name}</h1>
				</div>
				<p class="truncate text-fc-sm text-fc-fg-muted">{site.domain}</p>
			</div>
			<div class="flex flex-wrap items-center gap-3">
				<StatusDot
					tone={live ? 'success' : 'warning'}
					label={live ? 'Live' : 'Reconnecting…'}
					pulse={!live}
				/>
				<Button variant="ghost" icon={icons.download} onclick={exportCSV}>Export</Button>
			</div>
		</div>

		<section class="flex flex-col gap-4">
			<div class="flex flex-col gap-1">
				<h2 class="text-fc-lg font-semibold text-fc-fg">Tracking</h2>
				<p class="text-fc-sm text-fc-fg-muted">
					Paste this in the site's <code class="font-fc-mono">&lt;head&gt;</code>. Nothing else to
					configure.
				</p>
			</div>
			<Card class="flex flex-col gap-4">
				<SecretField value={trackingSnippet} sensitive={false} label="Script tag" />

				<SettingsRow
					label="Public dashboard"
					description={site.share_token
						? 'Anyone with the link can read these numbers — and only these.'
						: 'Create a read-only link you can hand to a client without an account.'}
					stacked={Boolean(site.share_token)}
				>
					{#if site.share_token}
						<div class="flex w-full flex-col gap-2">
							<SecretField value={shareUrl} sensitive={false} label="Share link" class="w-full" />
							<Button
								variant="ghost-danger"
								icon={icons.remove}
								class="self-start"
								onclick={() => (revokeShareOpen = true)}
							>
								Revoke link
							</Button>
						</div>
					{:else}
						<Button variant="outline" icon={icons.globe} onclick={generateShare}>
							Create public link
						</Button>
					{/if}
				</SettingsRow>
			</Card>
		</section>

		{#if !overview}
			<Skeleton class="h-64 w-full rounded-fc-md" />
		{:else}
			<section class="flex flex-col gap-4">
				<div class="flex flex-wrap items-center gap-3">
					<div class="min-w-0 flex-1">
						<Tabs
							items={ranges.map((r) => ({ id: r.key, label: r.label }))}
							value={selectedRange}
							onChange={selectRange}
							label="Date range"
						/>
					</div>
					<Button
						variant={activeFilterCount > 0 ? 'primary' : 'outline'}
						icon={icons.filter}
						onclick={() => (filtersOpen = true)}
					>
						Filters{activeFilterCount > 0 ? ` · ${activeFilterCount}` : ''}
					</Button>
				</div>

				{#if selectedRange === 'custom'}
					<div class="flex flex-wrap items-end gap-3">
						<Field label="From">
							<Input bind:value={customFrom} type="date" />
						</Field>
						<Field label="To">
							<Input bind:value={customTo} type="date" />
						</Field>
					</div>
				{/if}

				<div class="flex flex-wrap items-center gap-2">
					<span class="text-fc-xs text-fc-fg-muted">{dateRangeLabel}</span>
					{#each Object.entries(activeFilters) as [key, value] (key)}
						{#if value}
							<Badge tone="accent">{key}: {value}</Badge>
						{/if}
					{/each}
					{#if activeFilterCount > 0}
						<Button variant="ghost" size="sm" icon={icons.refresh} onclick={clearFilters}>
							Clear filters
						</Button>
					{/if}
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
				<div class="flex flex-wrap items-end justify-between gap-3">
					<div class="flex flex-col gap-1">
						<h2 class="text-fc-lg font-semibold text-fc-fg">Traffic over time</h2>
						<p class="text-fc-sm text-fc-fg-muted">Pageviews and unique visitors.</p>
					</div>
					<Select
						value={selectedGranularity}
						aria-label="Granularity"
						class="w-40"
						onchange={(e) => (selectedGranularity = e.currentTarget.value as Granularity)}
					>
						{#each visibleGranularities as g (g.key)}
							<option value={g.key}>{g.label}</option>
						{/each}
					</Select>
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

			<section class="flex flex-col gap-4">
				<div class="flex flex-wrap items-start justify-between gap-3">
					<div class="flex flex-col gap-1">
						<h2 class="text-fc-lg font-semibold text-fc-fg">Goals</h2>
						<p class="text-fc-sm text-fc-fg-muted">
							{goalConversions
								? `Measured against ${fmt(goalConversions.total_visitors)} unique visitors.`
								: 'Conversion rates for the pages and events that matter.'}
						</p>
					</div>
					<Button icon={icons.plus} onclick={openGoal}>Add goal</Button>
				</div>

				{#if hasConversions}
					<!-- SettingsRow draws the rule on its own top edge and drops it on the first
					     child, so the list needs no manual Divider and no index. -->
					<Card class="flex flex-col">
						{#each goalConversions?.goals ?? [] as goal (goal.id)}
							<SettingsRow
								label={goal.name}
								description="{goal.conversions} conversions · {goal.conversion_rate.toFixed(
									1
								)}% of visitors"
							>
								<Badge tone={goal.goal_type === 'event' ? 'accent' : 'neutral'}>
									{goal.goal_type === 'event' ? 'event' : 'pageview'}
								</Badge>
								<Button
									variant="ghost-danger"
									icon={icons.remove}
									aria-label="Delete {goal.name}"
									onclick={() => {
										goalDeleteTarget = { id: goal.id, name: goal.name };
										goalDeleteOpen = true;
									}}
								>
									Delete
								</Button>
							</SettingsRow>
						{/each}
					</Card>
				{:else if siteGoals.length > 0}
					<Alert tone="info">No conversions in this period yet.</Alert>
				{:else}
					<Alert tone="neutral">
						No goals yet. A goal is a page or a custom event you want a conversion rate for.
					</Alert>
				{/if}
			</section>

			{#if overview.performance && overview.performance.sample_count > 0}
				{@const perf = overview.performance}
				<section class="flex flex-col gap-4">
					<div class="flex flex-col gap-1">
						<h2 class="text-fc-lg font-semibold text-fc-fg">Page load</h2>
						<p class="text-fc-sm text-fc-fg-muted">
							Averaged over {fmt(perf.sample_count)} samples, in milliseconds.
						</p>
					</div>
					<Card class="flex flex-col gap-4">
						<BarChart
							series={[
								{
									name: 'Milliseconds',
									data: [
										perf.avg_dns,
										perf.avg_tcp,
										perf.avg_ttfb,
										perf.avg_dom_load,
										perf.avg_page_load
									].map((n) => Math.round(n))
								}
							]}
							labels={['DNS', 'TCP', 'TTFB', 'DOM', 'Load']}
							height={200}
							yFormat={(n) => `${n} ms`}
						/>
					</Card>
				</section>
			{/if}

			{#if (overview.top_countries?.length ?? 0) > 0}
				<section class="flex flex-col gap-4">
					<div class="flex flex-wrap items-start justify-between gap-3">
						<div class="flex flex-col gap-1">
							<h2 class="text-fc-lg font-semibold text-fc-fg">Where they are</h2>
							<p class="text-fc-sm text-fc-fg-muted">Visitors by country.</p>
						</div>
						{#if (overview.top_countries?.length ?? 0) > 8}
							<Button variant="ghost" iconRight={icons.arrow} onclick={() => openDrawer('countries')}>
								See all
							</Button>
						{/if}
					</div>
					<Card>
						<WorldMap countries={overview.top_countries} />
					</Card>
				</section>
			{/if}

			<section class="flex flex-col gap-4">
				<h2 class="text-fc-lg font-semibold text-fc-fg">Content</h2>
				<div class="grid gap-4 lg:grid-cols-3">
					<BarList
						title="Top pages"
						items={topPages}
						showSeeAll={(overview.top_pages?.length ?? 0) > 8}
						onSeeAll={() => openDrawer('pages')}
						onItemClick={setFilter('path')}
						class="lg:col-span-2"
					/>
					{#if devices.length > 0}
						<Card class="flex flex-col gap-4">
							<p class="text-fc-sm font-medium text-fc-fg">Devices</p>
							<DonutChart data={devices} class="flex-1" />
						</Card>
					{/if}

					<BarList
						title="Top referrers"
						items={topReferrers}
						showSeeAll={(overview.top_referrers?.length ?? 0) > 8}
						onSeeAll={() => openDrawer('referrers')}
						onItemClick={setFilter('referrer')}
						class="lg:col-span-2"
					/>
					{#if screens.length > 0}
						<Card class="flex flex-col gap-4">
							<p class="text-fc-sm font-medium text-fc-fg">Screens</p>
							<DonutChart data={screens} class="flex-1" />
						</Card>
					{/if}

					{#if entryPages.length > 0}
						<BarList
							title="Entry pages"
							items={entryPages}
							showSeeAll={(overview.top_entry_pages?.length ?? 0) > 8}
							onSeeAll={() => openDrawer('entry')}
							onItemClick={setFilter('path')}
						/>
					{/if}
					{#if exitPages.length > 0}
						<BarList
							title="Exit pages"
							items={exitPages}
							showSeeAll={(overview.top_exit_pages?.length ?? 0) > 8}
							onSeeAll={() => openDrawer('exit')}
							onItemClick={setFilter('path')}
						/>
					{/if}

					<BarList
						title="Browsers"
						items={browsers}
						showSeeAll={(overview.top_browsers?.length ?? 0) > 8}
						onSeeAll={() => openDrawer('browsers')}
						onItemClick={setFilter('browser')}
					/>
					<BarList
						title="Operating systems"
						items={operatingSystems}
						showSeeAll={(overview.top_os?.length ?? 0) > 8}
						onSeeAll={() => openDrawer('os')}
						onItemClick={setFilter('os')}
						class="lg:col-span-2"
					/>
				</div>
			</section>

			{#if utmSources.length > 0 || utmMediums.length > 0 || utmCampaigns.length > 0 || events.length > 0}
				<section class="flex flex-col gap-4">
					<h2 class="text-fc-lg font-semibold text-fc-fg">Campaigns and events</h2>
					<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
						{#if utmSources.length > 0}
							<BarList
								title="UTM sources"
								items={utmSources}
								showSeeAll={(overview.top_utm_sources?.length ?? 0) > 8}
								onSeeAll={() => openDrawer('utm_sources')}
							/>
						{/if}
						{#if utmMediums.length > 0}
							<BarList
								title="UTM mediums"
								items={utmMediums}
								showSeeAll={(overview.top_utm_mediums?.length ?? 0) > 8}
								onSeeAll={() => openDrawer('utm_mediums')}
							/>
						{/if}
						{#if utmCampaigns.length > 0}
							<BarList
								title="UTM campaigns"
								items={utmCampaigns}
								showSeeAll={(overview.top_utm_campaigns?.length ?? 0) > 8}
								onSeeAll={() => openDrawer('utm_campaigns')}
							/>
						{/if}
						{#if events.length > 0}
							<BarList
								title="Custom events"
								items={events}
								showSeeAll={(overview.top_events?.length ?? 0) > 8}
								onSeeAll={() => openDrawer('events')}
								class="md:col-span-2 lg:col-span-3"
							>
								{#snippet label(item)}
									<span class="truncate font-fc-mono text-fc-xs text-fc-fg">{item.label}</span>
								{/snippet}
							</BarList>
						{/if}
					</div>
				</section>
			{/if}
		{/if}
	</div>
{/if}

<Drawer bind:open={filtersOpen} title="Filters" description="Narrow every number on this page." showClose>
	<div class="flex flex-col gap-4">
		{#each filterFields as f (f.key)}
			<Field label={f.label}>
				<Input
					value={activeFilters[f.key] ?? ''}
					placeholder={f.placeholder}
					oninput={(e) => (activeFilters = { ...activeFilters, [f.key]: e.currentTarget.value })}
				/>
			</Field>
		{/each}
	</div>

	{#snippet footer()}
		<div class="flex gap-2">
			<Button variant="ghost" size="lg" icon={icons.refresh} class="flex-1" onclick={clearFilters}>
				Clear
			</Button>
			<Button size="lg" icon={icons.check} class="flex-1" onclick={() => (filtersOpen = false)}>
				Done
			</Button>
		</div>
	{/snippet}
</Drawer>

<Drawer bind:open={goalOpen} title="New goal" description="A goal turns a page or an event into a conversion rate." showClose>
	<form id="goal-form" class="flex flex-col gap-4" onsubmit={saveGoal}>
		<Field label="Name" helper="How it appears in the goals list.">
			<Input bind:value={goalName} placeholder="Signup" required disabled={goalSaving} />
		</Field>

		<Field label="Type">
			<Select bind:value={goalType} disabled={goalSaving}>
				<option value="pageview">Pageview — someone reaches a page</option>
				<option value="event">Custom event — the script reports one</option>
			</Select>
		</Field>

		{#if goalType === 'pageview'}
			<Field label="Path">
				<Input bind:value={goalPagePath} placeholder="/pricing" disabled={goalSaving} />
			</Field>
			<Field label="Match">
				<Select bind:value={goalMatchType} disabled={goalSaving}>
					<option value="exact">Exactly this path</option>
					<option value="starts_with">Starts with it</option>
					<option value="contains">Contains it</option>
				</Select>
			</Field>
		{:else}
			<Field label="Event name" helper="The string your script passes to the tracker.">
				<Input bind:value={goalEventName} placeholder="signup" disabled={goalSaving} class="font-fc-mono" />
			</Field>
		{/if}

		{#if goalError}
			<Alert tone="danger" title="Not saved">{goalError}</Alert>
		{/if}
	</form>

	{#snippet footer()}
		<div class="flex gap-2">
			<Button variant="ghost" size="lg" class="flex-1" disabled={goalSaving} onclick={() => (goalOpen = false)}>
				Cancel
			</Button>
			<Button
				type="submit"
				form="goal-form"
				size="lg"
				icon={icons.plus}
				class="flex-1"
				disabled={goalSaving ||
					!goalName.trim() ||
					(goalType === 'pageview' ? !goalPagePath.trim() : !goalEventName.trim())}
			>
				{goalSaving ? 'Saving…' : 'Add goal'}
			</Button>
		</div>
	{/snippet}
</Drawer>

<StatsDrawer
	bind:open={drawerOpen}
	title={activeDrawer?.title ?? ''}
	items={activeDrawer?.items() ?? []}
	onFilter={activeDrawer?.filterKey ? setFilter(activeDrawer.filterKey) : undefined}
/>

<ConfirmModal
	bind:open={goalDeleteOpen}
	tone="danger"
	title="Delete {goalDeleteTarget?.name ?? 'this goal'}?"
	description="The goal stops being measured. Past conversions are not stored separately, so its history goes with it."
	confirmLabel="Delete goal"
	cancelLabel="Keep it"
	onConfirm={deleteGoal}
	onCancel={() => (goalDeleteTarget = null)}
/>

<ConfirmModal
	bind:open={revokeShareOpen}
	tone="danger"
	title="Revoke the public link?"
	description="Anyone holding it gets a 'not found' page from then on, and a new link will be a different URL. Nothing about the site's data changes."
	confirmLabel="Revoke link"
	cancelLabel="Keep it"
	onConfirm={revokeShare}
/>
