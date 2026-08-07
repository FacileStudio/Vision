<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { Workspace, WorkspaceMember } from '$lib';
	import { userStore } from '$lib/stores/user.svelte';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
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

	const wsId = $derived(Number(page.params.id));
	const myUserId = $derived(Number(userStore.value?.id));
	const myRole = $derived(workspace?.role ?? 'viewer');
	const isOwnerOrAdmin = $derived(myRole === 'owner' || myRole === 'admin');
	const isOwner = $derived(myRole === 'owner');

	const roleLabels: Record<string, string> = {
		owner: 'Owner',
		admin: 'Admin',
		editor: 'Editor',
		viewer: 'Viewer'
	};

	const roleTones = { owner: 'owner', admin: 'admin' } as const;

	function roleTone(role: string) {
		return roleTones[role as keyof typeof roleTones] ?? 'neutral';
	}

	async function load() {
		loading = true;
		try {
			const [ws, m] = await Promise.all([
				api.workspaces.get(wsId),
				api.workspaces.members(wsId)
			]);
			workspace = ws;
			members = m;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load this space.';
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
			if (workspaceStore.current?.id === updated.id) workspaceStore.current = updated;
			editingName = false;
			toast.success('Space renamed.');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Could not rename the space.');
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
			toast.success(`${inviteEmail.trim()} was added.`);
			inviteEmail = '';
		} catch (e) {
			inviteError = e instanceof Error ? e.message : 'Could not add that person.';
		} finally {
			inviting = false;
		}
	}

	async function updateRole(userId: number, role: string) {
		try {
			await api.workspaces.updateMember(wsId, userId, role);
			members = members.map((m) => (m.user_id === userId ? { ...m, role } : m));
			toast.success('Role updated.');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Could not change that role.');
		}
	}

	async function removeMember() {
		const target = removeTarget;
		if (!target) return;
		await api.workspaces.removeMember(wsId, target.user_id);
		members = members.filter((m) => m.user_id !== target.user_id);
		removeTarget = null;
		toast.neutral(`${target.name || target.email} no longer has access.`);
	}

	function forgetSpace() {
		workspaceStore.all = workspaceStore.all.filter((w) => w.id !== wsId);
		if (workspaceStore.current?.id === wsId) {
			workspaceStore.current = workspaceStore.all[0] ?? null;
		}
		goto('/team');
	}

	async function leaveWorkspace() {
		await api.workspaces.leave(wsId);
		toast.neutral('You left the space.');
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
		toast.neutral('Space deleted.');
		deleteText = '';
		forgetSpace();
	}
</script>

<svelte:head><title>{workspace?.name ?? 'Space'} — Vision</title></svelte:head>

