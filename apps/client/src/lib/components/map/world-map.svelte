<script lang="ts">
	import { Chart, GeoPath, Svg } from 'layerchart';
	import { geoNaturalEarth1 } from 'd3-geo';
	import { feature } from 'topojson-client';
	import type { Topology } from 'topojson-specification';
	import type { FeatureCollection, Geometry } from 'geojson';
	import { getNumericCode, getAlpha2FromNumeric, getCountryName } from './country-codes';

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

	let hoveredName = $derived(() => {
		if (!hoveredId) return '';
		const a2 = getAlpha2FromNumeric(hoveredId);
		return a2 ? getCountryName(a2) : '';
	});

	let hoveredCount = $derived(() => {
		if (!hoveredId) return 0;
		return countsByNumeric().get(hoveredId) ?? 0;
	});

	function getFill(featureId: string): string {
		const isHovered = hoveredId === featureId;
		const count = countsByNumeric().get(featureId);
		if (!count) return isHovered ? 'oklch(0.88 0 0)' : 'oklch(0.95 0 0)';
		const intensity = 0.15 + 0.85 * (count / maxCount);
		const lightness = 0.85 - intensity * 0.65;
		return `oklch(${isHovered ? Math.max(0.1, lightness - 0.1) : lightness} 0 0)`;
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
		const mx = e.clientX - rect.left;
		const my = e.clientY - rect.top;
		const factor = e.deltaY > 0 ? 0.9 : 1.1;
		zoomAt(factor, mx, my);
	}

	function handleMouseDown(e: MouseEvent) {
		dragging = true;
		dragStartX = e.clientX;
		dragStartY = e.clientY;
		panStartX = panX;
		panStartY = panY;
	}

	function handleMouseMove(e: MouseEvent) {
		if (dragging) {
			panX = panStartX + (e.clientX - dragStartX) / scale;
			panY = panStartY + (e.clientY - dragStartY) / scale;
		}
	}

	function handleMouseUp() {
		dragging = false;
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
</script>

<svelte:window onmouseup={handleMouseUp} onmousemove={handleMouseMove} />

<div
	bind:this={containerEl}
	class="relative h-[300px] w-full overflow-hidden rounded-lg sm:h-[400px]"
	style="cursor: {dragging ? 'grabbing' : 'grab'}"
>
	<div
		class="h-full w-full"
		role="img"
		onwheel={handleWheel}
		onmousedown={handleMouseDown}
	>
		<Chart geo={{ projection: geoNaturalEarth1, fitGeojson: geojson }}>
			<Svg>
				<g transform="translate({panX * scale}, {panY * scale}) scale({scale})">
					{#each geojson.features as f}
						<GeoPath
							geojson={f}
							fill={getFill(String(f.id))}
							stroke="oklch(0.85 0 0)"
							stroke-width={0.5 / scale}
							style="transition: fill 0.15s ease"
							onmouseenter={(e: MouseEvent) => {
								hoveredId = String(f.id);
								tooltipX = e.clientX;
								tooltipY = e.clientY;
								tooltipVisible = true;
							}}
							onmousemove={(e: MouseEvent) => {
								tooltipX = e.clientX;
								tooltipY = e.clientY;
							}}
							onmouseleave={() => {
								hoveredId = null;
								tooltipVisible = false;
							}}
						/>
					{/each}
				</g>
			</Svg>
		</Chart>
	</div>

	<div class="absolute right-2 top-2 flex flex-col gap-1">
		<button
			onclick={zoomIn}
			class="flex h-7 w-7 items-center justify-center rounded border bg-background text-sm hover:bg-muted"
		>+</button>
		<button
			onclick={zoomOut}
			class="flex h-7 w-7 items-center justify-center rounded border bg-background text-sm hover:bg-muted"
		>-</button>
		<button
			onclick={resetView}
			class="flex h-7 w-7 items-center justify-center rounded border bg-background text-xs hover:bg-muted"
		>R</button>
	</div>

	{#if tooltipVisible && hoveredName()}
		<div
			class="pointer-events-none fixed z-50 rounded border bg-background px-2.5 py-1.5 text-sm shadow-sm"
			style="left: {tooltipX + 12}px; top: {tooltipY - 10}px"
		>
			<span class="font-medium">{hoveredName()}</span>
			{#if hoveredCount() > 0}
				<span class="ml-2 text-muted-foreground">{hoveredCount()}</span>
			{/if}
		</div>
	{/if}
</div>
