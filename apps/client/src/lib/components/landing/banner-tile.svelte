<script lang="ts">
	import { Bar, Stage, Tile, Window, type Act } from '@facile/mise-en-scene';
	import { Badge } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let banner = $state(true);
	let clean = $state(false);

	function reset() {
		banner = true;
		clean = false;
	}

	function poster() {
		banner = false;
		clean = true;
	}

	async function script(act: Act) {
		await act.wait(1600);
		banner = false;
		await act.wait(600);
		clean = true;
		await act.wait(2800);
		clean = false;
		await act.wait(300);
		banner = true;
		await act.wait(600);
	}
</script>

<Tile title={t('landing.card.banner.title')} body={t('landing.card.banner.text')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<Window style="top:12px;left:20px;width:280px;height:216px">
				<div class="bt-site">
					<div class="bt-page">
						<b class="bt-brand">atelier-lune.fr</b>
						<span class="bt-hero"></span>
						<Bar w="82%" />
						<Bar w="64%" />
						<Bar w="72%" />
						<span class="bt-badge" class:bt-on={clean}>
							<Badge tone="success">{t('landing.demo.noCookie')}</Badge>
						</span>
					</div>
					<div class="bt-banner" class:bt-gone={!banner}>
						<p>{t('landing.demo.consent')}</p>
						<div class="bt-buttons">
							<span class="bt-button">{t('landing.demo.decline')}</span>
							<span class="bt-button bt-primary">{t('landing.demo.accept')}</span>
						</div>
					</div>
				</div>
			</Window>
		{/snippet}
	</Stage>
</Tile>

<style>
	.bt-site {
		position: relative;
		flex: 1;
		min-height: 0;
	}
	.bt-page {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 12px 14px;
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
	}
	.bt-brand {
		font-family: var(--font-fc-title);
		font-size: 13px;
		font-weight: 700;
	}
	.bt-hero {
		height: 54px;
		border-radius: 8px;
		background: var(--mes-w-side);
	}
	.bt-badge,
	.bt-banner {
		transition:
			opacity 0.45s ease,
			transform 0.45s ease;
	}
	.bt-badge {
		align-self: flex-start;
		margin-top: 4px;
		opacity: 0;
		transform: translateY(6px);
	}
	.bt-on {
		opacity: 1;
		transform: none;
	}
	.bt-banner {
		position: absolute;
		right: 10px;
		bottom: 10px;
		left: 10px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 10px 12px;
		border-radius: 10px;
		background: var(--mes-w-side);
		box-shadow: 0 8px 24px -8px rgb(0 0 0 / 0.3);
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
		font-size: 12px;
	}
	.bt-gone {
		opacity: 0;
		transform: translateY(24px);
	}
	.bt-buttons {
		display: flex;
		justify-content: flex-end;
		gap: 6px;
	}
	.bt-button {
		padding: 4px 10px;
		border-radius: 999px;
		background: var(--mes-w-body);
		font-size: 11px;
		font-weight: 600;
	}
	.bt-primary {
		background: var(--color-fc-accent);
		color: var(--color-fc-accent-fg);
	}
	@media (prefers-reduced-motion: reduce) {
		.bt-badge,
		.bt-banner {
			transition: none;
		}
	}
</style>
