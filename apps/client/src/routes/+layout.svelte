<script lang="ts">
	import '../app.css';
	import { browser } from '$app/environment';
	import { Toaster, provideLocale, registerIcons } from '@facile/muse';
	import { museIconData } from '$lib/icons/register';
	import mdi from '$lib/icons/mdi.json';
	import solar from '$lib/icons/solar.json';
	import { forgetStoredToken } from '$lib';
	import { theme } from '$lib/theme.svelte';
	import { bindLocale, localeSource } from '$lib/i18n/index.svelte';

	let { children } = $props();

	registerIcons({ ...museIconData(solar), ...museIconData(mdi) });

	const i18n = provideLocale(localeSource);
	bindLocale(() => i18n.locale);

	/*
	 * muse renders <iconify-icon> elements; the custom element has to be registered by the
	 * consumer. Imported dynamically so the package's module-scope customElements.define
	 * never runs outside a browser, such as under svelte-check or bun test.
	 */
	if (browser) {
		void import('iconify-icon');
		theme.restore();
		forgetStoredToken();
	}
</script>

{@render children()}

<!-- One Toaster for the whole app, outside the router, so a navigation cannot unmount a
     toast mid-flight. The extra bottom padding keeps the stack clear of MobileNav. -->
<Toaster class="pb-28 md:pb-6" />
