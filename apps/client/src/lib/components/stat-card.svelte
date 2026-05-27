<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';

	let {
		label,
		value,
		trend,
		pulse = false
	}: {
		label: string;
		value: string;
		trend?: { text: string; color: string };
		pulse?: boolean;
	} = $props();
</script>

<Card.Root size="sm">
	<Card.Header>
		<Card.Description>{label}</Card.Description>
		<Card.Title class="text-2xl font-semibold tabular-nums {pulse ? 'flex items-center gap-2' : ''}">
			{#if pulse}
				<span class="relative flex h-2 w-2">
					<span
						class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75"
					></span>
					<span class="relative inline-flex h-2 w-2 rounded-full bg-green-500"></span>
				</span>
			{/if}
			{value}
		</Card.Title>
	</Card.Header>
	{#if trend && trend.text !== '—'}
		<Card.Content>
			<span class="text-xs {trend.color}">{trend.text}</span>
		</Card.Content>
	{/if}
</Card.Root>
