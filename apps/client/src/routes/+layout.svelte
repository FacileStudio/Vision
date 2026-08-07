<script lang="ts">
	import '../app.css';
	import { browser } from '$app/environment';
	import { Toaster } from '@facile/muse';
	import { setTheme, theme } from '$lib/theme.svelte';

	let { children } = $props();

	/*
	 * muse renders <iconify-icon> elements; the custom element has to be registered by the
	 * consumer. Imported dynamically because this layout is server-rendered and the package
	 * calls customElements.define at module scope.
	 */
	if (browser) {
		void import('iconify-icon');
		setTheme(theme.mode);
	}
</script>

{@render children()}

<!-- One Toaster for the whole app, outside the router, so a navigation cannot unmount a
     toast mid-flight. The extra bottom padding keeps the stack clear of MobileNav. -->
<Toaster class="pb-28 md:pb-6" />
