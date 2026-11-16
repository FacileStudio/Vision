import type { AppKey } from '$lib/i18n/messages';
import { ApiError } from './bulle-api';

const STATUS_KEYS: Record<number, AppKey> = {
	409: 'ideas.error.busy',
	412: 'bulle.error.notLinked',
	404: 'ideas.error.gone',
	429: 'bulle.error.capped'
};

/** The message key of a failed Bulle call. */
export function bulleErrorKey(error: unknown): AppKey {
	return (error instanceof ApiError && STATUS_KEYS[error.status]) || 'bulle.error.unreachable';
}

/** The upgrade address a 402 carries, or null when the error is anything else. */
export function planRefusal(error: unknown): { upgradeUrl: string } | null {
	return error instanceof ApiError && error.status === 402 ? { upgradeUrl: error.upgradeUrl ?? '' } : null;
}
