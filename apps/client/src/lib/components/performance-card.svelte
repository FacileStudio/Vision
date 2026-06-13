<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import { fmt } from '$lib/utils/analytics';

	let {
		performance
	}: {
		performance: {
			avg_dns: number;
			avg_tcp: number;
			avg_ttfb: number;
			avg_dom_load: number;
			avg_page_load: number;
			sample_count: number;
		};
	} = $props();

	const metrics = $derived([
		{ label: 'DNS', value: performance.avg_dns },
		{ label: 'TCP', value: performance.avg_tcp },
		{ label: 'TTFB', value: performance.avg_ttfb },
		{ label: 'DOM Load', value: performance.avg_dom_load },
		{ label: 'Page Load', value: performance.avg_page_load }
	]);
</script>

<Card.Root class="mb-8">
	<Card.Header>
		<div class="flex items-center justify-between">
			<Card.Title>Page Load Performance</Card.Title>
			<span class="text-xs text-muted-foreground">{fmt(performance.sample_count)} samples</span>
		</div>
	</Card.Header>
	<Card.Content>
		<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-5">
			{#each metrics as m (m.label)}
				<div>
					<p class="text-xs text-muted-foreground">{m.label}</p>
					<p class="text-lg font-semibold">{Math.round(m.value)}ms</p>
				</div>
			{/each}
		</div>
	</Card.Content>
</Card.Root>
