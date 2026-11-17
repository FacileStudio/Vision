<script lang="ts">
	import { Card, StatCard } from '@facile/muse';
	import { api, type BotsResponse } from '$lib';
	import BarList from '$lib/components/bar-list.svelte';
	import { fmt } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';
	import type { AppKey } from '$lib/i18n/messages';

	let { siteId, from, to }: { siteId: number; from: string; to: string } = $props();

	const CATEGORIES = ['search', 'ai', 'social', 'monitor', 'tool', 'headless', 'automation', 'spoofed'];

	let data = $state<BotsResponse | null>(null);
	let latest = 0;

	function categoryLabel(name: string): string {
		return CATEGORIES.includes(name) ? t(`bots.cat.${name}` as AppKey) : name;
	}

	const categories = $derived(
		(data?.categories ?? []).map((c) => ({ label: categoryLabel(c.name), count: c.hits }))
	);
	const bots = $derived(
		(data?.bots ?? []).map((b) => ({ label: b.name, category: categoryLabel(b.category), count: b.hits }))
	);
	const paths = $derived((data?.paths ?? []).map((p) => ({ label: p.path, count: p.hits })));

	$effect(() => {
		const request = ++latest;
		api.analytics
			.bots(siteId, from, to)
			.then((res) => {
				if (request === latest) data = res;
			})
			.catch(() => {
				if (request === latest) data = null;
			});
	});
</script>

{#if data}
	<section class="flex flex-col gap-4">
		<div class="flex flex-col gap-1">
			<h2 class="text-fc-lg font-semibold text-fc-fg">{t('bots.title')}</h2>
			<p class="text-fc-sm text-fc-fg-muted">{t('bots.text')}</p>
		</div>
		{#if data.total > 0}
			<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
				<StatCard label={t('bots.total')} value={fmt(data.total)} />
				<BarList title={t('bots.byCategory')} items={categories} class="lg:col-span-2" />
				<BarList title={t('bots.top')} items={bots} class="lg:col-span-2">
					{#snippet label(item)}
						<span class="flex min-w-0 items-baseline gap-2">
							<span class="truncate text-fc-fg">{item.label}</span>
							<span class="shrink-0 text-fc-xs text-fc-fg-muted">
								{(item as (typeof bots)[number]).category}
							</span>
						</span>
					{/snippet}
				</BarList>
				<BarList title={t('bots.paths')} items={paths}>
					{#snippet label(item)}
						<span class="truncate font-fc-mono text-fc-xs text-fc-fg">{item.label}</span>
					{/snippet}
				</BarList>
			</div>
		{:else}
			<Card>
				<p class="text-fc-sm text-fc-fg-muted">{t('bots.empty')}</p>
			</Card>
		{/if}
	</section>
{/if}
