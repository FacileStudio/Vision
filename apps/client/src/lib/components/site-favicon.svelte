<script lang="ts">
	import { twMerge } from '@facile/muse';

	let {
		domain,
		name = '',
		class: className = ''
	}: {
		domain: string;
		name?: string;
		class?: string;
	} = $props();

	let loaded = $state(false);
	let errored = $state(false);

	let initial = $derived((name || domain).charAt(0).toUpperCase());

	/*
	 * A site's favicon is fetched from the site itself, so it fails often — a 404, a
	 * captive portal, a domain that is not up yet. The initial underneath is not a
	 * placeholder, it is what most rows actually render.
	 */
	function handleLoad(e: Event) {
		const img = e.currentTarget as HTMLImageElement;
		if (img.naturalWidth > 0) loaded = true;
		else errored = true;
	}
</script>

<span
	class={twMerge(
		'relative inline-flex shrink-0 items-center justify-center overflow-hidden rounded-fc-xs bg-fc-surface text-fc-xs font-semibold text-fc-fg-muted',
		className
	)}
>
	<span aria-hidden="true">{initial}</span>
	{#if !errored}
		<img
			src="https://{domain}/favicon.ico"
			alt=""
			class="absolute inset-0 h-full w-full object-cover transition-opacity duration-150 {loaded
				? 'opacity-100'
				: 'opacity-0'}"
			onload={handleLoad}
			onerror={() => (errored = true)}
		/>
	{/if}
</span>
