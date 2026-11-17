<script lang="ts">
	import { LANDING_HEADING } from '@facile/muse';
	import BotsTile from './bots-tile.svelte';
	import CookieTile from './cookie-tile.svelte';
	import GoalsTile from './goals-tile.svelte';
	import SourcesTile from './sources-tile.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let { class: className = '' }: { class?: string } = $props();
</script>

<section class={className} aria-labelledby="cards-title">
	<h2 id="cards-title" class={LANDING_HEADING}>{t('landing.cards.title')}</h2>
	<div class="mt-8 grid gap-4 max-md:mx-auto max-md:max-w-md md:grid-cols-2">
		<CookieTile />
		<SourcesTile />
		<GoalsTile />
		<BotsTile />
	</div>
</section>
