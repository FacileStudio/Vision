import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export type WithoutChildren<T> = T extends infer U ? Omit<U, "children"> : never;

export type WithoutChildrenOrChild<T> = T extends infer U ? Omit<U, "children" | "child"> : never;

export type WithElementRef<T, El extends HTMLElement = HTMLElement> = T & { ref?: El | null };
