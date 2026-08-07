<script lang="ts">
	import { Drawer, Input } from '@facile/muse';
	import { fmt } from '$lib/utils/analytics';

	let {
		open = $bindable(false),
		title = '',
		items = [],
		onFilter
	}: {
		open?: boolean;
		title: string;
		items: { label: string; count: number }[];
		onFilter?: (value: string) => void;
	} = $props();

	let search = $state('');

	$effect(() => {
		if (!open) search = '';
	});

	let filtered = $derived(
		search ? items.filter((i) => i.label.toLowerCase().includes(search.toLowerCase())) : items
	);

	let maxCount = $derived(Math.max(...filtered.map((d) => d.count), 1));
	let total = $derived(items.reduce((sum, i) => sum + i.count, 0));
</script>

<Drawer bind:open {title} description="{fmt(total)} total · {items.length} entries" showClose>
	<div class="flex flex-col gap-3">
		<Input bind:value={search} placeholder="Search…" aria-label="Search {title}" />

		<div class="flex flex-col gap-0.5">
			<!-- Positional, same reasoning as bar-list: a ranked list keyed on API-supplied
			     text is one duplicate away from a fatal each_key_duplicate. -->
			{#each filtered as item, i (i)}
				{@const width = `${(item.count / maxCount) * 100}%`}
				{#if onFilter}
					<button
						type="button"
						class="relative w-full overflow-hidden rounded-fc-xs text-left transition-colors hover:bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
						onclick={() => {
							onFilter(item.label);
							open = false;
						}}
					>
						<span class="absolute inset-y-0 left-0 rounded-fc-xs bg-fc-surface" style:width aria-hidden="true"></span>
						<span class="relative flex justify-between gap-3 px-3 py-2 text-fc-sm">
							<span class="truncate text-fc-fg">{item.label}</span>
							<span class="shrink-0 tabular-nums text-fc-fg-muted">{fmt(item.count)}</span>
						</span>
					</button>
				{:else}
					<div class="relative overflow-hidden rounded-fc-xs">
						<span class="absolute inset-y-0 left-0 rounded-fc-xs bg-fc-surface" style:width aria-hidden="true"></span>
						<span class="relative flex justify-between gap-3 px-3 py-2 text-fc-sm">
							<span class="truncate text-fc-fg">{item.label}</span>
							<span class="shrink-0 tabular-nums text-fc-fg-muted">{fmt(item.count)}</span>
						</span>
					</div>
				{/if}
			{/each}

			{#if filtered.length === 0}
				<p class="py-8 text-center text-fc-sm text-fc-fg-muted">Nothing matches “{search}”.</p>
			{/if}
		</div>
	</div>
</Drawer>
