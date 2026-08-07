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
	 * A site's favicon is fetched from the site itself — never a third-party favicon service,
	 * which would hand every tracked domain to someone else and is the opposite of what this
	 * product sells. So it fails often: a 404, a captive portal, a domain that is not up yet.
	 * The initial underneath is not a placeholder, it is what most rows actually render.
	 *
	 * `load` firing is not proof of a usable image. A server that answers /favicon.ico with an
	 * HTML error page, or with a truncated or malformed .ico, can still reach `load` — and the
	 * element then paints the browser's broken-image glyph on top of the initial, which is
	 * what was showing in the sites list. `decode()` is the check that actually rejects those,
	 * and the 8px floor rejects the 1×1 tracking pixel some hosts return instead of a 404.
	 */
	async function handleLoad(e: Event) {
		const img = e.currentTarget as HTMLImageElement;
		try {
			await img.decode();
		} catch {
			errored = true;
			return;
		}
		if (img.naturalWidth >= 8) loaded = true;
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
		<!-- `hidden` until it has decoded, not just transparent: an <img> that failed still
		     paints its broken glyph, and opacity-0 hides the pixels while leaving the box
		     laid out over the initial. `no-referrer` keeps Vision's own URLs out of the
		     tracked site's logs — the request is unavoidable, telling them who sent it is not. -->
		<img
			src="https://{domain}/favicon.ico"
			alt=""
			referrerpolicy="no-referrer"
			loading="lazy"
			class="absolute inset-0 h-full w-full object-cover transition-opacity duration-150 {loaded
				? 'opacity-100'
				: 'invisible opacity-0'}"
			onload={handleLoad}
			onerror={() => (errored = true)}
		/>
	{/if}
</span>
