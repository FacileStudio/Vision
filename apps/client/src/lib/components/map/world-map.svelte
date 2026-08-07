<script lang="ts">
	import { geoNaturalEarth1, geoPath } from 'd3-geo';
	import { feature } from 'topojson-client';
	import type { Topology } from 'topojson-specification';
	import type { FeatureCollection, Geometry } from 'geojson';
	import { ChartTooltip, IconButton, icons } from '@facile/muse';
	import { getNumericCode, getAlpha2FromNumeric, getCountryName } from './country-codes';

	import worldData from 'world-atlas/countries-110m.json';

	type Props = {
		countries: { country: string; count: number }[];
	};

	let { countries }: Props = $props();

	const WIDTH = 800;
	const HEIGHT = 400;

	/*
	 * The projection is fitted once against a fixed viewBox and every path string is built
	 * at module scope: the geometry never changes, only the fills do. The svg then scales
	 * with CSS, which is the one place muse's "render at real pixel dimensions" rule does
	 * not apply — a map is a picture, not an axis with type on it.
	 */
	const topology = worldData as unknown as Topology;
	const geojson = feature(
		topology,
		topology.objects.countries
	) as unknown as FeatureCollection<Geometry>;

	const projection = geoNaturalEarth1().fitSize([WIDTH, HEIGHT], geojson);
	const path = geoPath(projection);
	const shapes = geojson.features.map((f) => ({ id: String(f.id), d: path(f) ?? '' }));

	let countsByNumeric = $derived.by(() => {
		const map = new Map<string, number>();
		for (const c of countries) {
			const numeric = getNumericCode(c.country);
			if (numeric) map.set(numeric, c.count);
		}
		return map;
	});

	let maxCount = $derived(Math.max(1, ...countries.map((c) => c.count)));

	let hoveredId = $state<string | null>(null);
	let tooltipX = $state(0);
	let tooltipY = $state(0);
	let tooltipVisible = $state(false);

	let containerEl: HTMLDivElement | undefined = $state();
	let scale = $state(1);
	let panX = $state(0);
	let panY = $state(0);
	let dragging = $state(false);
	let dragStartX = $state(0);
	let dragStartY = $state(0);
	let panStartX = $state(0);
	let panStartY = $state(0);

	let hoveredName = $derived.by(() => {
		if (!hoveredId) return '';
		const a2 = getAlpha2FromNumeric(hoveredId);
		return a2 ? getCountryName(a2) : '';
	});

	let hoveredCount = $derived(hoveredId ? (countsByNumeric.get(hoveredId) ?? 0) : 0);

	/*
	 * A choropleth needs a sequential ramp and the chart tokens are categorical, so the
	 * ramp is one token at varying opacity: `fc-accent` is the foreground ink, so it stays
	 * the darkest thing on a light page and the lightest on a dark one without a second
	 * set of values. Countries with no traffic are `fc-surface` — a fill, not a faint
	 * accent, so "no data" never reads as "a little data".
	 */
	function fillFor(id: string) {
		const count = countsByNumeric.get(id);
		const hovered = hoveredId === id;
		if (!count) {
			return { fill: 'var(--color-fc-surface)', opacity: hovered ? 0.75 : 1 };
		}
		const share = count / maxCount;
		return { fill: 'var(--color-fc-accent)', opacity: Math.min(1, 0.15 + 0.75 * share + (hovered ? 0.15 : 0)) };
	}

	function zoomAt(factor: number, cx: number, cy: number) {
		const newScale = Math.min(8, Math.max(0.5, scale * factor));
		panX += cx * (1 / newScale - 1 / scale);
		panY += cy * (1 / newScale - 1 / scale);
		scale = newScale;
	}

	function handleWheel(e: WheelEvent) {
		e.preventDefault();
		if (!containerEl) return;
		const rect = containerEl.getBoundingClientRect();
		zoomAt(e.deltaY > 0 ? 0.9 : 1.1, e.clientX - rect.left, e.clientY - rect.top);
	}

	function handleMouseDown(e: MouseEvent) {
		dragging = true;
		dragStartX = e.clientX;
		dragStartY = e.clientY;
		panStartX = panX;
		panStartY = panY;
	}

	function handleMouseMove(e: MouseEvent) {
		if (!dragging) return;
		panX = panStartX + (e.clientX - dragStartX) / scale;
		panY = panStartY + (e.clientY - dragStartY) / scale;
	}

	function handleMouseUp() {
		dragging = false;
	}

	function track(e: MouseEvent) {
		if (!containerEl) return;
		const rect = containerEl.getBoundingClientRect();
		tooltipX = e.clientX - rect.left;
		tooltipY = e.clientY - rect.top;
	}

	function resetView() {
		scale = 1;
		panX = 0;
		panY = 0;
	}

	function zoomIn() {
		if (!containerEl) return;
		const rect = containerEl.getBoundingClientRect();
		zoomAt(1.3, rect.width / 2, rect.height / 2);
	}

	function zoomOut() {
		if (!containerEl) return;
		const rect = containerEl.getBoundingClientRect();
		zoomAt(0.7, rect.width / 2, rect.height / 2);
	}

	/*
	 * Resolving a code to a display name is a *collapsing* transform: two distinct codes can
	 * land on one name (an unmapped code falls through to its raw value, and several of those
	 * are the empty string), so the rows have to be merged afterwards or the same country
	 * appears twice with split counts — and this table used to key its `{#each}` on the name,
	 * which turned that into a fatal each_key_duplicate.
	 *
	 * Merge first, then rank and slice, so a country split across two codes is ranked on its
	 * real total rather than on whichever half was larger.
	 */
	const ranked = $derived.by(() => {
		const byName = new Map<string, { name: string; count: number }>();
		for (const c of countries) {
			const a2 = getAlpha2FromNumeric(getNumericCode(c.country) ?? '');
			const name = (a2 ? getCountryName(a2) : c.country) || 'Unknown';
			const found = byName.get(name);
			if (found) found.count += c.count;
			else byName.set(name, { name, count: c.count });
		}
		return [...byName.values()].sort((a, b) => b.count - a.count).slice(0, 20);
	});
