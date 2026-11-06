import { locale } from './index.svelte';
import { countryName, fmt as formatNumber, formatDateRange, formatHour } from '../utils/analytics';

/** Formats a date or instant with `Intl.DateTimeFormat` in the current locale. */
export function formatDate(value: Date | string | number, options: Intl.DateTimeFormatOptions = {}): string {
	return new Intl.DateTimeFormat(locale(), options).format(new Date(value));
}

/** An integer with the current locale's grouping. */
export const fmt = (n: number): string => formatNumber(n, locale());

/** "Jan 5 – Feb 4, 2026" in the current locale. */
export const dateRange = (from: string, to: string): string => formatDateRange(from, to, locale());

/** An hour of the day, 0 to 23, in the current locale's clock. */
export const hourLabel = (hour: number): string => formatHour(hour, locale());

/** A region code as a country name in the current locale. */
export const country = (code: string): string => countryName(code, locale());
