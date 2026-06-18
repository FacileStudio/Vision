<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { Workspace, WorkspaceMember } from '$lib';
	import { userStore } from '$lib/stores/user.svelte';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import Icon from '@iconify/svelte';
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import Users from '@lucide/svelte/icons/users';
	import Globe from '@lucide/svelte/icons/globe';

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

	let showDeleteConfirm = $state(false);
	let deleteText = $state('');
	let deleting = $state(false);

	let showLeaveConfirm = $state(false);
	let leaving = $state(false);

	const wsId = $derived(Number(page.params.id));
	const myUserId = $derived(Number(userStore.value?.id));
	const myRole = $derived(workspace?.role ?? 'viewer');
	const isOwnerOrAdmin = $derived(myRole === 'owner' || myRole === 'admin');
	const isOwner = $derived(myRole === 'owner');

	async function load() {
		loading = true;
		try {
			const [ws, m] = await Promise.all([
				api.workspaces.get(wsId),
				api.workspaces.members(wsId)
			]);
			workspace = ws;
			members = m;
		} catch (e: any) {
			error = e.message || 'Failed to load workspace.';
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
		} catch {
		} finally {
			savingName = false;
		}
	}

	async function invite() {
		if (!inviteEmail.trim() || inviting) return;
		inviting = true;
		inviteError = '';
		try {
			const member = await api.workspaces.addMember(wsId, inviteEmail.trim(), inviteRole);
			members = [...members, member];
			inviteEmail = '';
		} catch (e: any) {
			inviteError = e.message || 'Failed to invite member.';
		} finally {
			inviting = false;
		}
	}

	async function updateRole(userId: number, role: string) {
		try {
			await api.workspaces.updateMember(wsId, userId, role);
			members = members.map((m) => (m.user_id === userId ? { ...m, role } : m));
		} catch {}
	}

	async function removeMember(userId: number) {
		try {
			await api.workspaces.removeMember(wsId, userId);
			members = members.filter((m) => m.user_id !== userId);
		} catch {}
	}

	async function leaveWorkspace() {
		leaving = true;
		try {
			await api.workspaces.leave(wsId);
			workspaceStore.all = workspaceStore.all.filter((w) => w.id !== wsId);
			if (workspaceStore.current?.id === wsId) {
				workspaceStore.current = workspaceStore.all[0] ?? null;
			}
			goto('/team');
		} catch {
		} finally {
			leaving = false;
		}
	}

	async function deleteWorkspace() {
		deleting = true;
		try {
			await api.workspaces.delete(wsId);
			workspaceStore.all = workspaceStore.all.filter((w) => w.id !== wsId);
			if (workspaceStore.current?.id === wsId) {
				workspaceStore.current = workspaceStore.all[0] ?? null;
			}
			goto('/team');
		} catch {
		} finally {
			deleting = false;
		}
	}

	function getInitials(name: string): string {
		return name
			.split(' ')
			.map((w) => w[0])
			.filter(Boolean)
			.slice(0, 2)
			.join('')
			.toUpperCase();
	}

	function roleLabel(role: string): string {
		switch (role) {
			case 'owner': return 'Owner';
			case 'admin': return 'Admin';
			case 'editor': return 'Editor';
			case 'viewer': return 'Viewer';
			default: return role;
		}
	}
</script>

<svelte:head><title>{workspace?.name ?? 'Workspace'} — Vision</title></svelte:head>

