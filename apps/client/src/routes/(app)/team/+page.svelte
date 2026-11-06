<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Workspace } from '$lib';
	import { Avatar, Badge, Button, Card, EmptyState, Input, Skeleton, icons, toast } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let workspaces = $state<Workspace[]>([]);
	let loading = $state(true);
	let search = $state('');
	let registreSpaces = $state('');

	const filtered = $derived(
		search.trim()
			? workspaces.filter((w) => w.name.toLowerCase().includes(search.trim().toLowerCase()))
			: workspaces
	);

	onMount(async () => {
		try {
			const [list, cfg] = await Promise.all([
				api.workspaces.list(),
				fetch('/api/auth/config').then((r) => r.json())
			]);
			workspaces = list;
			registreSpaces = cfg.oidc_enabled && cfg.oidc_issuer ? `${cfg.oidc_issuer.replace(/\/$/, '')}/spaces` : '';
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('team.loadFailed'));
		} finally {
			loading = false;
		}
	});

	const roleLabels = $derived<Record<string, string>>({
		owner: t('role.owner'),
		admin: t('role.admin'),
		editor: t('role.editor'),
		viewer: t('role.viewer')
	});

	/* `owner` and `admin` are the two role tones in the shared vocabulary; everything else
	   is a plain member and gets `neutral`. */
	const roleTones = { owner: 'owner', admin: 'admin' } as const;

	function roleTone(role: string) {
		return roleTones[role as keyof typeof roleTones] ?? 'neutral';
	}
</script>

<svelte:head><title>{t('team.pageTitle')}</title></svelte:head>

<div class="flex flex-col gap-10">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex min-w-0 flex-col gap-2">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('team.heading')}</h1>
			<p class="text-fc-sm text-fc-fg-muted">{t('team.subtitle')}</p>
		</div>
		{#if registreSpaces}
			<Button icon={icons.space} href={registreSpaces}>{t('team.manageRegistre')}</Button>
		{:else if !loading}
			<Button icon={icons.plus} onclick={() => goto('/team/new')}>{t('team.newSpace')}</Button>
		{/if}
	</div>

	<section class="flex flex-col gap-4">
		{#if workspaces.length > 3}
			<Input
				bind:value={search}
				placeholder={t('team.searchPlaceholder')}
				aria-label={t('team.searchAria')}
			/>
		{/if}

		{#if loading}
			{#each [0, 1, 2] as row (row)}
				<Skeleton class="h-20 w-full rounded-fc-md" />
			{/each}
		{:else if filtered.length === 0}
			<EmptyState
				icon={icons.space}
				title={search ? t('team.noMatch', { query: search }) : t('team.empty')}
				description={search ? t('team.noMatchHint') : t('team.emptyHint')}
			/>
		{:else}
			{#each filtered as ws (ws.id)}
				<a
					href="/team/{ws.id}"
					class="rounded-fc-md focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
				>
					<Card class="flex items-center gap-4 transition-colors hover:bg-fc-surface">
						<Avatar name={ws.name} />
						<div class="flex min-w-0 flex-1 flex-col gap-1">
							<div class="flex min-w-0 items-center gap-2">
								<span class="truncate text-fc-sm font-medium text-fc-fg">{ws.name}</span>
								<Badge tone={roleTone(ws.role)}>{roleLabels[ws.role] ?? ws.role}</Badge>
							</div>
							<span class="text-fc-xs text-fc-fg-muted">
								{t('team.members', { count: ws.member_count })} · {t('team.sites', {
									count: ws.site_count
								})}
							</span>
						</div>
						<iconify-icon
							icon={icons.arrow}
							width="18"
							height="18"
							class="block shrink-0 text-fc-fg-muted"
						></iconify-icon>
					</Card>
				</a>
			{/each}
		{/if}
	</section>
</div>
