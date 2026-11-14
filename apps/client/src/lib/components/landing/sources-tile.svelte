<script lang="ts">
	import { Stage, Tile, Window, type Act } from '@facile/mise-en-scene';
	import { fmt } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

	const sources = $derived([
		{ label: 'Google', count: 412 },
		{ label: 'Instagram', count: 268 },
		{ label: t('source.direct'), count: 154 },
		{ label: 'LinkedIn', count: 97 }
	]);
	const pages = [
		{ label: '/', count: 538 },
		{ label: '/tarifs', count: 311 },
		{ label: '/blog/atelier', count: 184 },
		{ label: '/contact', count: 126 }
	];

	let tab = $state<'sources' | 'pages'>('sources');
	let filled = $state(false);
	const panels = $derived([
		{ id: 'sources' as const, rows: sources },
		{ id: 'pages' as const, rows: pages }
	]);

	function reset() {
		tab = 'sources';
		filled = false;
	}

	function poster() {
		tab = 'sources';
		filled = true;
	}

	async function script(act: Act) {
		await act.wait(400);
		filled = true;
		await act.wait(2600);
		tab = 'pages';
		await act.wait(3000);
		tab = 'sources';
		await act.wait(800);
	}
</script>

<Tile title={t('landing.card.sources.title')} body={t('landing.card.sources.text')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<Window style="top:12px;left:20px;width:280px;height:216px">
				<div class="st-body">
					<div class="st-tabs">
						<span class="st-pill" class:st-right={tab === 'pages'}></span>
						<span class="st-tab" class:st-active={tab === 'sources'}>{t('landing.demo.sources')}</span>
						<span class="st-tab" class:st-active={tab === 'pages'}>{t('landing.demo.pages')}</span>
					</div>
					<div class="st-panels">
						{#each panels as panel (panel.id)}
							<div class="st-panel" class:st-shown={tab === panel.id}>
								{#each panel.rows as row (row.label)}
									<div class="st-row">
										<span
											class="st-fill"
											style:transform="scaleX({filled ? row.count / panel.rows[0].count : 0})"
										></span>
										<span class="st-label">{row.label}</span>
										<span class="st-count tabular-nums">{fmt(row.count)}</span>
									</div>
								{/each}
							</div>
						{/each}
					</div>
				</div>
			</Window>
		{/snippet}
	</Stage>
</Tile>

<style>
	.st-body {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 12px 14px;
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
		font-size: 12px;
	}
	.st-tabs {
		position: relative;
		display: grid;
		grid-template-columns: 1fr 1fr;
		padding: 3px;
		border-radius: 999px;
		background: var(--mes-w-side);
	}
	.st-pill {
		position: absolute;
		top: 3px;
		bottom: 3px;
		left: 3px;
		width: calc(50% - 3px);
		border-radius: 999px;
		background: var(--mes-w-body);
		box-shadow: 0 1px 3px rgb(0 0 0 / 0.12);
		transition: transform 0.4s cubic-bezier(0.25, 0.8, 0.25, 1);
	}
	.st-right {
		transform: translateX(100%);
	}
	.st-tab {
		position: relative;
		padding: 4px 0;
		color: var(--color-fc-fg-muted);
		font-weight: 600;
		text-align: center;
		transition: color 0.3s ease;
	}
	.st-active {
		color: var(--color-fc-fg);
	}
	.st-panels {
		display: grid;
	}
	.st-panel {
		display: flex;
		flex-direction: column;
		grid-area: 1 / 1;
		gap: 4px;
		opacity: 0;
		transition: opacity 0.35s ease;
	}
	.st-shown {
		opacity: 1;
	}
	.st-row {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		height: 28px;
		padding: 0 10px;
	}
	.st-fill {
		position: absolute;
		inset: 0;
		border-radius: 6px;
		background: color-mix(in oklch, var(--color-fc-accent) 18%, transparent);
		transform-origin: left;
		transition: transform 0.9s cubic-bezier(0.25, 0.8, 0.25, 1);
	}
	.st-label,
	.st-count {
		position: relative;
		white-space: nowrap;
	}
	.st-count {
		color: var(--color-fc-fg-muted);
	}
	@media (prefers-reduced-motion: reduce) {
		.st-pill,
		.st-tab,
		.st-panel,
		.st-fill {
			transition: none;
		}
	}
</style>
