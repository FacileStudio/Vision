<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { APIKeyItem } from '$lib';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import { formatDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';
	import { SecretField, SettingsSection, TokensSection, toast, type TokenDraft, type TokenItem } from '@facile/muse';

	let keys = $state<APIKeyItem[]>([]);

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
			keys = await api.apiKeys.list(workspaceStore.current?.id);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('apiKeys.loadFailed'));
		}
	}

	onMount(load);

	$effect(() => {
		workspaceStore.current;
		load();
	});

	async function create(draft: Omit<TokenDraft, 'days'>): Promise<string> {
		const resp = await api.apiKeys.create({
			name: draft.name,
			scopes: draft.scope ?? 'read',
			workspace_id: workspaceStore.current?.id
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

<TokensSection tokens={items} {scopes} expiry={false} onCreate={create} onRevoke={revoke}>
	<SettingsSection title={t('apiKeys.endpointTitle')} description={t('apiKeys.endpointText')}>
		<SecretField value={endpoint} sensitive={false} label={t('apiKeys.baseUrl')} />
	</SettingsSection>
</TokensSection>
