import type { ApprovalRequest, Idea, IdeaRun } from '@facile/muse';
import type { AppKey } from '$lib/i18n/messages';
import type { BulleIdea, IdeaRunResponse } from './bulle';
import { safeHref } from './bulle';

type Translate = (key: AppKey, params?: Record<string, string | number>) => string;

/** What the page knows about one idea's run: its id, the last poll, or why it stopped. */
export type RunEntry = { runId: string | null; view: IdeaRunResponse | null; error: AppKey | null };

/** A readable name for a Bulle tool: the prefix and underscores go, the first letter is capitalised. */
export function toolLabel(tool: string): string {
	const words = tool.replace(/^[a-z]+_/, '').replaceAll('_', ' ').trim();
	return words ? words[0]!.toUpperCase() + words.slice(1) : tool;
}

function approvalOf(view: IdeaRunResponse, t: Translate): ApprovalRequest | undefined {
	const first = view.pending_approvals?.[0];
	if (!first) return undefined;
	const tool = toolLabel(first.tool);
	const args = first.args && typeof first.args === 'object' ? (first.args as Record<string, unknown>) : undefined;
	return { id: first.call_id, tool, summary: t('ideas.approval.summary', { tool }), args };
}

/** The run muse draws for an entry: failed with a reason, running until the first poll, then Bulle's state. */
export function runOf(entry: RunEntry, t: Translate): IdeaRun {
	if (entry.error) return { state: 'failed', error: t(entry.error) };
	if (!entry.view) return { state: 'running' };
	const { view } = entry;
	if (view.state === 'failed') return { state: 'failed', error: view.summary || t('ideas.error.run') };
	const outputs = (view.outputs ?? []).flatMap((output) => {
		const href = safeHref(output.url);
		return href ? [{ label: output.preview || output.tool, href }] : [];
	});
	return { state: view.state, summary: view.summary || undefined, outputs, approval: approvalOf(view, t) };
}

/** True while the entry still needs polling. */
export function isActive(entry: RunEntry | undefined): boolean {
	if (!entry || entry.error) return false;
	return !entry.view || entry.view.state === 'running' || entry.view.state === 'needs_approval';
}

/** One card: Bulle's idea plus its run in this session, with the continue-in-Bulle link once it ran. */
export function toIdea(view: BulleIdea, entry: RunEntry | undefined, t: Translate): Idea {
	return {
		id: view.id,
		kind: view.kind,
		title: view.title,
		why: view.why,
		draft: view.draft || undefined,
		sources: (view.sources ?? []).map((source) => ({ label: source.label, href: safeHref(source.ref) })),
		status: view.status as Idea['status'],
		expiresAt: view.expires_at || undefined,
		run: entry ? runOf(entry, t) : undefined,
		continueHref: safeHref(entry?.view?.continue_url)
	};
}