<div class="flex flex-col gap-10">
	<a
		href="/team"
		class="inline-flex items-center gap-1.5 self-start rounded-fc-sm text-fc-sm text-fc-fg-muted transition-colors hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
	>
		<iconify-icon icon={icons.chevronLeft} width="16" height="16" class="block"></iconify-icon>
		Back to teams
	</a>

	{#if loading}
		<Skeleton class="h-24 w-full rounded-fc-md" />
		<Skeleton class="h-64 w-full rounded-fc-md" />
	{:else if error}
		<Alert tone="danger" title="Could not open this space">{error}</Alert>
	{:else if workspace}
		<div class="flex items-start gap-4">
			<Avatar name={workspace.name} size="lg" />
			<div class="flex min-w-0 flex-1 flex-col gap-2">
				{#if editingName}
					<div class="flex flex-wrap items-center gap-2">
						<Input bind:value={wsName} class="min-w-48 flex-1" aria-label="Space name" />
						<Button icon={icons.check} disabled={savingName} onclick={saveName}>Save</Button>
						<Button variant="ghost" onclick={() => (editingName = false)}>Cancel</Button>
					</div>
				{:else}
					<div class="flex min-w-0 items-center gap-2">
						<h1 class="truncate text-fc-2xl font-semibold text-fc-fg">{workspace.name}</h1>
						{#if isOwnerOrAdmin}
							<IconButton
								variant="ghost"
								aria-label="Rename space"
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
						{members.length}
						{members.length === 1 ? 'member' : 'members'} · {workspace.site_count}
						{workspace.site_count === 1 ? 'site' : 'sites'}
					</span>
					<Badge tone={roleTone(workspace.role)}>{roleLabels[workspace.role] ?? workspace.role}</Badge>
				</div>
			</div>
		</div>

		<SettingsSection
			title="Members"
			description="Owners cannot be removed, and nobody can change their own role."
		>
			{#each members as member (member.id)}
				<SettingsRow
					label={member.name || member.email}
					description={member.name ? member.email : undefined}
				>
					{#if member.user_id === myUserId}
						<span class="text-fc-xs text-fc-fg-muted">you</span>
					{/if}
					{#if isOwnerOrAdmin && member.role !== 'owner' && member.user_id !== myUserId}
						<Select
							value={member.role}
							aria-label="Role for {member.name || member.email}"
							class="w-32"
							onchange={(e) => updateRole(member.user_id, e.currentTarget.value)}
						>
							<option value="admin">Admin</option>
							<option value="editor">Editor</option>
							<option value="viewer">Viewer</option>
						</Select>
						<Button
							variant="ghost-danger"
							icon={icons.remove}
							aria-label="Remove {member.name || member.email}"
							onclick={() => {
								removeTarget = member;
								removeOpen = true;
							}}
						>
							Remove
						</Button>
					{:else}
						<Badge tone={roleTone(member.role)}>{roleLabels[member.role] ?? member.role}</Badge>
					{/if}
				</SettingsRow>
			{/each}
		</SettingsSection>

		{#if isOwnerOrAdmin}
			<SettingsSection
				title="Add someone"
				description="They need a Vision account already — adding grants access immediately."
			>
				<form class="flex flex-col gap-3 sm:flex-row sm:items-end" onsubmit={invite}>
					<div class="min-w-0 flex-1">
						<Field label="Email">
							<Input bind:value={inviteEmail} type="email" placeholder="name@studio.com" disabled={inviting} />
						</Field>
					</div>
					<Field label="Role">
						<Select bind:value={inviteRole} class="sm:w-40" disabled={inviting}>
							<option value="viewer">Viewer</option>
							<option value="editor">Editor</option>
							<option value="admin">Admin</option>
						</Select>
					</Field>
					<Button type="submit" icon={icons.plus} disabled={inviting || !inviteEmail.trim()}>
						{inviting ? 'Adding…' : 'Add'}
					</Button>
				</form>

				{#if inviteError}
					<Alert tone="danger" title="Not added">{inviteError}</Alert>
				{/if}
			</SettingsSection>
		{/if}

		<SettingsSection title="Danger zone" description="Irreversible, and nobody can undo it for you.">
			{#if isOwner}
				<SettingsRow
					label="Delete this space"
					description="Every site must be removed first. Members lose access the moment it goes."
				>
					<Button variant="danger" icon={icons.remove} onclick={() => (deleteOpen = true)}>
						Delete space
					</Button>
				</SettingsRow>
			{:else}
				<SettingsRow
					label="Leave this space"
					description="You lose access to every site in it. An owner can add you back."
				>
					<Button variant="danger" icon={icons.logout} onclick={() => (leaveOpen = true)}>
						Leave space
					</Button>
				</SettingsRow>
			{/if}
		</SettingsSection>
	{/if}
</div>

<ConfirmModal
	bind:open={removeOpen}
	tone="danger"
	title="Remove {removeTarget?.name || removeTarget?.email || 'this member'}?"
	description="They lose access to every site in this space immediately. The stats they looked at are unaffected — this only revokes access."
	confirmLabel="Remove"
	cancelLabel="Keep access"
	onConfirm={removeMember}
	onCancel={() => (removeTarget = null)}
/>

<ConfirmModal
	bind:open={leaveOpen}
	tone="danger"
	title="Leave {workspace?.name ?? 'this space'}?"
	description="You lose access to every site in it, and only an owner can add you back."
	confirmLabel="Leave space"
	cancelLabel="Stay"
	onConfirm={leaveWorkspace}
/>

<ConfirmModal
	bind:open={deleteOpen}
	tone="danger"
	title="Delete {workspace?.name ?? 'this space'}?"
	description="The space and its membership are gone for good. Sites must already be removed."
	confirmLabel="Delete space"
	cancelLabel="Keep it"
	onConfirm={deleteWorkspace}
	onCancel={() => (deleteText = '')}
>
	<Field label="Type DELETE to confirm">
		<Input bind:value={deleteText} placeholder="DELETE" autocomplete="off" />
	</Field>
</ConfirmModal>
