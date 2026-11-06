<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api, logout as logoutSession } from '$lib';
	import type { UserProfile } from '$lib';
	import { userStore } from '$lib/stores/user.svelte';
	import { formatDate } from '$lib/i18n/format';
	import { setSessionLocale, t } from '$lib/i18n/index.svelte';
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

	let ssoOnly = $state(false);

	onMount(async () => {
		try {
			profile = await api.auth.me();
			name = profile.name ?? '';
			email = profile.email;
		} catch (e) {
			loadError = e instanceof Error ? e.message : t('profile.loadFailed');
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
			setSessionLocale(profile.locale);
			toast.success(t('profile.updated'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('profile.saveFailed'));
		} finally {
			savingProfile = false;
		}
	}

	async function logout() {
		await logoutSession();
		goto('/login');
	}
</script>

<div class="flex flex-col gap-10">
	{#if loadError}
		<Alert tone="danger" title={t('profile.loadFailed')}>{loadError}</Alert>
	{:else if !profile}
		<Skeleton class="h-32 w-full rounded-fc-md" />
	{:else}
		<ProfileCard
			name={profile.name || profile.email}
			email={profile.email}
			avatar={profile.avatar_url || undefined}
			meta={[
				{ label: t('profile.memberSince'), value: formatDate(profile.created_at, { dateStyle: 'medium' }) },
				{
					label: t('profile.avatar'),
					value: profile.avatar_source === 'oidc' ? t('profile.avatarSso') : t('profile.avatarInitials')
				}
			]}
		/>

		<SettingsSection title={t('profile.accountTitle')} description={t('profile.accountText')}>
			<form class="flex flex-col gap-4" onsubmit={updateProfile}>
				<div class="grid gap-4 sm:grid-cols-2">
					<Field label={t('profile.displayName')} helper={t('profile.displayNameHelper')}>
						<Input
							bind:value={name}
							maxlength={80}
							placeholder={t('profile.namePlaceholder')}
							disabled={savingProfile}
						/>
					</Field>
					<Field
						label={t('common.email')}
						helper={ssoOnly ? t('profile.emailManaged') : t('profile.emailHelper')}
					>
						<Input bind:value={email} type="email" required disabled={savingProfile || ssoOnly} />
					</Field>
				</div>

				<Button type="submit" size="lg" icon={icons.check} class="self-start" disabled={savingProfile}>
					{#if savingProfile}<Spinner size="sm" />{/if}
					{savingProfile ? t('common.saving') : t('profile.save')}
				</Button>
			</form>
		</SettingsSection>

		<SettingsSection title={t('profile.sessionTitle')} description={t('profile.sessionText')}>
			<SettingsRow label={t('profile.logout')} description={t('profile.logoutText')}>
				<Button variant="outline" icon={icons.logout} onclick={logout}>{t('profile.logout')}</Button>
			</SettingsRow>
		</SettingsSection>
	{/if}
</div>
