import type { ApprovalsToolbarOptions, ToolbarViewOption, } from '@/constants/toolbar';
import type { TagsTriType } from '@/types/util';
import type { ServiceSummary } from '@/utils/api/types/config/summary';
import type { Table, VisibilityState } from '@tanstack/react-table';
import { createContext, type Dispatch, type SetStateAction, useContext, } from 'react';

export type ToolbarContextValue = {
	/* Current values reflected from URL/search params */
	values: ApprovalsToolbarOptions;

	/* Sets the 'search' value in the URL */
	setSearch: (value: string) => void;
	/* Sets the 'tags' value in the URL */
	setTags: (value: TagsTriType) => void;
	/* Sets the 'hide' value in the URL */
	setHide: (value: number[]) => void;
	/* Sets the 'view' value in the URL */
	setView: (value: ToolbarViewOption) => void;
	/* Toggles the 'editMode' value in the URL */
	toggleEditMode: () => void;

	/* Tanstack table instance */
	tableInstance?: Table<ServiceSummary>;
	/* Function to set the table instance */
	setTableInstance: (value: Table<ServiceSummary> | undefined) => void;
	/* Order of columns in the table */
	tableColumnOrder: string[];
	/* Sets the order of columns in the table */
	setTableColumnOrder: Dispatch<SetStateAction<string[]>>;
	/* Visible columns in the table */
	tableColumnVisibility: VisibilityState;
	/* Sets the visibility of a column in the table */
	setTableColumnVisibility: Dispatch<SetStateAction<VisibilityState>>;

	/* Saves the current order of the service list */
	onSaveOrder: () => Promise<void>;
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
