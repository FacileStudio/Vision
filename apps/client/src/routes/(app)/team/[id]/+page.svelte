<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { Workspace, WorkspaceMember } from '$lib';
	import { userStore } from '$lib/stores/user.svelte';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import { t } from '$lib/i18n/index.svelte';
	import {
		Alert,
		Avatar,
		Badge,
		Button,
		Card,
		ConfirmModal,
		Field,
		IconButton,
		Input,
		Select,
		SettingsRow,
		SettingsSection,
		Skeleton,
		icons,
		toast
	} from '@facile/muse';

	let workspace = $state<Workspace | null>(null);
	let members = $state<WorkspaceMember[]>([]);
	let loading = $state(true);
	let error = $state('');

	let editingName = $state(false);
	let wsName = $state('');
	let savingName = $state(false);

	let inviteEmail = $state('');
	let inviteRole = $state('viewer');
	let inviting = $state(false);
	let inviteError = $state('');

	let removeTarget = $state<WorkspaceMember | null>(null);
	let removeOpen = $state(false);

	let deleteOpen = $state(false);
	let deleteText = $state('');

	let leaveOpen = $state(false);

	let registreSpaces = $state('');

	const wsId = $derived(Number(page.params.id));
	const myUserId = $derived(Number(userStore.value?.id));
	const myRole = $derived(workspace?.role ?? 'viewer');
	const isOwnerOrAdmin = $derived(myRole === 'owner' || myRole === 'admin');
	const isOwner = $derived(myRole === 'owner');
	const canEditMembers = $derived(isOwnerOrAdmin && !registreSpaces);
	const canEditSpace = $derived(!workspace?.manage_url);
	const manageURL = $derived(workspace?.manage_url || registreSpaces);

	const roleLabels = $derived<Record<string, string>>({
		owner: t('role.owner'),
		admin: t('role.admin'),
		editor: t('role.editor'),
		viewer: t('role.viewer')
	});

	const roleTones = { owner: 'owner', admin: 'admin' } as const;

	function roleTone(role: string) {
		return roleTones[role as keyof typeof roleTones] ?? 'neutral';
	}

	async function load() {
		loading = true;
		try {
			const [ws, m, cfg] = await Promise.all([
				api.workspaces.get(wsId),
				api.workspaces.members(wsId),
				fetch('/api/auth/config').then((r) => r.json())
			]);
			workspace = ws;
			members = m;
			registreSpaces = cfg.oidc_enabled && cfg.oidc_issuer ? `${cfg.oidc_issuer.replace(/\/$/, '')}/spaces` : '';
		} catch (e) {
			error = e instanceof Error ? e.message : t('space.loadFailed');
		} finally {
			loading = false;
		}
	}

	onMount(load);

	async function saveName() {
		if (!wsName.trim() || savingName) return;
		savingName = true;
		try {
			const updated = await api.workspaces.update(wsId, wsName.trim());
			workspace = updated;
			workspaceStore.all = workspaceStore.all.map((w) => (w.id === updated.id ? updated : w));
			editingName = false;
			toast.success(t('space.renamed'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('space.renameFailed'));
		} finally {
			savingName = false;
		}
	}

	async function invite(e: Event) {
		e.preventDefault();
		if (!inviteEmail.trim() || inviting) return;
		inviting = true;
		inviteError = '';
		try {
			const member = await api.workspaces.addMember(wsId, inviteEmail.trim(), inviteRole);
			members = [...members, member];
			toast.success(t('space.added', { email: inviteEmail.trim() }));
			inviteEmail = '';
		} catch (e) {
			inviteError = e instanceof Error ? e.message : t('space.addFailed');
		} finally {
			inviting = false;
		}
	}

	async function updateRole(userId: number, role: string) {
		try {
			await api.workspaces.updateMember(wsId, userId, role);
			members = members.map((m) => (m.user_id === userId ? { ...m, role } : m));
			toast.success(t('space.roleUpdated'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('space.roleFailed'));
		}
	}

	async function removeMember() {
		const target = removeTarget;
		if (!target) return;
		await api.workspaces.removeMember(wsId, target.user_id);
		members = members.filter((m) => m.user_id !== target.user_id);
		removeTarget = null;
		toast.neutral(t('space.removed', { name: target.name || target.email }));
	}

	function forgetSpace() {
		workspaceStore.all = workspaceStore.all.filter((w) => w.id !== wsId);
		goto('/team');
	}

	async function leaveWorkspace() {
		await api.workspaces.leave(wsId);
		toast.neutral(t('space.left'));
		forgetSpace();
	}

	/*
	 * The typed confirmation stays inside the dialog rather than in front of it: rejecting
	 * the promise is what keeps ConfirmModal open, so a mistyped word costs a retry instead
	 * of reopening the whole flow.
	 */
	async function deleteWorkspace() {
		if (deleteText !== 'DELETE') {
			throw new Error('confirmation text does not match');
		}
		await api.workspaces.delete(wsId);
		toast.neutral(t('space.deleted'));
		deleteText = '';
		forgetSpace();
	}
</script>

<svelte:head><title>{workspace?.name ?? t('space.fallbackTitle')} — Vision</title></svelte:head>

<div class="flex flex-col gap-10">
	<a
		href="/team"
		class="inline-flex items-center gap-1.5 self-start rounded-fc-sm text-fc-sm text-fc-fg-muted transition-colors hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
	>
		<iconify-icon icon={icons.chevronLeft} width="16" height="16" class="block"></iconify-icon>
		{t('newSpace.back')}
	</a>

	{#if loading}
		<Skeleton class="h-24 w-full rounded-fc-md" />
		<Skeleton class="h-64 w-full rounded-fc-md" />
	{:else if error}
		<Alert tone="danger" title={t('space.openFailed')}>{error}</Alert>
	{:else if workspace}
		<div class="flex items-start gap-4">
			<Avatar name={workspace.name} size="lg" />
			<div class="flex min-w-0 flex-1 flex-col gap-2">
				{#if editingName}
					<div class="flex flex-wrap items-center gap-2">
						<Input bind:value={wsName} class="min-w-48 flex-1" aria-label={t('space.nameAria')} />
						<Button icon={icons.check} disabled={savingName} onclick={saveName}>{t('common.save')}</Button>
						<Button variant="ghost" onclick={() => (editingName = false)}>{t('common.cancel')}</Button>
					</div>
				{:else}
					<div class="flex min-w-0 items-center gap-2">
						<h1 class="truncate text-fc-2xl font-semibold text-fc-fg">{workspace.name}</h1>
						{#if isOwnerOrAdmin && canEditSpace}
							<IconButton
								variant="ghost"
								aria-label={t('space.renameAria')}
								onclick={() => {
									wsName = workspace?.name ?? '';
									editingName = true;
								}}
							>
								<iconify-icon icon={icons.edit} width="18" height="18" class="block"></iconify-icon>
							</IconButton>
						{/if}
					</div>
				{/if}
				<div class="flex flex-wrap items-center gap-3 text-fc-sm text-fc-fg-muted">
					<span>
						{t('team.members', { count: members.length })} · {t('team.sites', {
							count: workspace.site_count
						})}
					</span>
					<Badge tone={roleTone(workspace.role)}>{roleLabels[workspace.role] ?? workspace.role}</Badge>
				</div>
			</div>
		</div>

		<SettingsSection
			title={t('space.membersTitle')}
			description={registreSpaces ? t('space.membersManaged') : t('space.membersLocal')}
		>
			{#if registreSpaces}
				<SettingsRow label={t('space.manageLabel')} description={t('space.manageText')}>
					<Button variant="secondary" icon={icons.space} href={manageURL}>{t('team.manageRegistre')}</Button>
				</SettingsRow>
			{/if}
			{#each members as member (member.id)}
				<SettingsRow class="*:flex-1 *:flex-wrap">
					<div class="flex min-w-0 flex-1 items-center gap-3">
						<Avatar src={member.avatar_url || undefined} name={member.name || member.email} size="sm" />
						<div class="flex min-w-0 flex-col gap-1">
							<span class="truncate text-fc-sm font-medium text-fc-fg">{member.name || member.email}</span>
							{#if member.name}
								<p class="truncate text-fc-xs text-fc-fg-muted">{member.email}</p>
							{/if}
						</div>
					</div>
					{#if member.user_id === myUserId}
						<span class="text-fc-xs text-fc-fg-muted">{t('common.you')}</span>
					{/if}
					{#if canEditMembers && member.role !== 'owner' && member.user_id !== myUserId}
						<Select
							value={member.role}
							aria-label={t('space.roleAria', { name: member.name || member.email })}
							class="w-32"
							onchange={(e) => updateRole(member.user_id, e.currentTarget.value)}
						>
							<option value="admin">{t('role.admin')}</option>
							<option value="editor">{t('role.editor')}</option>
							<option value="viewer">{t('role.viewer')}</option>
						</Select>
						<Button
							variant="ghost-danger"
							icon={icons.remove}
							aria-label={t('space.removeAria', { name: member.name || member.email })}
							onclick={() => {
								removeTarget = member;
								removeOpen = true;
							}}
						>
							{t('common.remove')}
						</Button>
					{:else}
						<Badge tone={roleTone(member.role)}>{roleLabels[member.role] ?? member.role}</Badge>
					{/if}
				</SettingsRow>
			{/each}
		</SettingsSection>

		{#if canEditMembers}
			<SettingsSection
				title={t('space.addTitle')}
				description={t('space.addText')}
			>
				<form class="flex flex-col gap-3 sm:flex-row sm:items-end" onsubmit={invite}>
					<div class="min-w-0 flex-1">
						<Field label={t('common.email')}>
							<Input
								bind:value={inviteEmail}
								type="email"
								placeholder={t('space.addEmailPlaceholder')}
								disabled={inviting}
							/>
						</Field>
					</div>
					<Field label={t('common.role')}>
						<Select bind:value={inviteRole} class="sm:w-40" disabled={inviting}>
							<option value="viewer">{t('role.viewer')}</option>
							<option value="editor">{t('role.editor')}</option>
							<option value="admin">{t('role.admin')}</option>
						</Select>
					</Field>
					<Button type="submit" icon={icons.plus} disabled={inviting || !inviteEmail.trim()}>
						{inviting ? t('common.adding') : t('common.add')}
					</Button>
				</form>

				{#if inviteError}
					<Alert tone="danger" title={t('space.addErrorTitle')}>{inviteError}</Alert>
				{/if}
			</SettingsSection>
		{/if}

		{#if canEditSpace && (isOwner || !registreSpaces)}
			<SettingsSection title={t('space.dangerTitle')} description={t('space.dangerText')}>
				{#if isOwner}
					<SettingsRow
						label={t('space.deleteLabel')}
						description={t('space.deleteText')}
					>
						<Button variant="ghost-danger" icon={icons.remove} onclick={() => (deleteOpen = true)}>
							{t('space.deleteButton')}
						</Button>
					</SettingsRow>
				{:else}
					<SettingsRow
						label={t('space.leaveLabel')}
						description={t('space.leaveText')}
					>
						<Button variant="ghost-danger" icon={icons.logout} onclick={() => (leaveOpen = true)}>
							{t('space.leaveButton')}
						</Button>
					</SettingsRow>
				{/if}
			</SettingsSection>
		{/if}
	{/if}
</div>

<ConfirmModal
	bind:open={removeOpen}
	tone="danger"
	title={t('space.removeTitle', { name: removeTarget?.name || removeTarget?.email || t('space.thisMember') })}
	description={t('space.removeBody')}
	confirmLabel={t('common.remove')}
	cancelLabel={t('space.removeKeep')}
	onConfirm={removeMember}
	onCancel={() => (removeTarget = null)}
/>

<ConfirmModal
	bind:open={leaveOpen}
	tone="danger"
	title={t('space.leaveTitle', { name: workspace?.name ?? t('space.thisSpace') })}
	description={t('space.leaveBody')}
	confirmLabel={t('space.leaveButton')}
	cancelLabel={t('space.leaveStay')}
	onConfirm={leaveWorkspace}
/>

<ConfirmModal
	bind:open={deleteOpen}
	tone="danger"
	title={t('space.deleteTitle', { name: workspace?.name ?? t('space.thisSpace') })}
	description={t('space.deleteBody')}
	confirmLabel={t('space.deleteButton')}
	cancelLabel={t('common.keepIt')}
	onConfirm={deleteWorkspace}
	onCancel={() => (deleteText = '')}
>
	<Field label={t('space.deleteType')}>
		<Input bind:value={deleteText} placeholder="DELETE" autocomplete="off" />
	</Field>
</ConfirmModal>
