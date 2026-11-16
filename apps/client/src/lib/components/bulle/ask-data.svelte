<script lang="ts">
	import { Alert, Button, Field, Spinner, Textarea } from '@facile/muse';
	import { safeHref } from '$lib/bulle';
	import type { AskRun } from '$lib/ask.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let { ask, onAsk, disabled = false }: { ask: AskRun; onAsk: (question: string) => void; disabled?: boolean } = $props();

	let question = $state('');

	const continueHref = $derived(safeHref(ask.run?.continue_url));
	const ready = $derived(question.trim().length > 0 && !ask.busy && !disabled);

	function submit(event: SubmitEvent) {
		event.preventDefault();
		if (ready) onAsk(question.trim());
	}
</script>

<form class="flex min-w-0 flex-col gap-3" onsubmit={submit}>
	<Field label={t('bulle.ask.label')} helper={t('bulle.ask.hint')}>
		<Textarea bind:value={question} rows={3} maxlength={500} placeholder={t('bulle.ask.placeholder')} />
	</Field>
	<Button type="submit" size="lg" class="self-start" disabled={!ready}>
		{#if ask.starting}<Spinner size="sm" />{/if}
		{t('bulle.ask.submit')}
	</Button>
</form>

{#if ask.error}
	<Alert tone="danger" class="mt-3 wrap-anywhere">{t(ask.error)}</Alert>
{/if}

{#if ask.run && ask.run.state !== 'running'}
	<div class="mt-3 flex min-w-0 flex-col gap-3 rounded-fc-md bg-fc-surface p-4" aria-live="polite">
		{#if ask.run.state === 'done'}
			<h4 class="m-0 text-fc-xs font-medium text-fc-fg-muted">{t('bulle.ask.answer')}</h4>
			<p class="m-0 text-fc-sm wrap-anywhere whitespace-pre-line text-fc-fg">{ask.run.summary}</p>
		{:else if ask.run.state === 'needs_approval'}
			<p class="m-0 text-fc-sm text-fc-fg">{t('bulle.ask.needsApproval')}</p>
		{:else}
			<p class="m-0 text-fc-sm text-fc-fg">{t('bulle.ask.failed')}</p>
		{/if}
		{#if continueHref}
			<Button href={continueHref} target="_blank" rel="noopener noreferrer" variant="secondary" size="lg" class="self-start">
				{t('bulle.ask.continue')}
			</Button>
		{/if}
	</div>
{/if}
