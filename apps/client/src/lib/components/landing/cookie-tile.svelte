<script lang="ts">
	import { Bar, Stage, Tile, Window, type Act } from '@facile/mise-en-scene';
	import { Badge, Icon, icons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let shown = $state(0);

	const facts = $derived([t('landing.demo.cookies'), t('landing.demo.adTrackers')]);

	function reset() {
		shown = 0;
	}

	function poster() {
		shown = 3;
	}

	async function script(act: Act) {
		await act.wait(900);
		for (let step = 1; step <= 3; step += 1) {
			shown = step;
			await act.wait(700);
		}
		await act.wait(2600);
		shown = 0;
		await act.wait(600);
	}
</script>

<Tile title={t('landing.card.cookie.title')} body={t('landing.card.cookie.text')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<Window style="top:12px;left:20px;width:280px;height:216px">
				<div class="ct-page">
					<b class="ct-brand">atelier-lune.fr</b>
					<span class="ct-hero"></span>
					<Bar w="82%" />
					<Bar w="64%" />
					<span class="ct-badge" class:ct-on={shown >= 1}>
						<Badge tone="success">{t('landing.demo.noCookie')}</Badge>
					</span>
					<ul class="ct-facts">
						{#each facts as fact, index (index)}
							<li class="ct-fact" class:ct-on={shown >= index + 2}>
								<Icon icon={icons.check} size={14} />{fact}
							</li>
						{/each}
					</ul>
				</div>
			</Window>
		{/snippet}
	</Stage>
</Tile>

<style>
	.ct-page {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 12px 14px;
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
	}
	.ct-brand {
		font-family: var(--font-fc-title);
		font-size: 13px;
		font-weight: 700;
	}
	.ct-hero {
		height: 44px;
		border-radius: 8px;
		background: var(--mes-w-side);
	}
	.ct-badge,
	.ct-fact {
		opacity: 0;
		transform: translateY(6px);
		transition:
			opacity 0.45s ease,
			transform 0.45s ease;
	}
	.ct-badge {
		align-self: flex-start;
		margin-top: 2px;
	}
	.ct-facts {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.ct-fact {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 3px 8px;
		border-radius: 999px;
		background: var(--mes-w-side);
		font-size: 11px;
		font-weight: 600;
		white-space: nowrap;
	}
	.ct-on {
		opacity: 1;
		transform: none;
	}
	@media (prefers-reduced-motion: reduce) {
		.ct-badge,
		.ct-fact {
			transition: none;
		}
	}
</style>
