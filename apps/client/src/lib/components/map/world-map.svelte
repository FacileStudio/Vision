<script lang="ts">
	import { Chart, GeoPath, Svg } from 'layerchart';
	import { geoNaturalEarth1 } from 'd3-geo';
	import { feature } from 'topojson-client';
	import type { Topology } from 'topojson-specification';
	import type { FeatureCollection, Geometry } from 'geojson';
	import { getNumericCode } from './country-codes';

	import worldData from 'world-atlas/countries-110m.json';

	type Props = {
		countries: { country: string; count: number }[];
	};

	let { countries }: Props = $props();

	const topology = worldData as unknown as Topology;
	const geojson = feature(
		topology,
		topology.objects.countries
	) as unknown as FeatureCollection<Geometry>;

	let countsByNumeric = $derived(() => {
		const map = new Map<string, number>();
		for (const c of countries) {
			const numeric = getNumericCode(c.country);
			if (numeric) map.set(numeric, c.count);
		}
		return map;
	});

	let maxCount = $derived(Math.max(1, ...countries.map((c) => c.count)));

	function getFill(featureId: string): string {
		const count = countsByNumeric().get(featureId);
		if (!count) return 'var(--muted)';
		const intensity = 0.15 + 0.85 * (count / maxCount);
		return `oklch(${0.95 - intensity * 0.8} 0 0)`;
	}
</script>

<div class="h-[280px] w-full sm:h-[340px]">
	<Chart geo={{ projection: geoNaturalEarth1, fitGeojson: geojson }}>
		<Svg>
			{#each geojson.features as f}
				<GeoPath
					geojson={f}
					fill={getFill(String(f.id))}
					stroke="var(--border)"
					stroke-width={0.5}
				/>
			{/each}
		</Svg>
	</Chart>
</div>
