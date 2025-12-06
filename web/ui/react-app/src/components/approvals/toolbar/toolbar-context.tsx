import { createContext, useContext } from 'react';
import type { ApprovalsToolbarOptions } from '@/constants/toolbar';
import type { TagsTriType } from '@/types/util';

export type ToolbarContextValue = {
	/* Current values reflected from URL/search params */
	values: ApprovalsToolbarOptions;

	/* Sets the 'search' value in the URL */
	setSearch: (value: string) => void;
	/* Sets the 'tags' value in the URL */
	setTags: (value: TagsTriType) => void;
	/* Sets the 'hide' value in the URL */
	setHide: (value: number[]) => void;
	/* Toggles the 'editMode' value in the URL */
	toggleEditMode: () => void;

	/* Saves the current order of the service list */
	onSaveOrder: () => void;
	/* Whether the order of the service list has changed */
	hasOrderChanged: boolean;
};

const ToolbarContext = createContext<ToolbarContextValue | null>(null);

export const ToolbarProvider = ToolbarContext.Provider;

export const useToolbar = (): ToolbarContextValue => {
	const ctx = useContext(ToolbarContext);
	if (!ctx) {
		throw new Error('useToolbar must be used within a ToolbarProvider');
	}
	return ctx;
};
