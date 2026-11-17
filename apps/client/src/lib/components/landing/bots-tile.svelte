<script lang="ts">
	import { Stage, Tile, Window, type Act } from '@facile/mise-en-scene';
	import { Badge, Icon, icons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let shown = $state(0);

	const bots = $derived([
		{ name: 'GPTBot', kind: t('landing.demo.botAi') },
		{ name: 'Googlebot', kind: t('landing.demo.botSearch') },
		{ name: 'HeadlessChrome', kind: t('landing.demo.botHeadless') }
	]);

	function reset() {
		shown = 0;
	}

	function poster() {
		shown = 4;
	}

	async function script(act: Act) {
		await act.wait(900);
		for (let step = 1; step <= 4; step += 1) {
			shown = step;
			await act.wait(700);
		}
		await act.wait(2600);
		shown = 0;
		await act.wait(600);
	}
</script>

<Tile title={t('landing.card.bots.title')} body={t('landing.card.bots.text')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<Window style="top:12px;left:20px;width:280px;height:216px">
				<div class="bt-body">
					<span class="bt-badge" class:bt-on={shown >= 4}>
						<Badge tone="success">{t('landing.demo.realOnly')}</Badge>
					</span>
					{#each bots as bot, index (bot.name)}
						<div class="bt-row" class:bt-out={shown >= index + 1}>
							<div class="bt-text">
								<span class="bt-name">{bot.name}</span>
								<span class="bt-kind">{bot.kind}</span>
							</div>
							<span class="bt-mark">
								<Icon icon={icons.filter} size={12} />{t('landing.demo.filtered')}
							</span>
						</div>
					{/each}
				</div>
			</Window>
		{/snippet}
	</Stage>
</Tile>

<style>
	.bt-body {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 10px 14px;
		line-height: 1.25;
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
		font-size: 12px;
	}
	.bt-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		padding: 5px 10px;
		border-radius: 10px;
		background: var(--mes-w-side);
	}
	.bt-text {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 1px;
		min-width: 0;
		transition: opacity 0.45s ease;
	}
	.bt-name {
		overflow: hidden;
		font-weight: 600;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.bt-kind {
		overflow: hidden;
		color: var(--color-fc-fg-muted);
		font-size: 11px;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.bt-mark {
		display: inline-flex;
		flex: none;
		align-items: center;
		gap: 4px;
		color: var(--color-fc-fg-muted);
		font-size: 11px;
		font-weight: 600;
		white-space: nowrap;
		opacity: 0;
		transform: translateX(6px);
		transition:
			opacity 0.45s ease,
			transform 0.45s ease;
	}
	.bt-out .bt-text {
		opacity: 0.45;
	}
	.bt-out .bt-mark {
		opacity: 1;
		transform: none;
	}
	.bt-badge {
		align-self: flex-start;
		opacity: 0;
		transform: translateY(6px);
		transition:
			opacity 0.45s ease,
			transform 0.45s ease;
	}
	.bt-on {
		opacity: 1;
		transform: none;
	}
	@media (prefers-reduced-motion: reduce) {
		.bt-text,
		.bt-mark,
		.bt-badge {
			transition: none;
		}
	}
</style>
