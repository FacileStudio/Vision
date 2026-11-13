<script lang="ts">
	import type { SiteSummary } from '$lib';
	import { Badge, Button, Card, Sparkline, StatusDot, icons } from '@facile/muse';
	import SiteFavicon from '$lib/components/site-favicon.svelte';
	import { trendPercent } from '$lib/utils/analytics';
	import { fmt } from '$lib/i18n/format';
	import { locale, t } from '$lib/i18n/index.svelte';

	let {
		site,
		showSpace,
		onDelete
	}: { site: SiteSummary; showSpace: boolean; onDelete: (site: SiteSummary) => void } = $props();

	const trend = $derived(trendPercent(site.unique_visitors, site.prev_unique_visitors, locale()));
	const viewsTrend = $derived(trendPercent(site.pageviews, site.prev_pageviews, locale()));
	const series = $derived(site.series.map((p) => p.count));
</script>

<Card class="flex items-center gap-3 sm:gap-4">
	<a
		href="/sites/{site.site_id}"
		class="flex min-w-0 flex-1 flex-wrap items-center gap-x-6 gap-y-3 rounded-fc-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
	>
		<span class="flex min-w-0 flex-1 basis-56 items-center gap-3">
			<SiteFavicon domain={site.domain} name={site.name} class="size-8 shrink-0" />
			<span class="flex min-w-0 flex-col gap-1">
				<span class="flex min-w-0 items-center gap-2">
					<span class="truncate text-fc-sm font-medium text-fc-fg">{site.name}</span>
					{#if site.live > 0}
						<StatusDot tone="success" pulse label={t('sites.live', { count: fmt(site.live) })} class="shrink-0 text-fc-xs" />
					{/if}
				</span>
				<span class="flex min-w-0 items-center gap-2">
					<span class="truncate text-fc-xs text-fc-fg-muted">{site.domain}</span>
					{#if showSpace}<Badge class="shrink-0">{site.workspace_name}</Badge>{/if}
				</span>
			</span>
		</span>

		<span class="flex items-center gap-6">
			<span class="flex w-20 flex-col">
				<span class="text-fc-xs text-fc-fg-muted">{t('stat.visitors')}</span>
				<span class="text-fc-md font-semibold tabular-nums text-fc-fg">{fmt(site.unique_visitors)}</span>
				<span class="text-fc-xs tabular-nums {trend.color}" title={t('sites.vsPrevious')}>{trend.text}</span>
			</span>
			<span class="flex w-20 flex-col">
				<span class="text-fc-xs text-fc-fg-muted">{t('stat.pageviews')}</span>
				<span class="text-fc-md font-semibold tabular-nums text-fc-fg">{fmt(site.pageviews)}</span>
				<span class="text-fc-xs tabular-nums {viewsTrend.color}" title={t('sites.vsPrevious')}>{viewsTrend.text}</span>
			</span>
			<Sparkline
				data={series}
				height={36}
				class="w-24 sm:w-32"
				aria-label={t('sites.trendAria', { name: site.name })}
				emptyLabel={t('sites.noTraffic')}
			/>
		</span>
	</a>
	<Button
		variant="ghost-danger"
		icon={icons.remove}
		aria-label={t('sites.deleteAria', { name: site.name })}
		class="shrink-0"
		onclick={() => onDelete(site)}
	>
		<span class="hidden lg:inline">{t('common.delete')}</span>
	</Button>
</Card>
