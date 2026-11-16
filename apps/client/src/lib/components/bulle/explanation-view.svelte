<script lang="ts">
	import { Badge } from '@facile/muse';
	import type { Explanation } from '$lib/bulle';
	import { t } from '$lib/i18n/index.svelte';

	let { explanation }: { explanation: Explanation } = $props();

	const tones = { low: 'neutral', medium: 'info', high: 'success' } as const;
</script>

<!--
	Every string below comes from a model that read visitor-written text. It is printed as
	text with `{...}` and never as markup, and no value becomes a link.
-->
<div class="flex min-w-0 flex-col gap-4 rounded-fc-md bg-fc-surface p-4">
	<p class="m-0 text-fc-sm wrap-anywhere whitespace-pre-line text-fc-fg">{explanation.summary}</p>

	{#if explanation.likely_causes.length > 0}
		<div class="flex flex-col gap-2">
			<h4 class="m-0 text-fc-xs font-medium text-fc-fg-muted">{t('bulle.explain.causes')}</h4>
			<ul class="m-0 flex list-none flex-col gap-3 p-0">
				{#each explanation.likely_causes as cause, i (i)}
					<li class="flex min-w-0 flex-col gap-1">
						<span class="text-fc-sm font-medium wrap-anywhere text-fc-fg">{cause.cause}</span>
						<span class="text-fc-sm wrap-anywhere text-fc-fg-muted">{cause.evidence}</span>
						<Badge tone={tones[cause.confidence] ?? 'neutral'}>{t(`bulle.confidence.${cause.confidence}`)}</Badge>
					</li>
				{/each}
			</ul>
		</div>
	{/if}

	{#if explanation.what_to_check.length > 0}
		<div class="flex flex-col gap-2">
			<h4 class="m-0 text-fc-xs font-medium text-fc-fg-muted">{t('bulle.explain.check')}</h4>
			<ul class="m-0 flex list-disc flex-col gap-1 pl-5 text-fc-sm text-fc-fg">
				{#each explanation.what_to_check as item, i (i)}
					<li class="wrap-anywhere">{item}</li>
				{/each}
			</ul>
		</div>
	{/if}
</div>
