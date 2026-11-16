import { askPrompt, bulleApi, pageContext, runActive, type BulleRun } from './bulle';
import { bulleErrorKey, planRefusal } from './bulle-errors';
import type { AppKey } from '$lib/i18n/messages';

const POLL_MS = 2000;
const GIVE_UP_MS = 4 * 60 * 1000;

/**
 * One ask-your-data run on a site page: starts it through Vision's /bulle/runs, polls it
 * every two seconds while it works, and stops polling when it ends, needs the person in
 * Bulle, takes too long, or the page goes away.
 */
export class AskRun {
	starting = $state(false);
	run = $state<BulleRun | null>(null);
	error = $state<AppKey | null>(null);
	upgradeUrl = $state<string | null>(null);

	private timer: ReturnType<typeof setTimeout> | null = null;
	private startedAt = 0;

	get busy(): boolean {
		return this.starting || (this.run !== null && this.run.state === 'running');
	}

	async start(input: { spaceId: number | null; siteId: number; title: string; question: string; locale: string; href: string }) {
		this.stop();
		this.starting = true;
		this.error = null;
		this.run = null;
		try {
			const { run_id, session_id } = await bulleApi.startRun(input.spaceId, {
				title: input.title.slice(0, 120),
				prompt: askPrompt(input.siteId, input.question, input.locale),
				page: pageContext(input.siteId, input.title, input.href)
			});
			this.run = { id: run_id, session_id, state: 'running', summary: '', continue_url: '' };
			this.startedAt = Date.now();
			await this.poll();
		} catch (error) {
			const refusal = planRefusal(error);
			if (refusal) this.upgradeUrl = refusal.upgradeUrl;
			else this.error = bulleErrorKey(error);
		} finally {
			this.starting = false;
		}
	}

	stop() {
		if (this.timer) clearTimeout(this.timer);
		this.timer = null;
	}

	private async poll() {
		if (!this.run || !runActive(this.run)) return;
		try {
			this.run = await bulleApi.run(this.run.id);
		} catch (error) {
			this.error = bulleErrorKey(error);
			return;
		}
		if (runActive(this.run) && Date.now() - this.startedAt < GIVE_UP_MS) {
			this.timer = setTimeout(() => void this.poll(), POLL_MS);
		}
	}
}