</script>

<svelte:window onmouseup={handleMouseUp} onmousemove={handleMouseMove} />

<div
	bind:this={containerEl}
	class="relative h-[300px] w-full overflow-hidden rounded-fc-md sm:h-[400px]"
	style="cursor: {dragging ? 'grabbing' : 'grab'}"
>
	<!-- The svg is decoration; the ranked table below is the accessible representation,
	     the same split every muse chart makes. -->
	<svg
		viewBox="0 0 {WIDTH} {HEIGHT}"
		class="h-full w-full"
		aria-hidden="true"
		onwheel={handleWheel}
		onmousedown={handleMouseDown}
		role="presentation"
	>
		<g transform="translate({panX * scale}, {panY * scale}) scale({scale})">
			<!-- Keyed by position, not by shape.id: three features in world-atlas 110m carry no
			     id at all, so they all stringify to "undefined" and Svelte kills the whole page
			     with each_key_duplicate. `shapes` is built once at module scope and never
			     reordered, so the index IS the stable identity here. -->
			{#each shapes as shape, i (i)}
				{@const paint = fillFor(shape.id)}
				<path
					role="presentation"
					d={shape.d}
					fill={paint.fill}
					fill-opacity={paint.opacity}
					stroke="var(--color-fc-border)"
					stroke-width={0.5 / scale}
					style="transition: fill-opacity 0.15s ease"
					onmouseenter={(e) => {
						hoveredId = shape.id;
						track(e);
						tooltipVisible = true;
					}}
					onmousemove={track}
					onmouseleave={() => {
						hoveredId = null;
						tooltipVisible = false;
					}}
				/>
			{/each}
		</g>
	</svg>

	<div class="sr-only">
		<table>
			<caption>Visitors by country</caption>
			<thead>
				<tr><th scope="col">Country</th><th scope="col">Visitors</th></tr>
			</thead>
			<tbody>
				<!-- Keyed by position: this is a ranked table, and a name is a rendering of the
					     data, not an identity for it. `ranked` merges by name so the two agree —
					     but the index is what stays correct if that merge ever stops. -->
					{#each ranked as row, i (i)}
					<tr><td>{row.name}</td><td>{row.count}</td></tr>
				{/each}
			</tbody>
		</table>
	</div>

	<div class="absolute right-2 top-2 flex flex-col gap-1">
		<IconButton variant="ghost" aria-label="Zoom in" onclick={zoomIn} class="size-9 bg-fc-component">
			<iconify-icon icon={icons.plus} width="16" height="16" class="block"></iconify-icon>
		</IconButton>
		<IconButton variant="ghost" aria-label="Zoom out" onclick={zoomOut} class="size-9 bg-fc-component">
			<iconify-icon icon={icons.minus} width="16" height="16" class="block"></iconify-icon>
		</IconButton>
		<IconButton variant="ghost" aria-label="Reset view" onclick={resetView} class="size-9 bg-fc-component">
			<iconify-icon icon={icons.refresh} width="16" height="16" class="block"></iconify-icon>
		</IconButton>
	</div>

	<ChartTooltip
		x={tooltipX}
		y={tooltipY}
		title={hoveredName}
		rows={hoveredName ? [{ name: 'Visitors', value: String(hoveredCount) }] : []}
		visible={tooltipVisible && hoveredName.length > 0}
	/>
</div>
