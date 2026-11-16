<script lang="ts">
	import { onDestroy } from 'svelte';
	import { Field, IdeasPage, IdeasUpsell, Select, toast, type ApprovalRequest, type Idea } from '@facile/muse';
	import { upgradeLink } from '$lib/bulle';
	import { ideasStore } from '$lib/ideas.svelte';
	import { toIdea } from '$lib/ideas';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let chosen = $state<number | null>(null);

	const spaceId = $derived(chosen ?? workspaceStore.all[0]?.id ?? null);
	const cards = $derived((ideasStore.page?.ideas ?? []).map((idea) => toIdea(idea, ideasStore.runs[idea.id], t)));
	const upgradeHref = $derived(upgradeLink(ideasStore.upgradeUrl || ideasStore.page?.upgrade_url));

	const examples = $derived([
		{ kind: 'anomaly', title: t('ideas.example.drop.title'), why: t('ideas.example.drop.why') },
		{ kind: 'digest', title: t('ideas.example.digest.title'), why: t('ideas.example.digest.why') },
		{ kind: 'goal', title: t('ideas.example.goal.title'), why: t('ideas.example.goal.why') }
	]);

	$effect(() => {
		if (spaceId !== null) void ideasStore.load(spaceId);
	});

	onDestroy(() => ideasStore.stop());

	async function dismiss(idea: Idea) {
		if (await ideasStore.dismiss(idea.id)) toast.info(t('ideas.dismissed'));
	}

	function answer(idea: Idea, approval: ApprovalRequest, approve: boolean) {
		void ideasStore.decide(idea.id, approval.id, approve);
	}
</script>

<svelte:head>
	<title>{t('ideas.title')}</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
	{#if workspaceStore.all.length > 1}
		<Field label={t('ideas.space')} class="max-w-xs">
			<Select value={String(spaceId)} onchange={(e) => (chosen = Number(e.currentTarget.value))}>
				{#each workspaceStore.all as workspace (workspace.id)}
					<option value={String(workspace.id)}>{workspace.name}</option>
				{/each}
			</Select>
		</Field>
	{/if}
	{#if ideasStore.upsell}
		<IdeasUpsell {upgradeHref} {examples} />
	{:else}
		<IdeasPage
			ideas={cards}
			loading={ideasStore.loading}
			generating={ideasStore.generating}
			error={ideasStore.error ? t(ideasStore.error) : null}
			lastGeneratedAt={ideasStore.page?.last_batch_at ?? null}
			busyIds={ideasStore.busyIds}
			description={t('ideas.description')}
			onGenerate={() => void ideasStore.generate()}
			onRetry={() => void (spaceId !== null && ideasStore.load(spaceId))}
			onRun={(idea) => void ideasStore.run(idea.id)}
			onDismiss={dismiss}
			onApprove={(idea, approval) => answer(idea, approval, true)}
			onDecline={(idea, approval) => answer(idea, approval, false)}
		/>
	{/if}
</div>
