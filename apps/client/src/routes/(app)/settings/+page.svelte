<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api, logout as logoutSession } from '$lib';
	import type { UserProfile } from '$lib';
	import { userStore } from '$lib/stores/user.svelte';
	import {
		Alert,
		Button,
		Field,
		Input,
		ProfileCard,
		SettingsRow,
		SettingsSection,
		Skeleton,
		Spinner,
		icons,
		toast
	} from '@facile/muse';

	let profile = $state<UserProfile | null>(null);
	let loadError = $state('');

	let name = $state('');
	let email = $state('');
	let savingProfile = $state(false);

	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordError = $state('');
	let savingPassword = $state(false);

	let ssoOnly = $state(false);

	const mismatch = $derived(
		confirmPassword.length > 0 && newPassword !== confirmPassword
			? 'The two passwords do not match.'
			: undefined
	);

	onMount(async () => {
		try {
			profile = await api.auth.me();
			name = profile.name ?? '';
			email = profile.email;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Could not load your profile.';
		}
		try {
			const cfg = await fetch('/api/auth/config').then((r) => r.json());
			ssoOnly = cfg.sso_only ?? false;
		} catch {}
	});

	async function updateProfile(e: Event) {
		e.preventDefault();
		savingProfile = true;
		try {
			profile = await api.auth.updateProfile(name, email);
			name = profile.name ?? '';
			email = profile.email;
			userStore.value = profile;
			toast.success('Profile updated.');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Could not save your profile.');
		} finally {
			savingProfile = false;
		}
	}

	async function changePassword(e: Event) {
		e.preventDefault();
		passwordError = '';
		if (newPassword !== confirmPassword) {
			passwordError = 'The two passwords do not match.';
			return;
		}
		savingPassword = true;
		try {
			await api.auth.changePassword(currentPassword, newPassword);
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
			toast.success('Password changed.');
		} catch (e) {
			passwordError = e instanceof Error ? e.message : 'Could not change your password.';
		} finally {
			savingPassword = false;
		}
	}

	async function logout() {
		await logoutSession();
		goto('/login');
	}
</script>

<div class="flex flex-col gap-10">
	{#if loadError}
		<Alert tone="danger" title="Could not load your profile">{loadError}</Alert>
	{:else if !profile}
		<Skeleton class="h-32 w-full rounded-fc-md" />
	{:else}
		<ProfileCard
			name={profile.name || profile.email}
			email={profile.email}
			avatar={profile.avatar_url || undefined}
			meta={[
				{ label: 'Member since', value: new Date(profile.created_at).toLocaleDateString() },
				{
					label: 'Avatar',
					value: profile.avatar_source === 'oidc' ? 'Managed in single sign-on' : 'Initials'
				}
			]}
		/>

		<SettingsSection title="Account" description="How you appear to everyone else in your spaces.">
			<form class="flex flex-col gap-4" onsubmit={updateProfile}>
				<div class="grid gap-4 sm:grid-cols-2">
					<Field label="Display name" helper="Shown to the other members of your spaces.">
						<Input bind:value={name} maxlength={80} placeholder="Your name" disabled={savingProfile} />
					</Field>
					<Field
						label="Email"
						helper={ssoOnly
							? 'Managed by Facile SSO — change it at porte.facile.studio.'
							: 'Used to sign in and to be added to a space.'}
					>
						<Input bind:value={email} type="email" required disabled={savingProfile || ssoOnly} />
					</Field>
				</div>

				<Button type="submit" size="lg" icon={icons.check} class="self-start" disabled={savingProfile}>
					{#if savingProfile}<Spinner size="sm" />{/if}
					{savingProfile ? 'Saving…' : 'Save changes'}
				</Button>
			</form>
		</SettingsSection>

		{#if !ssoOnly}
			<SettingsSection
				title="Password"
				description="Only for local accounts. Signing in through SSO does not use one."
			>
				<form class="flex flex-col gap-4" onsubmit={changePassword}>
					<Field label="Current password">
						<Input
							bind:value={currentPassword}
							type="password"
							autocomplete="current-password"
							required
							disabled={savingPassword}
						/>
					</Field>
					<div class="grid gap-4 sm:grid-cols-2">
						<Field label="New password" helper="At least eight characters.">
							<Input
								bind:value={newPassword}
								type="password"
								autocomplete="new-password"
								minlength={8}
								required
								disabled={savingPassword}
							/>
						</Field>
						<Field label="Confirm new password" error={mismatch}>
							<Input
								bind:value={confirmPassword}
								type="password"
								autocomplete="new-password"
								minlength={8}
								required
								disabled={savingPassword}
							/>
						</Field>
					</div>

					{#if passwordError}
						<Alert tone="danger" title="Not changed">{passwordError}</Alert>
					{/if}

					<Button
						type="submit"
						size="lg"
						icon={icons.shield}
						class="self-start"
						disabled={savingPassword || !currentPassword || !newPassword}
					>
						{#if savingPassword}<Spinner size="sm" />{/if}
						{savingPassword ? 'Changing…' : 'Change password'}
					</Button>
				</form>
			</SettingsSection>
		{/if}

		<SettingsSection title="Session" description="This device only.">
			<SettingsRow
				label="Log out"
				description="Ends this session here. Your other devices stay signed in."
			>
				<Button variant="outline" icon={icons.logout} onclick={logout}>Log out</Button>
			</SettingsRow>
		</SettingsSection>
	{/if}
</div>
