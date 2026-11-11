<script lang="ts">
	import { Badge, Button, Card, icons } from '@facile/muse';
	import type { PendingAction } from '$lib/backend';
	import { formatDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';
	import SiteFavicon from './site-favicon.svelte';

	let {
		action,
		busy,
		onconfirm,
		ondismiss
	}: {
		action: PendingAction;
		busy: boolean;
		onconfirm: (action: PendingAction) => void;
		ondismiss: (action: PendingAction) => void;
	} = $props();

	const target = $derived(action.target);
	const sharing = $derived(action.kind === 'share_site');

	/** Names the client that asked: the host of a client URL, else its id. */
	function clientName(clientId: string): string {
		if (clientId === 'bulle') return 'Bulle';
		try {
			return new URL(clientId).host;
		} catch {
			return clientId || '?';
		}
	}
</script>

<Card class="flex flex-col gap-4">
	<div class="flex flex-wrap items-center justify-between gap-2">
		<Badge tone={sharing ? 'warning' : 'danger'} class="shrink-0 whitespace-nowrap">{t(`pending.kind.${action.kind}`)}</Badge>
		<span class="text-fc-xs text-fc-fg-muted">
			{t('pending.askedBy', { client: clientName(action.client_id) })} · {formatDate(action.created_at, { dateStyle: 'medium', timeStyle: 'short' })}
		</span>
	</div>

	{#if target}
		<a
			href={target.url}
			class="flex items-center gap-3 rounded-fc-md bg-fc-surface p-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
		>
			{#if target.type === 'site'}
				<SiteFavicon domain={target.domain ?? ''} name={target.name} class="size-8 shrink-0" />
			{:else}
				<iconify-icon icon="solar:flag-linear" width="24" height="24" class="shrink-0 text-fc-fg-muted"></iconify-icon>
			{/if}
			<span class="flex min-w-0 flex-1 flex-col gap-1">
				<span class="break-words text-fc-md font-medium text-fc-fg">{target.name}</span>
				<span class="truncate text-fc-sm text-fc-fg-muted">
					{target.type === 'goal' ? t('pending.goalOn', { site: target.site ?? '?' }) : target.domain}
				</span>
			</span>
		</a>
		<p class="text-fc-sm text-fc-fg-muted">{t(`pending.note.${action.kind}`)}</p>
	{:else}
		<p class="break-words text-fc-sm text-fc-fg-muted">{action.summary}</p>
	{/if}

	{#if action.state === 'changed'}
		<p class="text-fc-sm text-fc-warning">{t('pending.changed')}</p>
	{:else if action.state === 'missing'}
		<p class="text-fc-sm text-fc-warning">{t('pending.missing')}</p>
	{/if}

	<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
		<Button variant="ghost" class="w-full sm:w-auto [@media(pointer:coarse)]:h-11" disabled={busy} onclick={() => ondismiss(action)}>
			{t('pending.dismiss')}
		</Button>
		{#if action.state === 'ready'}
			<Button
				variant={sharing ? 'primary' : 'danger'}
				class="w-full sm:w-auto [@media(pointer:coarse)]:h-11"
				disabled={busy}
				icon={sharing ? icons.eye : icons.remove}
				onclick={() => onconfirm(action)}
			>
				{t(`pending.confirm.${action.kind}`)}
			</Button>
		{/if}
	</div>
</Card>
