<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { Site, AnalyticsOverview } from '$lib';

	const apiBase = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:4000';
	let site = $state<Site | null>(null);
	let overview = $state<AnalyticsOverview | null>(null);

	onMount(async () => {
		const id = Number(page.params.id);
		site = await api.sites.get(id);
		overview = await api.analytics.overview(id);
	});

	async function rotateKey() {
		if (!site) return;
		site = await api.sites.rotateKey(site.id);
	}
</script>

{#if site}
	<div class="mb-8">
		<h1 class="text-2xl font-bold">{site.name}</h1>
		<p class="text-muted-foreground">{site.domain}</p>
	</div>

	<div class="mb-8 rounded-lg border p-4">
		<h2 class="font-semibold mb-2">Tracking Script</h2>
		<p class="text-sm text-muted-foreground mb-2">Add this to your website's &lt;head&gt;:</p>
		<pre class="rounded bg-muted p-3 text-xs overflow-x-auto">&lt;script defer src="{apiBase}/t/{site.api_key}.js"&gt;&lt;/script&gt;</pre>
		<button onclick={rotateKey} class="mt-3 text-sm text-destructive hover:underline">Rotate API Key</button>
	</div>

	{#if overview}
		<div class="grid gap-4 md:grid-cols-2 mb-8">
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Total Pageviews</p>
				<p class="text-3xl font-bold">{overview.total_pageviews.toLocaleString()}</p>
			</div>
			<div class="rounded-lg border p-4">
				<p class="text-sm text-muted-foreground">Unique Visitors</p>
				<p class="text-3xl font-bold">{overview.unique_visitors.toLocaleString()}</p>
			</div>
		</div>

		{#if overview.top_pages.length > 0}
			<div class="mb-8">
				<h2 class="font-semibold mb-3">Top Pages</h2>
				<div class="space-y-1">
					{#each overview.top_pages as entry}
						<div class="flex justify-between rounded px-3 py-2 text-sm hover:bg-muted">
							<span>{entry.path}</span>
							<span class="text-muted-foreground">{entry.count}</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		{#if overview.top_referrers.length > 0}
			<div class="mb-8">
				<h2 class="font-semibold mb-3">Top Referrers</h2>
				<div class="space-y-1">
					{#each overview.top_referrers as entry}
						<div class="flex justify-between rounded px-3 py-2 text-sm hover:bg-muted">
							<span>{entry.referrer}</span>
							<span class="text-muted-foreground">{entry.count}</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	{/if}
{/if}
