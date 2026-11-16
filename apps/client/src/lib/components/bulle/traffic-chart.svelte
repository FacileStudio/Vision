<script lang="ts">
	import { LineChart, type ChartMarker, type ChartSeries } from '@facile/muse';
	import type { Anomaly } from '$lib/bulle';
	import { bulleApi } from '$lib/bulle-api';
	import { anomalyMarkers, formatChange } from '$lib/utils/anomalies';
	import { locale, t } from '$lib/i18n/index.svelte';

	let {
		siteId,
		from,
		to,
		buckets,
		granularity,
		series,
		labels,
		onPick
	}: {
		siteId: number;
		from: string;
		to: string;
		buckets: string[];
		granularity: string;
		series: ChartSeries[];
		labels: string[];
		onPick?: (anomaly: Anomaly) => void;
	} = $props();

	let anomalies = $state<Anomaly[]>([]);

	$effect(() => {
		const query = { siteId, from, to };
		bulleApi
			.anomalies(query.siteId, query.from, query.to)
			.then((page) => {
				if (query.siteId === siteId && query.from === from && query.to === to) anomalies = page.anomalies;
			})
			.catch(() => (anomalies = []));
	});

	function label(anomaly: Anomaly): string {
		const metric =
			anomaly.metric === 'goal_conversions'
				? t('bulle.metric.goal_conversions', { goal: anomaly.goal_name || `#${anomaly.goal_id ?? ''}` })
				: t(`bulle.metric.${anomaly.metric}`);
		return `${metric} · ${formatChange(anomaly, locale()) || t(`bulle.direction.${anomaly.direction}`)}`;
	}

	const markers = $derived(anomalyMarkers(anomalies, buckets, granularity, label));
	function pick(marker: ChartMarker) {
		const id = (marker as ChartMarker & { id?: number }).id;
		const anomaly = anomalies.find((item) => item.id === id);
		if (anomaly) onPick?.(anomaly);
	}
</script>

<LineChart {series} {labels} {markers} onMarkerClick={onPick ? pick : undefined} area height={280} />
