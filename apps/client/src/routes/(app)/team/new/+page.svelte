<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import { Alert, Button, Card, Field, Input, Spinner, icons, toast } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let name = $state('');
	let creating = $state(false);
	let error = $state('');

	const tooShort = $derived(name.trim().length > 0 && name.trim().length < 2);

	async function create(e: Event) {
		e.preventDefault();
		const trimmed = name.trim();
		if (trimmed.length < 2) {
			error = t('newSpace.tooShortLong');
			return;
		}
		creating = true;
		error = '';
		try {
			const ws = await api.workspaces.create(trimmed);
			toast.success(t('newSpace.ready', { name: ws.name }));
			goto(`/team/${ws.id}`);
		} catch (e) {
			error = e instanceof Error ? e.message : t('newSpace.failed');
		} finally {
			creating = false;
		}
	}
</script>

<svelte:head><title>{t('newSpace.pageTitle')}</title></svelte:head>

<div class="mx-auto flex w-full max-w-fc-sm flex-col gap-8">
	<a
		href="/team"
		class="inline-flex items-center gap-1.5 self-start rounded-fc-sm text-fc-sm text-fc-fg-muted transition-colors hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
	>
		<iconify-icon icon={icons.chevronLeft} width="16" height="16" class="block"></iconify-icon>
		{t('newSpace.back')}
	</a>

	<div class="flex flex-col gap-2">
		<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('newSpace.heading')}</h1>
		<p class="text-fc-sm text-fc-fg-muted">{t('newSpace.text')}</p>
	</div>

	<Card>
		<form class="flex flex-col gap-4" onsubmit={create}>
			<Field
				label={t('common.name')}
				error={tooShort ? t('newSpace.tooShort') : undefined}
				helper={t('newSpace.nameHelper')}
			>
				<Input bind:value={name} placeholder={t('newSpace.namePlaceholder')} disabled={creating} />
			</Field>

			{#if error}
				<Alert tone="danger" title={t('newSpace.errorTitle')}>{error}</Alert>
			{/if}

			<Button
				type="submit"
				size="lg"
				icon={icons.plus}
				class="self-start"
				disabled={creating || name.trim().length < 2}
			>
				{#if creating}<Spinner size="sm" />{/if}
				{creating ? t('common.creating') : t('newSpace.submit')}
			</Button>
		</form>
	</Card>
</div>
