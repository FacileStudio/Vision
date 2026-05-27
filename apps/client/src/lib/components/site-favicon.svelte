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
</script>

<div class="relative shrink-0 rounded overflow-hidden bg-muted {className}">
	<span class="absolute inset-0 flex items-center justify-center text-[10px] font-medium text-muted-foreground select-none">
		{initial}
	</span>
	{#if !errored}
		<img
			src="https://www.google.com/s2/favicons?domain={domain}&sz=32"
			alt=""
			class="relative h-full w-full transition-opacity duration-150 {loaded ? 'opacity-100' : 'opacity-0'}"
			onload={() => (loaded = true)}
			onerror={() => (errored = true)}
		/>
	{/if}
</div>
