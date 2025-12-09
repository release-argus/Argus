import { LayoutGrid, List } from 'lucide-react';
import type { TagsTriType } from '@/types/util';

/* Hide value filters for the toolbar. */
export const HideValue = {
	Inactive: 3,
	Skipped: 2,
	Updatable: 1,
	UpToDate: 0,
} as const;

export type HideValueType = (typeof HideValue)[keyof typeof HideValue];

/* View options for the toolbar. */
export const APPROVALS_TOOLBAR_VIEW = {
	GRID: { icon: LayoutGrid, label: 'Grid', value: 'grid' },
	TABLE: { icon: List, label: 'Table', value: 'table' },
} as const;
export const approvalsToolbarViewOptions = Object.values(
	APPROVALS_TOOLBAR_VIEW,
);
export type ToolbarViewOption =
	(typeof approvalsToolbarViewOptions)[number]['value'];

export const DEFAULT_VIEW_VALUE = APPROVALS_TOOLBAR_VIEW.GRID.value;

export const isToolbarViewOption = (
	option: string | null,
): option is ToolbarViewOption =>
	approvalsToolbarViewOptions.some((o) => o.value === option);

export type ApprovalsToolbarOptions = {
	search: string;
	tags: TagsTriType;
	editMode: boolean;
	hide: HideValueType[];
	view: ToolbarViewOption;
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
	VIEW: 'view',
} as const;

/* Storage key for the visible table columns. */
export const TABLE_COLUMNS_VISIBLE_STORAGE_KEY = 'tableColumnsVisible';
/* Storage key for the order of table columns. */
export const TABLE_COLUMNS_ORDER_STORAGE_KEY = 'tableColumnsOrder';
