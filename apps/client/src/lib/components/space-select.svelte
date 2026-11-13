<script lang="ts">
	import { Select } from '@facile/muse';
	import { workspaceStore } from '$lib/stores/workspace.svelte';
	import { t } from '$lib/i18n/index.svelte';

	let {
		value = $bindable(''),
		preferred = '',
		disabled = false
	}: { value?: string; preferred?: string; disabled?: boolean } = $props();

	const options = $derived(workspaceStore.writable.map((w) => ({ value: String(w.id), label: w.name })));

	$effect(() => {
		if (options.some((o) => o.value === value)) return;
		value = options.find((o) => o.value === preferred)?.value ?? options[0]?.value ?? '';
	});
</script>

<Select wrapperClass="w-full sm:w-56" aria-label={t('spaceFilter.label')} {options} bind:value {disabled} />
