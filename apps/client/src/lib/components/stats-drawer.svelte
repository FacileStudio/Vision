<script lang="ts">
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import Icon from '@iconify/svelte';

	let {
		open = false,
		onclose,
		title = '',
		items = [],
		onFilter
	}: {
		open: boolean;
		onclose: () => void;
		title: string;
		items: { label: string; count: number }[];
		onFilter?: (value: string) => void;
	} = $props();

	let search = $state('');

	$effect(() => {
		if (!open) search = '';
	});

	let filtered = $derived(
		search
			? items.filter((i) => i.label.toLowerCase().includes(search.toLowerCase()))
			: items
	);

	let maxCount = $derived(Math.max(...filtered.map((d) => d.count), 1));
	let total = $derived(items.reduce((sum, i) => sum + i.count, 0));

	function fmt(n: number): string {
		return n.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
	}
</script>

<Sheet.Root {open} onOpenChange={(v) => { if (!v) onclose(); }}>
	<Sheet.Content side="right" class="sm:max-w-lg">
		<Sheet.Header>
			<Sheet.Title>{title}</Sheet.Title>
			<Sheet.Description>{fmt(total)} total · {items.length} entries</Sheet.Description>
		</Sheet.Header>
		<div class="px-4">
			<div class="relative">
				<Icon icon="solar:magnifer-linear" class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
				<input
					type="text"
					bind:value={search}
					placeholder="Search…"
					class="w-full rounded-md border bg-background pl-9 pr-3 py-2 text-sm placeholder:text-muted-foreground"
				/>
			</div>
		</div>
		<div class="flex-1 overflow-y-auto px-4">
			<div class="space-y-1">
				{#each filtered as item}
					{#if onFilter}
						<button
							class="relative w-full text-left cursor-pointer hover:bg-muted/30 rounded transition-colors"
							onclick={() => { onFilter(item.label); onclose(); }}
						>
							<div
								class="absolute inset-y-0 left-0 rounded bg-muted/50"
								style="width: {(item.count / maxCount) * 100}%"
							></div>
							<div class="relative flex justify-between px-3 py-1.5 text-sm">
								<span class="truncate mr-2">{item.label}</span>
								<span class="text-muted-foreground tabular-nums shrink-0">{fmt(item.count)}</span>
							</div>
						</button>
					{:else}
						<div class="relative">
							<div
								class="absolute inset-y-0 left-0 rounded bg-muted/50"
								style="width: {(item.count / maxCount) * 100}%"
							></div>
							<div class="relative flex justify-between px-3 py-1.5 text-sm">
								<span class="truncate mr-2">{item.label}</span>
								<span class="text-muted-foreground tabular-nums shrink-0">{fmt(item.count)}</span>
							</div>
						</div>
					{/if}
				{/each}
				{#if filtered.length === 0}
					<p class="text-sm text-muted-foreground text-center py-8">No results</p>
				{/if}
			</div>
		</div>
	</Sheet.Content>
</Sheet.Root>
