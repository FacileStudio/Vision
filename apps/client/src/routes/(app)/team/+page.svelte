<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { WorkspaceMember } from '$lib';
	import Icon from '@iconify/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import { userStore } from '$lib/stores/user.svelte';

	let members = $state<WorkspaceMember[]>([]);
	let inviteEmail = $state('');
	let inviteRole = $state('viewer');
	let inviting = $state(false);
	let error = $state('');
	let editingName = $state(false);
	let wsName = $state('');

	const isOwnerOrAdmin = $derived(
		workspaceStore.current?.role === 'owner' || workspaceStore.current?.role === 'admin'
	);

	async function loadMembers() {
		if (!workspaceStore.current) return;
		try {
			members = await api.workspaces.members(workspaceStore.current.id);
		} catch {}
	}

	onMount(loadMembers);

	async function invite() {
		if (!workspaceStore.current || !inviteEmail.trim()) return;
		inviting = true;
		error = '';
		try {
			const member = await api.workspaces.addMember(workspaceStore.current.id, inviteEmail.trim(), inviteRole);
			members = [...members, member];
			inviteEmail = '';
		} catch (e: any) {
			error = e.message;
		} finally {
			inviting = false;
		}
	}

	async function removeMember(userId: number) {
		if (!workspaceStore.current) return;
		try {
			await api.workspaces.removeMember(workspaceStore.current.id, userId);
			members = members.filter((m) => m.user_id !== userId);
		} catch {}
	}

	async function updateRole(userId: number, role: string) {
		if (!workspaceStore.current) return;
		try {
			await api.workspaces.updateMember(workspaceStore.current.id, userId, role);
			members = members.map((m) => (m.user_id === userId ? { ...m, role } : m));
		} catch {}
	}

	async function saveWorkspaceName() {
		if (!workspaceStore.current || !wsName.trim()) return;
		try {
			const updated = await api.workspaces.update(workspaceStore.current.id, wsName.trim());
			workspaceStore.current = updated;
			workspaceStore.all = workspaceStore.all.map((w) =>
				w.id === updated.id ? updated : w
			);
			editingName = false;
		} catch {}
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

<svelte:head><title>Team — Vision</title></svelte:head>

<div class="max-w-lg">
	<div class="mb-6 flex items-center justify-between">
		<h1 class="text-2xl font-bold">Team</h1>
	</div>

	{#if workspaceStore.current}
		<div class="space-y-6">
			<div class="rounded-lg border p-6">
				<h2 class="mb-1 text-lg font-semibold">Workspace</h2>
				<p class="mb-4 text-sm text-muted-foreground">Manage your workspace name and team members</p>

				{#if editingName}
					<div class="flex items-center gap-2">
						<Input bind:value={wsName} class="flex-1" />
						<Button size="sm" onclick={saveWorkspaceName}>Save</Button>
						<Button size="sm" variant="ghost" onclick={() => (editingName = false)}>Cancel</Button>
					</div>
				{:else}
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium">{workspaceStore.current.name}</span>
						{#if isOwnerOrAdmin}
							<button
								onclick={() => { wsName = workspaceStore.current?.name ?? ''; editingName = true; }}
								class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
							>
								<Icon icon="solar:pen-linear" class="h-3.5 w-3.5" />
							</button>
						{/if}
					</div>
				{/if}
			</div>

			<div class="rounded-lg border p-6">
				<h2 class="mb-1 text-lg font-semibold">Members</h2>
				<p class="mb-4 text-sm text-muted-foreground">
					{members.length} {members.length === 1 ? 'member' : 'members'}
				</p>

				<div class="space-y-2">
					{#each members as member (member.id)}
						<div class="flex items-center gap-3 rounded-lg border border-border/60 px-3 py-2.5">
							<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-border bg-foreground text-xs font-semibold text-background overflow-hidden">
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
								<p class="truncate text-sm font-medium">{member.name || member.email}</p>
								{#if member.name}
									<p class="truncate text-xs text-muted-foreground">{member.email}</p>
								{/if}
							</div>

							{#if isOwnerOrAdmin && member.role !== 'owner' && member.user_id !== Number(userStore.value?.id)}
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
									class="flex size-7 cursor-pointer items-center justify-center shrink-0 rounded-full bg-destructive text-white hover:bg-destructive/90"
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
						{#if error}
							<p class="mt-2 text-sm text-destructive">{error}</p>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	{:else}
		<p class="text-sm text-muted-foreground">Loading workspace...</p>
	{/if}
</div>
