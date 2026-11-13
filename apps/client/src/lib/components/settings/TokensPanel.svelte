<script lang="ts">
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { APIKeyItem } from '$lib';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import SpaceSelect from '$lib/components/space-select.svelte';
	import { formatDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';
	import { SecretField, SettingsSection, TokensSection, toast, type TokenDraft, type TokenItem } from '@facile/muse';

	let keys = $state<APIKeyItem[]>([]);
	let spaceId = $state('');

	const endpoint = $derived(`${page.url.origin}/api`);

	const scopes = $derived([
		{ value: 'read', label: t('apiKeys.scopeReadOption') },
		{ value: 'read,write', label: t('apiKeys.scopeWriteOption') }
	]);

	const items = $derived<TokenItem[]>(
		keys.map((key) => ({
			id: String(key.id),
			name: key.name,
			prefix: `${key.prefix}_****${key.key_hint}`,
			scope: key.scopes.includes('write') ? t('apiKeys.scopeWrite') : t('apiKeys.scopeRead'),
			lastUsed: key.last_used_at ? formatDate(key.last_used_at, { dateStyle: 'medium' }) : undefined,
			revoked: !key.is_active
		}))
	);

	async function load() {
		try {
			keys = await api.apiKeys.list(Number(spaceId) || undefined);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('apiKeys.loadFailed'));
		}
	}

	$effect(() => {
		if (spaceId || workspaceStore.writable.length === 0) load();
	});

	async function create(draft: Omit<TokenDraft, 'days'>): Promise<string> {
		const resp = await api.apiKeys.create({
			name: draft.name,
			scopes: draft.scope ?? 'read',
			workspace_id: Number(spaceId) || undefined
		});
		await load();
		return resp.key;
	}

	async function revoke(target: TokenItem) {
		await api.apiKeys.revoke(Number(target.id));
		toast.neutral(t('apiKeys.revoked'));
		await load();
	}
</script>

<div class="flex flex-col gap-10">
	<div class={workspaceStore.writable.length > 1 ? 'contents' : 'hidden'}>
		<SettingsSection title={t('settings.space.title')} description={t('settings.space.text')}>
			<SpaceSelect bind:value={spaceId} />
		</SettingsSection>
	</div>
	<TokensSection tokens={items} {scopes} expiry={false} onCreate={create} onRevoke={revoke}>
		<SettingsSection title={t('apiKeys.endpointTitle')} description={t('apiKeys.endpointText')}>
			<SecretField value={endpoint} sensitive={false} label={t('apiKeys.baseUrl')} />
		</SettingsSection>
	</TokensSection>
</div>
