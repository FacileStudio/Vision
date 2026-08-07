<script lang="ts">
	import { Button, Card, icons } from '@facile/muse';
	import { fmt } from '$lib/utils/analytics';
	import type { Snippet } from 'svelte';

	let {
		title,
		items,
		showSeeAll = false,
		onSeeAll,
		onItemClick,
		label: labelSnippet,
		empty = 'Nothing recorded yet.',
		class: className = ''
	}: {
		title: string;
		items: { label: string; count: number }[];
		showSeeAll?: boolean;
		onSeeAll?: () => void;
		onItemClick?: (label: string) => void;
		label?: Snippet<[{ label: string; count: number }]>;
		empty?: string;
		class?: string;
	} = $props();

	/*
	 * The bar is scaled against the largest row, not against the total: these lists are
	 * top-N slices of a long tail, so a share-of-total bar would render every row as a
	 * sliver and say nothing about the ranking, which is the whole point of the list.
	 */
	let maxCount = $derived(Math.max(...items.map((d) => d.count), 1));
</script>

<Card class="flex flex-col gap-4 {className}">
	<div class="flex items-center justify-between gap-3">
		<p class="text-fc-sm font-medium text-fc-fg">{title}</p>
		{#if showSeeAll && onSeeAll}
			<Button variant="ghost" size="sm" iconRight={icons.arrow} onclick={onSeeAll}>See all</Button>
		{/if}
	</div>

	{#if items.length === 0}
		<p class="text-fc-sm text-fc-fg-muted">{empty}</p>
	{:else}
		<div class="flex flex-col gap-0.5">
			{#each items as item (item.label)}
				{@const width = `${(item.count / maxCount) * 100}%`}
				{#if onItemClick}
					<button
						type="button"
						class="relative w-full overflow-hidden rounded-fc-xs text-left transition-colors hover:bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
						onclick={() => onItemClick(item.label)}
					>
						<span
							class="absolute inset-y-0 left-0 rounded-fc-xs bg-fc-surface"
							style:width
							aria-hidden="true"
						></span>
						<span class="relative flex justify-between gap-3 px-3 py-1.5 text-fc-sm">
							{#if labelSnippet}
								{@render labelSnippet(item)}
							{:else}
								<span class="truncate text-fc-fg">{item.label}</span>
							{/if}
							<span class="shrink-0 tabular-nums text-fc-fg-muted">{fmt(item.count)}</span>
						</span>
					</button>
				{:else}
					<div class="relative overflow-hidden rounded-fc-xs">
						<span
							class="absolute inset-y-0 left-0 rounded-fc-xs bg-fc-surface"
							style:width
							aria-hidden="true"
						></span>
						<span class="relative flex justify-between gap-3 px-3 py-1.5 text-fc-sm">
							{#if labelSnippet}
								{@render labelSnippet(item)}
							{:else}
								<span class="truncate text-fc-fg">{item.label}</span>
							{/if}
							<span class="shrink-0 tabular-nums text-fc-fg-muted">{fmt(item.count)}</span>
						</span>
					</div>
				{/if}
			{/each}
		</div>
	{/if}
</Card>