<div class="mx-auto max-w-lg">
	<a href="/team" class="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground">
		<ArrowLeft class="h-4 w-4" />
		Back to teams
	</a>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="h-5 w-5 animate-spin rounded-full border-2 border-border border-t-foreground"></div>
		</div>
	{:else if error}
		<div class="rounded-lg border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">{error}</div>
	{:else if workspace}
		<div class="space-y-6">
			<!-- Header -->
			<div class="flex items-start gap-4">
				<div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-foreground text-lg font-bold text-background">
					{workspace.name.charAt(0).toUpperCase()}
				</div>
				<div class="min-w-0 flex-1">
					{#if editingName}
						<div class="flex items-center gap-2">
							<Input bind:value={wsName} class="flex-1" />
							<Button size="sm" onclick={saveName} disabled={savingName}>Save</Button>
							<Button size="sm" variant="ghost" onclick={() => (editingName = false)}>Cancel</Button>
						</div>
					{:else}
						<div class="flex items-center gap-2">
							<h1 class="truncate text-2xl font-bold">{workspace.name}</h1>
							{#if isOwnerOrAdmin}
								<button
									onclick={() => { wsName = workspace?.name ?? ''; editingName = true; }}
									class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
								>
									<Icon icon="solar:pen-linear" class="h-3.5 w-3.5" />
								</button>
							{/if}
						</div>
					{/if}
					<div class="mt-1 flex items-center gap-3 text-xs text-muted-foreground">
						<span class="flex items-center gap-1">
							<Users class="h-3 w-3" />
							{members.length} {members.length === 1 ? 'member' : 'members'}
						</span>
						<span class="flex items-center gap-1">
							<Globe class="h-3 w-3" />
							{workspace.site_count} {workspace.site_count === 1 ? 'site' : 'sites'}
						</span>
						<span class="rounded-full bg-muted px-2 py-0.5 text-[10px] font-medium">
							{roleLabel(workspace.role)}
						</span>
					</div>
				</div>
			</div>

			<Separator />

			<!-- Members -->
			<div>
				<h2 class="mb-3 text-lg font-semibold">Members</h2>
				<div class="space-y-2">
					{#each members as member (member.id)}
						<div class="flex items-center gap-3 rounded-lg border border-border/60 px-3 py-2.5">
							<div class="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-full border border-border bg-foreground text-xs font-semibold text-background">
								{#if member.avatar_url}
									<img
										src="/api{member.avatar_url}"
										alt={member.name || member.email}
										class="h-full w-full object-cover"
									/>
								{:else}
									{getInitials(member.name || member.email)}
								{/if}
							</div>
							<div class="min-w-0 flex-1">
								<p class="truncate text-sm font-medium">
									{member.name || member.email}
									{#if member.user_id === myUserId}
										<span class="text-xs text-muted-foreground">(you)</span>
									{/if}
								</p>
								{#if member.name}
									<p class="truncate text-xs text-muted-foreground">{member.email}</p>
								{/if}
							</div>

							{#if isOwnerOrAdmin && member.role !== 'owner' && member.user_id !== myUserId}
								<select
									value={member.role}
									onchange={(e) => updateRole(member.user_id, (e.target as HTMLSelectElement).value)}
									class="rounded-md border bg-background px-2 py-1 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
								>
									<option value="admin">Admin</option>
									<option value="editor">Editor</option>
									<option value="viewer">Viewer</option>
								</select>
								<button
									onclick={() => removeMember(member.user_id)}
									class="flex size-7 shrink-0 cursor-pointer items-center justify-center rounded-full bg-destructive text-white hover:bg-destructive/90"
									title="Remove member"
								>
									<Icon icon="solar:trash-bin-2-linear" class="h-3.5 w-3.5" />
								</button>
							{:else}
								<span class="shrink-0 rounded-full bg-muted px-2.5 py-0.5 text-xs text-muted-foreground">
									{roleLabel(member.role)}
								</span>
							{/if}
						</div>
					{/each}
				</div>

				{#if isOwnerOrAdmin}
					<div class="mt-4 rounded-lg border border-dashed p-4">
						<p class="mb-3 text-xs font-medium text-muted-foreground">Invite a member</p>
						<form onsubmit={(e) => { e.preventDefault(); invite(); }} class="flex items-end gap-2">
							<div class="flex-1">
								<Input
									bind:value={inviteEmail}
									placeholder="email@example.com"
									type="email"
								/>
							</div>
							<select
								bind:value={inviteRole}
								class="rounded-md border bg-background px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
							>
								<option value="viewer">Viewer</option>
								<option value="editor">Editor</option>
								<option value="admin">Admin</option>
							</select>
							<Button type="submit" size="sm" disabled={inviting || !inviteEmail.trim()}>
								{inviting ? 'Inviting...' : 'Invite'}
							</Button>
						</form>
						{#if inviteError}
							<p class="mt-2 text-sm text-destructive">{inviteError}</p>
						{/if}
					</div>
				{/if}
			</div>

			<Separator />

			<!-- Danger zone -->
			<div>
				{#if !isOwner}
					<div class="rounded-lg border border-destructive/30 p-4">
						<h3 class="mb-1 text-sm font-semibold text-destructive">Leave workspace</h3>
						<p class="mb-3 text-xs text-muted-foreground">
							You will lose access to all sites and data in this workspace.
						</p>
						{#if showLeaveConfirm}
							<div class="flex items-center gap-2">
								<Button size="sm" variant="destructive" onclick={leaveWorkspace} disabled={leaving}>
									{leaving ? 'Leaving...' : 'Confirm leave'}
								</Button>
								<Button size="sm" variant="ghost" onclick={() => (showLeaveConfirm = false)}>Cancel</Button>
							</div>
						{:else}
							<Button size="sm" variant="destructive" onclick={() => (showLeaveConfirm = true)}>
								Leave workspace
							</Button>
						{/if}
					</div>
				{/if}

				{#if isOwner}
					<div class="rounded-lg border border-destructive/30 p-4">
						<h3 class="mb-1 text-sm font-semibold text-destructive">Delete workspace</h3>
						<p class="mb-3 text-xs text-muted-foreground">
							This action is irreversible. All sites must be removed first.
						</p>
						{#if showDeleteConfirm}
							<p class="mb-2 text-xs text-muted-foreground">
								Type <strong>DELETE</strong> to confirm:
							</p>
							<div class="flex items-center gap-2">
								<Input bind:value={deleteText} placeholder="DELETE" class="w-32" />
								<Button
									size="sm"
									variant="destructive"
									onclick={deleteWorkspace}
									disabled={deleting || deleteText !== 'DELETE'}
								>
									{deleting ? 'Deleting...' : 'Delete'}
								</Button>
								<Button size="sm" variant="ghost" onclick={() => { showDeleteConfirm = false; deleteText = ''; }}>
									Cancel
								</Button>
							</div>
						{:else}
							<Button size="sm" variant="destructive" onclick={() => (showDeleteConfirm = true)}>
								Delete workspace
							</Button>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
