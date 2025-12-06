import type { TagsTriType } from '@/types/util';

/* Hide value filters for the toolbar. */
export const HideValue = {
	Inactive: 3,
	Skipped: 2,
	Updatable: 1,
	UpToDate: 0,
} as const;

export type HideValueType = (typeof HideValue)[keyof typeof HideValue];

export type ApprovalsToolbarOptions = {
	search: string;
	tags: TagsTriType;
	editMode: boolean;
	hide: HideValueType[];
};

/* Active service filters */
export const ACTIVE_HIDE_VALUES: HideValueType[] = [
	HideValue.UpToDate,
	HideValue.Updatable,
	HideValue.Skipped,
] as const;

/* Hide option filters for the toolbar. */
export const toolbarHideOptions = [
	{ key: 'upToDate', label: 'Hide up to date', value: HideValue.UpToDate },
	{ key: 'updatable', label: 'Hide updatable', value: HideValue.Updatable },
	{ key: 'skipped', label: 'Hide skipped', value: HideValue.Skipped },
	{ key: 'inactive', label: 'Hide inactive', value: HideValue.Inactive },
] as const;

/* Default hide value filters for the toolbar. */
export const DEFAULT_HIDE_VALUE: HideValueType[] = [
	HideValue.Inactive,
] as const;

/* Query params for the toolbar. */
export const URL_PARAMS = {
	EDIT_MODE: 'editMode',
	HIDE: 'hide',
	SEARCH: 'search',
	TAGS_EXCLUDE: 'tags_exclude',
	TAGS_INCLUDE: 'tags',
} as const;
