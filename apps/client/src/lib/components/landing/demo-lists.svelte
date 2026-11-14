<script lang="ts">
	import { fmt } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

	type Row = { label: string; count: number };

	let { shown, compact = false }: { shown: number; compact?: boolean } = $props();

	const sources: Row[] = $derived([
		{ label: 'Google', count: 412 },
		{ label: 'Instagram', count: 268 },
		{ label: t('source.direct'), count: 154 }
	]);
	const pages: Row[] = [
		{ label: '/', count: 538 },
		{ label: '/tarifs', count: 311 },
		{ label: '/contact', count: 126 }
	];
	const lists = $derived(
		compact
			? [{ title: t('block.referrers'), rows: sources }]
			: [
					{ title: t('block.referrers'), rows: sources },
					{ title: t('block.pages'), rows: pages }
				]
	);
</script>

<div class="vl-lists" class:vl-one={compact}>
	{#each lists as list, column (list.title)}
		<div class="vl-card">
			<p class="vl-title">{list.title}</p>
			{#each list.rows as row, index (index)}
				{@const max = list.rows[0].count}
				<div class="vl-row" class:vl-on={index * lists.length + column < shown}>
					<span class="vl-fill" style:transform="scaleX({row.count / max})"></span>
					<span class="vl-label">{row.label}</span>
					<span class="vl-count tabular-nums">{fmt(row.count)}</span>
				</div>
			{/each}
		</div>
	{/each}
</div>

<style>
	.vl-lists {
		display: grid;
		flex: none;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 12px;
	}
	.vl-one {
		grid-template-columns: minmax(0, 1fr);
	}
	.vl-card {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 12px 14px;
		border-radius: 12px;
		background: var(--mes-w-side);
	}
	.vl-title {
		margin-bottom: 4px;
		color: var(--color-fc-fg);
		font-size: 12px;
		font-weight: 600;
	}
	.vl-row {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		height: 26px;
		padding: 0 10px;
		font-size: 12px;
		opacity: 0;
		transform: translateY(6px);
		transition:
			opacity 0.4s ease,
			transform 0.4s ease;
	}
	.vl-on {
		opacity: 1;
		transform: none;
	}
	.vl-fill {
		position: absolute;
		inset: 0;
		border-radius: 6px;
		background: color-mix(in oklch, var(--color-fc-fg) 8%, transparent);
		transform-origin: left;
	}
	.vl-label,
	.vl-count {
		position: relative;
		white-space: nowrap;
	}
	.vl-label {
		overflow: hidden;
		color: var(--color-fc-fg);
		text-overflow: ellipsis;
	}
	.vl-count {
		color: var(--color-fc-fg-muted);
	}
	@media (prefers-reduced-motion: reduce) {
		.vl-row {
			transition: none;
		}
	}
</style>
