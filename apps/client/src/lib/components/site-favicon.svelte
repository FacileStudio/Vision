<script lang="ts">
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

	function colorFromDomain(d: string): string {
		let hash = 0;
		for (let i = 0; i < d.length; i++) {
			hash = d.charCodeAt(i) + ((hash << 5) - hash);
		}
		return `hsl(${Math.abs(hash % 360)} 45% 45%)`;
	}
</script>

<div class="relative shrink-0 rounded overflow-hidden {className}" style="background: {colorFromDomain(domain)}">
	{#if !loaded}
		<span class="absolute inset-0 flex items-center justify-center text-[10px] font-bold text-white select-none">
			{initial}
		</span>
	{/if}
	{#if !errored}
		<img
			src="https://{domain}/favicon.ico"
			alt=""
			class="absolute inset-0 h-full w-full object-contain transition-opacity duration-150 {loaded ? 'opacity-100' : 'opacity-0'}"
			onload={() => (loaded = true)}
			onerror={() => (errored = true)}
		/>
	{/if}
</div>
