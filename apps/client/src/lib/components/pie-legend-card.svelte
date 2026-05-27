<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { PieChart } from 'layerchart';

	type PieItem = { key: string; label: string; value: number; color: string };

	let {
		title,
		data,
		configKey,
		configColor = 'var(--chart-1)',
		onItemClick,
		class: className = ''
	}: {
		title: string;
		data: PieItem[];
		configKey: string;
		configColor?: string;
		onItemClick?: (label: string) => void;
		class?: string;
	} = $props();

	let config = $derived(
		{ [configKey]: { label: title, color: configColor } } satisfies Chart.ChartConfig
	);
</script>

<Card.Root class={className}>
	<Card.Header>
		<Card.Title>{title}</Card.Title>
	</Card.Header>
	<Card.Content>
		<div class="min-h-[180px]">
			<Chart.Container {config} class="h-[180px] w-full">
				<PieChart {data} key="key" label="label" value="value" innerRadius={0.5} />
			</Chart.Container>
		</div>
		<div class="mt-3 space-y-1">
			{#each data as item (item.key)}
				<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
				<div
					role={onItemClick ? 'button' : undefined}
					tabindex={onItemClick ? 0 : undefined}
					class="flex items-center gap-2 text-xs {onItemClick
						? 'cursor-pointer hover:bg-muted/30 rounded px-1 py-0.5 transition-colors'
						: ''}"
					onclick={onItemClick ? () => onItemClick(item.label) : undefined}
					onkeydown={onItemClick
						? (e) => {
								if (e.key === 'Enter' || e.key === ' ') {
									e.preventDefault();
									onItemClick(item.label);
								}
							}
						: undefined}
				>
					<span class="h-2.5 w-2.5 rounded-full shrink-0" style="background: {item.color}"></span>
					<span class="truncate">{item.label}</span>
					<span class="text-muted-foreground tabular-nums ml-auto">{item.value}</span>
				</div>
			{/each}
		</div>
	</Card.Content>
</Card.Root>
