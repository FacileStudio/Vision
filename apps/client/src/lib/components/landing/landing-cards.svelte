<script lang="ts">
	import { LANDING_HEADING } from '@facile/muse';
	import BannerTile from './banner-tile.svelte';
	import GoalsTile from './goals-tile.svelte';
	import SourcesTile from './sources-tile.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let { class: className = '' }: { class?: string } = $props();
</script>

<section class={className} aria-labelledby="cards-title">
	<h2 id="cards-title" class={LANDING_HEADING}>{t('landing.cards.title')}</h2>
	<div class="mt-8 grid gap-4 max-lg:mx-auto max-lg:max-w-md lg:grid-cols-3">
		<BannerTile />
		<SourcesTile />
		<GoalsTile />
	</div>
</section>
