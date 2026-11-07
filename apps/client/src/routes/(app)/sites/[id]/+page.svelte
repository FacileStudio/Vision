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
	import type { AppKey } from '$lib/i18n/messages';
	import WorldMap from '$lib/components/map/world-map.svelte';
	import StatsDrawer from '$lib/components/stats-drawer.svelte';
	import SiteFavicon from '$lib/components/site-favicon.svelte';
	import BarList from '$lib/components/bar-list.svelte';
	import {
		type RangeKey,
		type Granularity,
		rangeKeys,
		rangeDates,
		trendPercent,
		decimal,
		classifyReferrer,
		allowedGranularities,
		defaultGranularity,
		formatChartDate,
		mergeByLabel
	} from '$lib/utils/analytics';
	import { country, dateRange, fmt, hourLabel } from '$lib/i18n/format';
	import { locale, t } from '$lib/i18n/index.svelte';

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

	const filterFields = $derived([
		{ key: 'country', label: t('filters.country'), placeholder: 'US' },
		{ key: 'browser', label: t('filters.browser'), placeholder: 'Chrome' },
		{ key: 'os', label: t('filters.os'), placeholder: 'macOS' },
		{ key: 'device', label: t('filters.device'), placeholder: 'Desktop' },
		{ key: 'path', label: t('filters.path'), placeholder: '/blog' },
		{ key: 'referrer', label: t('filters.referrer'), placeholder: 'google.com' }
	]);
	const rangeTabs = $derived(rangeKeys.map((key) => ({ id: key, label: t(`range.${key}`) })));

	const activeFilterCount = $derived(Object.values(activeFilters).filter(Boolean).length);
	const visibleGranularities = $derived(allowedGranularities(selectedRange));
	const dateRangeLabel = $derived.by(() => {
		const { from, to } = rangeDates(selectedRange, customFrom, customTo);
		return dateRange(from, to);
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
		trendPercent(overview?.total_pageviews ?? 0, overview?.prev_total_pageviews ?? 0, locale())
	);
	const visitorsTrend = $derived(
		trendPercent(overview?.unique_visitors ?? 0, overview?.prev_unique_visitors ?? 0, locale())
	);
	const vpvTrend = $derived(trendPercent(viewsPerVisitor, prevViewsPerVisitor, locale()));
	const bounceTrend = $derived(
		trendPercent(overview?.bounce_rate ?? 0, overview?.prev_bounce_rate ?? 0, locale())
	);

	const chartLabels = $derived(
		(overview?.pageviews_per_day ?? []).map((d) =>
			formatChartDate(d.date, selectedGranularity, locale())
		)
	);

	/*
	 * Two series, not four. The previous period used to ride along as a pair of ghost areas;
	 * muse charts assign colour by series index and give every series equal weight, so the
	 * comparison lives on the stat cards' deltas instead of as half-visible lines nobody
	 * could read the axis against.
	 */
	const trafficSeries = $derived([
		{ name: t('stat.pageviews'), data: (overview?.pageviews_per_day ?? []).map((d) => d.count) },
		{
			name: t('stat.visitors'),
			data: (overview?.pageviews_per_day ?? []).map(
				(d) => overview?.unique_visitors_per_day?.find((v) => v.date === d.date)?.count ?? 0
			)
		}
	]);

	const hourlySeries = $derived([
		{ name: t('stat.pageviews'), data: (overview?.hourly_distribution ?? []).map((h) => h.count) }
	]);
	const hourlyLabels = $derived(
		(overview?.hourly_distribution ?? []).map((h) => (h.hour % 3 === 0 ? hourLabel(h.hour) : ''))
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
			{ label: t('source.direct'), value: direct },
			{ label: t('source.search'), value: search },
			{ label: t('source.social'), value: social },
			{ label: t('source.other'), value: other }
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

	const drawers: Record<string, { title: AppKey; items: () => { label: string; count: number }[]; filterKey: string }> =
		{
			pages: {
				title: 'block.pages',
				items: () => all(overview?.top_pages, (d) => d.path),
				filterKey: 'path'
			},
			referrers: {
				title: 'block.referrers',
				items: () => all(overview?.top_referrers, (d) => d.referrer),
				filterKey: 'referrer'
			},
			browsers: {
				title: 'block.browsers',
				items: () => all(overview?.top_browsers, (d) => d.browser),
				filterKey: 'browser'
			},
			os: {
				title: 'block.os',
				items: () => all(overview?.top_os, (d) => d.os),
				filterKey: 'os'
			},
			entry: {
				title: 'block.entry',
				items: () => all(overview?.top_entry_pages, (d) => d.path),
				filterKey: 'path'
			},
			exit: {
				title: 'block.exit',
				items: () => all(overview?.top_exit_pages, (d) => d.path),
				filterKey: 'path'
			},
			countries: {
				title: 'block.countries',
				items: () => all(overview?.top_countries, (d) => d.country),
				filterKey: 'country'
			},
			utm_sources: {
				title: 'block.utmSources',
				items: () => all(overview?.top_utm_sources, (d) => d.value),
				filterKey: ''
			},
			utm_mediums: {
				title: 'block.utmMediums',
				items: () => all(overview?.top_utm_mediums, (d) => d.value),
				filterKey: ''
			},
			utm_campaigns: {
				title: 'block.utmCampaigns',
				items: () => all(overview?.top_utm_campaigns, (d) => d.value),
				filterKey: ''
			},
			events: {
				title: 'block.events',
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
				toast.danger(e instanceof Error ? e.message : t('site.loadFailed'));
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
		toast.success(t('share.created'));
	}

	async function revokeShare() {
		await api.sites.revokeShare(siteId);
		site = await api.sites.get(siteId);
		toast.neutral(t('share.revoked'));
	}

	async function exportCSV() {
		const { from, to } = rangeDates(selectedRange, customFrom, customTo);
		/* An SSO session has no bearer token to send, only a cookie — hence the
		   conditional header and credentials rather than an unconditional
		   `Bearer null`. */
		const token = getToken();
		const res = await fetch(`/api/analytics/${siteId}/export?from=${from}&to=${to}&format=csv`, {
			headers: token ? { Authorization: `Bearer ${token}` } : {},
			credentials: 'same-origin'
		});
		if (!res.ok) {
			toast.danger(t('site.exportFailed'));
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
			toast.success(t('goals.added'));
			await Promise.all([loadGoals(), refreshGoalConversions()]);
		} catch (e) {
			goalError = e instanceof Error ? e.message : t('goals.saveFailed');
		} finally {
			goalSaving = false;
		}
	}

	async function deleteGoal() {
		const target = goalDeleteTarget;
		if (!target) return;
		await api.goals.delete(target.id);
		goalDeleteTarget = null;
		toast.neutral(t('goals.deleted'));
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
					label={live ? t('site.live') : t('site.reconnecting')}
					pulse={!live}
				/>
				<Button variant="ghost" icon={icons.download} onclick={exportCSV}>{t('site.export')}</Button>
			</div>
		</div>

		<section class="flex flex-col gap-4">
			<div class="flex flex-col gap-1">
				<h2 class="text-fc-lg font-semibold text-fc-fg">{t('tracking.title')}</h2>
				<p class="text-fc-sm text-fc-fg-muted">{t('tracking.text')}</p>
			</div>
			<Card class="flex flex-col gap-4">
				<SecretField value={trackingSnippet} sensitive={false} label={t('tracking.script')} />

				<SettingsRow
					label={t('share.title')}
					description={site.share_token ? t('share.descriptionOn') : t('share.descriptionOff')}
					stacked={Boolean(site.share_token)}
				>
					{#if site.share_token}
						<div class="flex w-full flex-col gap-2">
							<SecretField value={shareUrl} sensitive={false} label={t('share.link')} class="w-full" />
							<Button
								variant="ghost-danger"
								icon={icons.remove}
								class="self-start"
								onclick={() => (revokeShareOpen = true)}
							>
								{t('share.revoke')}
							</Button>
						</div>
					{:else}
						<Button variant="secondary" icon={icons.globe} onclick={generateShare}>
							{t('share.create')}
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
							items={rangeTabs}
							value={selectedRange}
							onChange={selectRange}
							label={t('range.label')}
						/>
					</div>
					<Button
						variant={activeFilterCount > 0 ? 'primary' : 'outline'}
						icon={icons.filter}
						onclick={() => (filtersOpen = true)}
					>
						{activeFilterCount > 0
							? t('filters.buttonCount', { count: activeFilterCount })
							: t('filters.title')}
					</Button>
				</div>

				{#if selectedRange === 'custom'}
					<div class="flex flex-wrap items-end gap-3">
						<Field label={t('range.from')}>
							<Input bind:value={customFrom} type="date" />
						</Field>
						<Field label={t('range.to')}>
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
							{t('filters.clear')}
						</Button>
					{/if}
				</div>

				<div class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-5">
					<StatCard label={t('stat.pageviews')} value={fmt(overview.total_pageviews)}>
						<span class="text-fc-xs {pageviewsTrend.color}">{pageviewsTrend.text}</span>
					</StatCard>
					<StatCard label={t('stat.visitors')} value={fmt(overview.unique_visitors)}>
						<span class="text-fc-xs {visitorsTrend.color}">{visitorsTrend.text}</span>
					</StatCard>
					<StatCard label={t('stat.viewsPerVisitor')} value={decimal(viewsPerVisitor, locale())}>
						<span class="text-fc-xs {vpvTrend.color}">{vpvTrend.text}</span>
					</StatCard>
					<StatCard label={t('stat.bounceRate')} value="{decimal(overview.bounce_rate, locale())}%">
						<span class="text-fc-xs {bounceTrend.color}">{bounceTrend.text}</span>
					</StatCard>
					<StatCard label={t('stat.activeNow')} value={fmt(realtimeCount)}>
						<StatusDot tone="success" label={t('stat.rightNow')} pulse class="text-fc-xs" />
					</StatCard>
				</div>
			</section>

			<section class="flex flex-col gap-4">
				<div class="flex flex-wrap items-end justify-between gap-3">
					<div class="flex flex-col gap-1">
						<h2 class="text-fc-lg font-semibold text-fc-fg">{t('traffic.title')}</h2>
						<p class="text-fc-sm text-fc-fg-muted">{t('traffic.text')}</p>
					</div>
					<Select
						value={selectedGranularity}
						aria-label={t('granularity.label')}
						class="w-40"
						onchange={(e) => (selectedGranularity = e.currentTarget.value as Granularity)}
					>
						{#each visibleGranularities as key (key)}
							<option value={key}>{t(`granularity.${key}`)}</option>
						{/each}
					</Select>
				</div>

				<Card class="flex flex-col gap-4">
					<LineChart series={trafficSeries} labels={chartLabels} area height={280} />
				</Card>

				<div class="grid gap-4 lg:grid-cols-2">
					{#if trafficSources.length > 0}
						<Card class="flex flex-col gap-4">
							<p class="text-fc-sm font-medium text-fc-fg">{t('traffic.sources')}</p>
							<DonutChart
								data={trafficSources}
								centerLabel={t('traffic.centerLabel')}
								centerValue={fmt(overview.total_pageviews)}
								class="flex-1"
							/>
						</Card>
					{/if}

					{#if hourlySeries[0].data.length > 0}
						<Card class="flex flex-col gap-4">
							<p class="text-fc-sm font-medium text-fc-fg">{t('traffic.byHour')}</p>
							<BarChart series={hourlySeries} labels={hourlyLabels} height={220} />
						</Card>
					{/if}
				</div>
			</section>

			<section class="flex flex-col gap-4">
				<div class="flex flex-wrap items-start justify-between gap-3">
					<div class="flex flex-col gap-1">
						<h2 class="text-fc-lg font-semibold text-fc-fg">{t('goals.title')}</h2>
						<p class="text-fc-sm text-fc-fg-muted">
							{goalConversions
								? t('goals.measured', { count: goalConversions.total_visitors })
								: t('goals.hint')}
						</p>
					</div>
					<Button icon={icons.plus} onclick={openGoal}>{t('goals.add')}</Button>
				</div>

				{#if hasConversions}
					<!-- SettingsRow draws the rule on its own top edge and drops it on the first
					     child, so the list needs no manual Divider and no index. -->
					<Card class="flex flex-col">
						{#each goalConversions?.goals ?? [] as goal (goal.id)}
							<SettingsRow
								label={goal.name}
								description={t('goals.row', {
									count: goal.conversions,
									rate: `${decimal(goal.conversion_rate, locale())}%`
								})}
							>
								<Badge tone={goal.goal_type === 'event' ? 'accent' : 'neutral'}>
									{goal.goal_type === 'event' ? t('goals.typeEvent') : t('goals.typePageview')}
								</Badge>
								<Button
									variant="ghost-danger"
									icon={icons.remove}
									aria-label={t('goals.deleteAria', { name: goal.name })}
									onclick={() => {
										goalDeleteTarget = { id: goal.id, name: goal.name };
										goalDeleteOpen = true;
									}}
								>
									{t('common.delete')}
								</Button>
							</SettingsRow>
						{/each}
					</Card>
				{:else if siteGoals.length > 0}
					<Alert tone="info">{t('goals.noConversions')}</Alert>
				{:else}
					<Alert tone="neutral">{t('goals.empty')}</Alert>
				{/if}
			</section>

			{#if overview.performance && overview.performance.sample_count > 0}
				{@const perf = overview.performance}
				<section class="flex flex-col gap-4">
					<div class="flex flex-col gap-1">
						<h2 class="text-fc-lg font-semibold text-fc-fg">{t('perf.title')}</h2>
						<p class="text-fc-sm text-fc-fg-muted">{t('perf.text', { count: perf.sample_count })}</p>
					</div>
					<Card class="flex flex-col gap-4">
						<BarChart
							series={[
								{
									name: t('perf.series'),
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
							yFormat={(n) => t('perf.unit', { value: n })}
						/>
					</Card>
				</section>
			{/if}

			{#if (overview.top_countries?.length ?? 0) > 0}
				<section class="flex flex-col gap-4">
					<div class="flex flex-wrap items-start justify-between gap-3">
						<div class="flex flex-col gap-1">
							<h2 class="text-fc-lg font-semibold text-fc-fg">{t('map.title')}</h2>
							<p class="text-fc-sm text-fc-fg-muted">{t('map.text')}</p>
						</div>
						{#if (overview.top_countries?.length ?? 0) > 8}
							<Button variant="ghost" iconRight={icons.arrow} onclick={() => openDrawer('countries')}>
								{t('common.seeAll')}
							</Button>
						{/if}
					</div>
					<Card>
						<WorldMap countries={overview.top_countries} />
					</Card>
				</section>
			{/if}

			<section class="flex flex-col gap-4">
				<h2 class="text-fc-lg font-semibold text-fc-fg">{t('content.title')}</h2>
				<div class="grid gap-4 lg:grid-cols-3">
					<BarList
						title={t('block.pages')}
						items={topPages}
						showSeeAll={(overview.top_pages?.length ?? 0) > 8}
						onSeeAll={() => openDrawer('pages')}
						onItemClick={setFilter('path')}
						class="lg:col-span-2"
					/>
					{#if devices.length > 0}
						<Card class="flex flex-col gap-4">
							<p class="text-fc-sm font-medium text-fc-fg">{t('block.devices')}</p>
							<DonutChart data={devices} class="flex-1" />
						</Card>
					{/if}

					<BarList
						title={t('block.referrers')}
						items={topReferrers}
						showSeeAll={(overview.top_referrers?.length ?? 0) > 8}
						onSeeAll={() => openDrawer('referrers')}
						onItemClick={setFilter('referrer')}
						class="lg:col-span-2"
					/>
					{#if screens.length > 0}
						<Card class="flex flex-col gap-4">
							<p class="text-fc-sm font-medium text-fc-fg">{t('block.screens')}</p>
							<DonutChart data={screens} class="flex-1" />
						</Card>
					{/if}

					{#if entryPages.length > 0}
						<BarList
							title={t('block.entry')}
							items={entryPages}
							showSeeAll={(overview.top_entry_pages?.length ?? 0) > 8}
							onSeeAll={() => openDrawer('entry')}
							onItemClick={setFilter('path')}
						/>
					{/if}
					{#if exitPages.length > 0}
						<BarList
							title={t('block.exit')}
							items={exitPages}
							showSeeAll={(overview.top_exit_pages?.length ?? 0) > 8}
							onSeeAll={() => openDrawer('exit')}
							onItemClick={setFilter('path')}
						/>
					{/if}

					<BarList
						title={t('block.browsers')}
						items={browsers}
						showSeeAll={(overview.top_browsers?.length ?? 0) > 8}
						onSeeAll={() => openDrawer('browsers')}
						onItemClick={setFilter('browser')}
					/>
					<BarList
						title={t('block.os')}
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
					<h2 class="text-fc-lg font-semibold text-fc-fg">{t('campaigns.title')}</h2>
					<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
						{#if utmSources.length > 0}
							<BarList
								title={t('block.utmSources')}
								items={utmSources}
								showSeeAll={(overview.top_utm_sources?.length ?? 0) > 8}
								onSeeAll={() => openDrawer('utm_sources')}
							/>
						{/if}
						{#if utmMediums.length > 0}
							<BarList
								title={t('block.utmMediums')}
								items={utmMediums}
								showSeeAll={(overview.top_utm_mediums?.length ?? 0) > 8}
								onSeeAll={() => openDrawer('utm_mediums')}
							/>
						{/if}
						{#if utmCampaigns.length > 0}
							<BarList
								title={t('block.utmCampaigns')}
								items={utmCampaigns}
								showSeeAll={(overview.top_utm_campaigns?.length ?? 0) > 8}
								onSeeAll={() => openDrawer('utm_campaigns')}
							/>
						{/if}
						{#if events.length > 0}
							<BarList
								title={t('block.events')}
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

<Drawer bind:open={filtersOpen} title={t('filters.title')} description={t('filters.description')} showClose>
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
				{t('common.clear')}
			</Button>
			<Button size="lg" icon={icons.check} class="flex-1" onclick={() => (filtersOpen = false)}>
				{t('common.done')}
			</Button>
		</div>
	{/snippet}
</Drawer>

<Drawer bind:open={goalOpen} title={t('goals.newTitle')} description={t('goals.newDescription')} showClose>
	<form id="goal-form" class="flex flex-col gap-4" onsubmit={saveGoal}>
		<Field label={t('common.name')} helper={t('goals.nameHelper')}>
			<Input bind:value={goalName} placeholder={t('goals.namePlaceholder')} required disabled={goalSaving} />
		</Field>

		<Field label={t('goals.type')}>
			<Select bind:value={goalType} disabled={goalSaving}>
				<option value="pageview">{t('goals.typePageviewOption')}</option>
				<option value="event">{t('goals.typeEventOption')}</option>
			</Select>
		</Field>

		{#if goalType === 'pageview'}
			<Field label={t('goals.path')}>
				<Input bind:value={goalPagePath} placeholder="/pricing" disabled={goalSaving} />
			</Field>
			<Field label={t('goals.match')}>
				<Select bind:value={goalMatchType} disabled={goalSaving}>
					<option value="exact">{t('goals.matchExact')}</option>
					<option value="starts_with">{t('goals.matchStarts')}</option>
					<option value="contains">{t('goals.matchContains')}</option>
				</Select>
			</Field>
		{:else}
			<Field label={t('goals.eventName')} helper={t('goals.eventHelper')}>
				<Input bind:value={goalEventName} placeholder="signup" disabled={goalSaving} class="font-fc-mono" />
			</Field>
		{/if}

		{#if goalError}
			<Alert tone="danger" title={t('goals.notSaved')}>{goalError}</Alert>
		{/if}
	</form>

	{#snippet footer()}
		<div class="flex gap-2">
			<Button variant="ghost" size="lg" class="flex-1" disabled={goalSaving} onclick={() => (goalOpen = false)}>
				{t('common.cancel')}
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
				{goalSaving ? t('common.saving') : t('goals.add')}
			</Button>
		</div>
	{/snippet}
</Drawer>

<StatsDrawer
	bind:open={drawerOpen}
	title={activeDrawer ? t(activeDrawer.title) : ''}
	items={activeDrawer?.items() ?? []}
	onFilter={activeDrawer?.filterKey ? setFilter(activeDrawer.filterKey) : undefined}
/>

<ConfirmModal
	bind:open={goalDeleteOpen}
	tone="danger"
	title={t('goals.deleteTitle', { name: goalDeleteTarget?.name ?? t('goals.thisGoal') })}
	description={t('goals.deleteBody')}
	confirmLabel={t('goals.deleteConfirm')}
	cancelLabel={t('common.keepIt')}
	onConfirm={deleteGoal}
	onCancel={() => (goalDeleteTarget = null)}
/>

<ConfirmModal
	bind:open={revokeShareOpen}
	tone="danger"
	title={t('share.revokeTitle')}
	description={t('share.revokeBody')}
	confirmLabel={t('share.revoke')}
	cancelLabel={t('common.keepIt')}
	onConfirm={revokeShare}
/>
