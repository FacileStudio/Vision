<script lang="ts">
	import { Badge, Button, Spinner } from '@facile/muse';
	import type { Anomaly, Explanation } from '$lib/bulle';
	import { DIMENSIONS, formatChange, formatDay, formatMetric, movers, signed, type Dimension } from '$lib/utils/anomalies';
	import { locale, t } from '$lib/i18n/index.svelte';
	import type { AppKey } from '$lib/i18n/messages';
	import ExplanationView from './explanation-view.svelte';

	let {
		anomaly,
		explanation = null,
		explaining = false,
		error = null,
		canExplain = true,
		onExplain
	}: {
		anomaly: Anomaly;
		explanation?: Explanation | null;
		explaining?: boolean;
		error?: AppKey | null;
		canExplain?: boolean;
		onExplain: () => void;
	} = $props();

	const metric = $derived(
		anomaly.metric === 'goal_conversions'
			? t('bulle.metric.goal_conversions', { goal: anomaly.goal_name || `#${anomaly.goal_id ?? ''}` })
			: t(`bulle.metric.${anomaly.metric}`)
	);
	const day = $derived(formatDay(anomaly.day, locale()));
	const change = $derived(formatChange(anomaly, locale()));
	const comparison = $derived(
		change
			? t('bulle.compare', {
					value: formatMetric(anomaly.metric, anomaly.value, locale()),
					baseline: formatMetric(anomaly.metric, anomaly.baseline, locale()),
					change
				})
			: t('bulle.compareNew', {
					value: formatMetric(anomaly.metric, anomaly.value, locale()),
					baseline: formatMetric(anomaly.metric, anomaly.baseline, locale())
				})
	);
	const groups = $derived(
		DIMENSIONS.map((dimension) => ({ dimension, rows: movers(anomaly, dimension) })).filter((group) => group.rows.length > 0)
	);

	function label(dimension: Dimension, value: string): string {
		if (value !== '') return value;
		return dimension === 'referrers' ? t('bulle.direct') : t('bulle.unknown');
	}
</script>

<li id="anomaly-{anomaly.id}" class="flex min-w-0 list-none flex-col gap-3 rounded-fc-md bg-fc-component p-4">
	<div class="flex min-w-0 flex-wrap items-center gap-2">
		<span class="min-w-0 flex-1 text-fc-sm font-medium wrap-anywhere text-fc-fg">{metric}</span>
		<Badge tone={anomaly.direction === 'up' ? 'info' : 'warning'}>{t(`bulle.direction.${anomaly.direction}`)}</Badge>
	</div>
	<p class="m-0 text-fc-xs text-fc-fg-muted">{day}</p>
	<p class="m-0 text-fc-sm wrap-anywhere text-fc-fg">{comparison}</p>

	{#if groups.length > 0}
		<div class="flex flex-col gap-2">
			<h4 class="m-0 text-fc-xs font-medium text-fc-fg-muted">{t('bulle.movers')}</h4>
			{#each groups as group (group.dimension)}
				<div class="flex min-w-0 flex-col gap-1">
					<span class="text-fc-xs text-fc-fg-muted">{t(`bulle.dim.${group.dimension}`)}</span>
					<ul class="m-0 flex list-none flex-col gap-1 p-0">
						{#each group.rows as row, i (i)}
							<li class="flex min-w-0 flex-col text-fc-sm">
								<span class="wrap-anywhere text-fc-fg">{label(group.dimension, row.value)}</span>
								<span class="text-fc-xs text-fc-fg-muted">
									{t('bulle.mover', { count: row.count, delta: signed(row.delta, locale()) })}
								</span>
							</li>
						{/each}
					</ul>
				</div>
			{/each}
		</div>
	{/if}

	{#if explanation}
		<ExplanationView {explanation} />
	{:else if canExplain}
		<Button
			variant="secondary"
			size="lg"
			class="self-start"
			disabled={explaining}
			aria-label={t('bulle.explain.aria', { metric, day })}
			onclick={onExplain}
		>
			{#if explaining}<Spinner size="sm" />{/if}
			{t('bulle.explain')}
		</Button>
	{/if}
	{#if error}
		<p class="m-0 text-fc-sm wrap-anywhere text-fc-danger" role="alert">{t(error)}</p>
	{/if}
</li>
