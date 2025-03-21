export enum HideValue {
	UpToDate = 0,
	Updatable = 1,
	Skipped = 2,
	Inactive = 3,
}

export const DEFAULT_HIDE_VALUE = [HideValue.Inactive];

export const URL_PARAMS = {
	SEARCH: 'search',
	TAGS: 'tags',
	EDIT_MODE: 'editMode',
	HIDE: 'hide',
} as const;
