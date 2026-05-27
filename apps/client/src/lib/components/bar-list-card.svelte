<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import { fmt } from '$lib/utils/analytics';
	import type { Snippet } from 'svelte';

	let {
		title,
		items,
		showSeeAll = false,
		onSeeAll,
		onItemClick,
		label: labelSnippet,
		formatCount = false,
		class: className = ''
	}: {
		title: string;
		items: { label: string; count: number }[];
		showSeeAll?: boolean;
		onSeeAll?: () => void;
		onItemClick?: (label: string) => void;
		label?: Snippet<[{ label: string; count: number }]>;
		formatCount?: boolean;
		class?: string;
	} = $props();

	let maxCount = $derived(Math.max(...items.map((d) => d.count), 1));
</script>

<Card.Root class={className}>
	<Card.Header>
		<div class="flex items-center justify-between">
			<Card.Title>{title}</Card.Title>
			{#if showSeeAll && onSeeAll}
				<button
					onclick={onSeeAll}
					class="text-xs text-muted-foreground hover:text-foreground transition-colors"
				>
					See all →
				</button>
			{/if}
		</div>
	</Card.Header>
	<Card.Content>
		<div class="space-y-1">
			{#each items as item (item.label)}
				<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
				<div
					role={onItemClick ? 'button' : undefined}
					tabindex={onItemClick ? 0 : undefined}
					class="relative rounded transition-colors {onItemClick
						? 'cursor-pointer hover:bg-muted/30'
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
					<div
						class="absolute inset-y-0 left-0 rounded bg-muted/50"
						style="width: {(item.count / maxCount) * 100}%"
					></div>
					<div class="relative flex justify-between px-3 py-1.5 text-sm">
						{#if labelSnippet}
							{@render labelSnippet(item)}
						{:else}
							<span class="truncate mr-2">{item.label}</span>
						{/if}
						<span class="text-muted-foreground tabular-nums shrink-0">
							{formatCount ? fmt(item.count) : item.count}
						</span>
					</div>
				</div>
			{/each}
		</div>
	</Card.Content>
</Card.Root>
