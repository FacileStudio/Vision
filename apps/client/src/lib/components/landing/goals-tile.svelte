<script lang="ts">
	import { Roll, Stage, Tile, Window, type Act } from '@facile/mise-en-scene';
	import { Icon, icons } from '@facile/muse';
	import { fmt } from '$lib/i18n/format';
	import { locale, t } from '$lib/i18n/index.svelte';
	import { decimal } from '$lib/utils/analytics';

	const VISITORS = 1284;
	const START = { sale: 38, signup: 112 };

	let sale = $state(START.sale);
	let signup = $state(START.signup);
	let ping = $state(false);

	const goals = $derived([
		{ id: 'sale', name: t('landing.demo.goalSale'), path: '/merci', count: sale },
		{ id: 'signup', name: t('landing.demo.goalSignup'), path: '/bienvenue', count: signup }
	]);

	function reset() {
		sale = START.sale;
		signup = START.signup;
		ping = false;
	}

	function poster() {
		sale = START.sale + 3;
		signup = START.signup + 2;
	}

	async function script(act: Act) {
		const beats: Array<'sale' | 'signup'> = ['sale', 'signup', 'sale', 'sale', 'signup'];
		await act.wait(900);
		for (const beat of beats) {
			if (beat === 'sale') sale += 1;
			else signup += 1;
			ping = beat === 'sale';
			await act.wait(1100);
			ping = false;
			await act.wait(200);
		}
		await act.wait(1600);
	}

	function line(count: number): string {
		return t('goals.row', { count, rate: `${decimal((count / VISITORS) * 100, locale())}%` });
	}
</script>

<Tile title={t('landing.card.goals.title')} body={t('landing.card.goals.text')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<Window style="top:12px;left:20px;width:280px;height:216px">
				<div class="gt-body">
					<p class="gt-head"><Icon icon={icons.bolt} size={16} />{t('goals.title')}</p>
					{#each goals as goal (goal.id)}
						<div class="gt-row">
							<div class="gt-text">
								<span class="gt-name">
									{goal.name}
									<span class="gt-ping" class:gt-on={ping && goal.id === 'sale'}>+1</span>
								</span>
								<span class="gt-kind"><code>{goal.path}</code>{line(goal.count)}</span>
							</div>
							<Roll value={fmt(goal.count)} class="gt-count tabular-nums" />
						</div>
					{/each}
				</div>
			</Window>
		{/snippet}
	</Stage>
</Tile>

<style>
	.gt-body {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 12px 14px;
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
		font-size: 12px;
	}
	.gt-head {
		display: flex;
		align-items: center;
		gap: 6px;
		font-weight: 600;
	}
	.gt-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 10px;
		background: var(--mes-w-side);
	}
	.gt-text {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 3px;
		min-width: 0;
	}
	.gt-name {
		display: flex;
		align-items: center;
		gap: 6px;
		font-weight: 600;
		white-space: nowrap;
	}
	.gt-kind {
		overflow: hidden;
		color: var(--color-fc-fg-muted);
		font-size: 11px;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.gt-kind code {
		margin-right: 6px;
		color: var(--color-fc-fg);
		font-family: var(--font-fc-mono);
		font-size: 11px;
	}
	.gt-ping {
		color: var(--color-fc-success);
		font-size: 11px;
		font-weight: 700;
		opacity: 0;
		transform: translateY(4px);
		transition:
			opacity 0.3s ease,
			transform 0.3s ease;
	}
	.gt-on {
		opacity: 1;
		transform: none;
	}
	.gt-row :global(.gt-count) {
		flex: none;
		font-family: var(--font-fc-title);
		font-size: 22px;
		font-weight: 700;
	}
	@media (prefers-reduced-motion: reduce) {
		.gt-ping {
			transition: none;
		}
	}
</style>
