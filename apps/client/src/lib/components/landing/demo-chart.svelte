<script lang="ts">
	import { fmt, hourLabel } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

	let {
		count,
		hour,
		tip,
		hits,
		current = $bindable()
	}: { count: number; hour: number; tip: boolean; hits: number; current?: HTMLElement } = $props();

	const HOURS = [9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20];
	const SHAPE = [34, 48, 62, 71, 58, 66, 80, 74, 61, 52, 69];
</script>

<div class="vd-chart">
	<div class="vd-bars">
		{#each SHAPE.slice(SHAPE.length - count + 1) as height, i (i)}
			<span class="vd-bar" style:height="{height}%"></span>
		{/each}
		<span bind:this={current} class="vd-bar vd-now" style:height="{Math.min(hour, 100)}%">
			{#if tip}
				<span class="vd-tip">
					<b>{hourLabel(HOURS[HOURS.length - 1])}</b>
					<span class="vd-tip-row">{t('stat.visitors')}<b class="tabular-nums">{fmt(hits)}</b></span>
				</span>
			{/if}
		</span>
	</div>
	<div class="vd-axis">
		{#each HOURS.slice(HOURS.length - count) as h, i (h)}
			<span class:vd-hidden={(count - 1 - i) % 2 === 1}>{hourLabel(h)}</span>
		{/each}
	</div>
</div>

<style>
	.vd-chart {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 8px;
		min-height: 0;
		padding: 18px 18px 12px;
		border-radius: 12px;
		background: var(--mes-w-side);
	}
	:global(.vd-compact) .vd-chart {
		padding: 14px 12px 10px;
	}
	.vd-bars {
		display: flex;
		flex: 1;
		align-items: flex-end;
		gap: 10px;
		min-height: 0;
		padding-top: 52px;
	}
	:global(.vd-compact) .vd-bars {
		gap: 8px;
	}
	.vd-bar {
		position: relative;
		flex: 1;
		border-radius: 6px 6px 3px 3px;
		background: color-mix(in oklch, var(--color-fc-fg) 14%, transparent);
	}
	.vd-now {
		background: var(--color-fc-accent);
		transition: height 0.8s cubic-bezier(0.25, 0.8, 0.25, 1);
	}
	@media (prefers-reduced-motion: reduce) {
		.vd-now {
			transition: none;
		}
	}
	.vd-tip {
		position: absolute;
		right: 0;
		bottom: calc(100% + 8px);
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 8px 12px;
		border-radius: 10px;
		background: var(--mes-w-body);
		box-shadow: 0 8px 24px -8px rgb(0 0 0 / 0.25);
		font-size: 12px;
		white-space: nowrap;
	}
	.vd-tip b {
		font-weight: 700;
	}
	.vd-tip-row {
		display: flex;
		gap: 8px;
		color: var(--color-fc-fg-muted);
	}
	.vd-tip-row b {
		color: var(--color-fc-fg);
	}
	.vd-axis {
		display: flex;
		gap: 10px;
		color: var(--color-fc-fg-muted);
		font-size: 11px;
	}
	:global(.vd-compact) .vd-axis {
		gap: 8px;
	}
	.vd-axis span {
		flex: 1;
		min-width: 0;
		text-align: center;
		white-space: nowrap;
	}
	.vd-hidden {
		visibility: hidden;
	}
</style>
