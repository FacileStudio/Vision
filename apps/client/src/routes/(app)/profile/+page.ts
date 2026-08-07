import { redirect } from '@sveltejs/kit';

/*
 * Profile moved into Settings › Profile when the settings page was split into routed
 * sections. The old path is kept as a redirect because it was the sidebar's user-card
 * link for a year and is bookmarked.
 */
export function load() {
	redirect(308, '/settings');
}
