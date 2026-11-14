<script lang="ts">
	import { Bar, Cursor, Roll, Stage, Window, cursorAt, type Act } from '@facile/mise-en-scene';
	import { Icon, icons } from '@facile/muse';
	import { fmt } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';
	import DemoChart from './demo-chart.svelte';
	import DemoLists from './demo-lists.svelte';
	import DemoSidebar from './demo-sidebar.svelte';

	const START = { visitors: 1284, pageviews: 3906, live: 7, hour: 18 };
	const STEPS = [
		{ visitors: 3, pageviews: 7, live: 9 },
		{ visitors: 5, pageviews: 11, live: 12 },
		{ visitors: 2, pageviews: 6, live: 11 },
		{ visitors: 6, pageviews: 14, live: 15 },
		{ visitors: 4, pageviews: 9, live: 17 },
		{ visitors: 7, pageviews: 16, live: 21 }
	];

	let visitors = $state(START.visitors);
	let pageviews = $state(START.pageviews);
	let live = $state(START.live);
	let hour = $state(START.hour);
	let tip = $state(false);
	let shown = $state(0);
	let cursor = $state(cursorAt());
	let current = $state<HTMLElement>();

	function reset() {
		visitors = START.visitors;
		pageviews = START.pageviews;
		live = START.live;
		hour = START.hour;
		tip = false;
		shown = 0;
		cursor.visible = false;
	}

	function poster() {
		visitors = START.visitors + 27;
		pageviews = START.pageviews + 63;
		live = 21;
		hour = 88;
		tip = true;
		shown = STEPS.length;
	}

	async function script(act: Act) {
		await act.wait(900);
		for (const step of STEPS) {
			visitors += step.visitors;
			pageviews += step.pageviews;
			live = step.live;
			hour += 12;
			shown += 1;
			await act.wait(900);
		}
		cursor.visible = true;
		await act.move(cursor, act.point(current, 0.55, 0.35));
		tip = true;
		await act.wait(2400);
		reset();
		await act.wait(700);
	}
</script>

{#snippet stat(label: string, value: number, pulse = false)}
	<div class="vd-stat">
		<span class="vd-label">
			{#if pulse}<span class="vd-dot"></span>{/if}
			{label}
		</span>
		<Roll value={fmt(value)} class="vd-value tabular-nums" />
	</div>
{/snippet}

{#snippet header()}
	<div class="vd-head">
		<span class="vd-favicon"><Icon icon={icons.globe} size={16} /></span>
		<div class="vd-site">
			<b>atelier-lune.fr</b>
			<Bar w={92} h={7} />
		</div>
	</div>
{/snippet}

{#snippet sidebar()}
	<DemoSidebar />
{/snippet}

<div class="vd-host">
	<Stage
		width={976}
		height={640}
		compact={{ below: 640, width: 358, height: 600 }}
		{script}
		{reset}
		{poster}
		label={t('landing.demo')}
	>
		{#snippet children({ compact })}
			{#if compact}
				<Window style="top:8px;left:8px;width:342px;height:556px">
					<div class="vd-body vd-compact">
						{@render header()}
						<div class="vd-stats vd-two">
							{@render stat(t('stat.visitors'), visitors)}
							{@render stat(t('stat.activeNow'), live, true)}
						</div>
						<DemoChart count={7} {hour} {tip} hits={visitors - START.visitors + 96} bind:current />
						<DemoLists {shown} compact />
					</div>
				</Window>
			{:else}
				<Window style="top:8px;left:12px;width:952px;height:584px" {sidebar}>
					<div class="vd-body">
						{@render header()}
						<div class="vd-stats">
							{@render stat(t('stat.visitors'), visitors)}
							{@render stat(t('stat.pageviews'), pageviews)}
							{@render stat(t('stat.activeNow'), live, true)}
						</div>
						<DemoChart count={12} {hour} {tip} hits={visitors - START.visitors + 96} bind:current />
						<DemoLists {shown} />
					</div>
				</Window>
			{/if}
			<Cursor {cursor} />
		{/snippet}
	</Stage>
</div>

<style>
	.vd-host {
		--mes-shadow-hero: 0 16px 32px -14px rgb(0 0 0 / 0.2), 0 4px 12px -4px rgb(0 0 0 / 0.08);
	}
	.vd-body {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 16px;
		min-height: 0;
		padding: 22px 24px;
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
	}
	.vd-compact {
		gap: 12px;
		padding: 14px;
	}
	.vd-head {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.vd-favicon {
		display: inline-flex;
		flex: none;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 10px;
		background: var(--mes-w-side);
	}
	.vd-site {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
	}
	.vd-site b {
		font-family: var(--font-fc-title);
		font-size: 17px;
		font-weight: 700;
		white-space: nowrap;
	}
	.vd-stats {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 12px;
	}
	.vd-two {
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 8px;
	}
	.vd-stat {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 14px 16px;
		border-radius: 12px;
		background: var(--mes-w-side);
	}
	.vd-compact .vd-stat {
		padding: 12px;
	}
	.vd-label {
		display: flex;
		align-items: center;
		gap: 6px;
		color: var(--color-fc-fg-muted);
		font-size: 12px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.vd-dot {
		flex: none;
		width: 7px;
		height: 7px;
		border-radius: 999px;
		background: var(--color-fc-success);
	}
	.vd-stat :global(.vd-value) {
		font-family: var(--font-fc-title);
		font-size: 26px;
		font-weight: 700;
		line-height: 1.1;
	}
	.vd-compact .vd-stat :global(.vd-value) {
		font-size: 22px;
	}
</style>
