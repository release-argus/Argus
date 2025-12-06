import { type FC, memo, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import EditModeToggle from '@/components/approvals/toolbar/edit-mode-toggle';
import FilterDropdown from '@/components/approvals/toolbar/filter-dropdown';
import SearchBar from '@/components/approvals/toolbar/search-bar';
import TagSelect from '@/components/approvals/toolbar/tag-select';
import { ToolbarProvider } from '@/components/approvals/toolbar/toolbar-context';
import { ButtonGroup } from '@/components/ui/button-group';
import { TooltipProvider } from '@/components/ui/tooltip';
import {
	type ApprovalsToolbarOptions,
	DEFAULT_HIDE_VALUE,
	URL_PARAMS,
} from '@/constants/toolbar';
import type { TagsTriType } from '@/types/util';

type ApprovalsToolbarProps = {
	/* Toolbar values from query params */
	values: ApprovalsToolbarOptions;

	/* Callback for edit mode toggle */
	onEditModeToggle: (value: boolean) => void;
	/* Callback for saving the current ordering */
	onSaveOrder: () => void;
	/* Whether the service ordering has changed */
	hasOrderChanged: boolean;
};

type URLParam = boolean | number[] | readonly number[] | string | string[];

/**
 * ApprovalsToolbar
 *
 * Toolbar for the 'approvals' view.
 * Manages search, tag filters, edit mode, hide settings, and service order state via URL parameters.
 */

const ApprovalsToolbar: FC<ApprovalsToolbarProps> = ({
	values,

	onEditModeToggle,
	onSaveOrder,
	hasOrderChanged,
}) => {
	const [, setSearchParams] = useSearchParams();

	// Add/remove a query param.
	const updateURLParam = useCallback(
		(
			key: (typeof URL_PARAMS)[keyof typeof URL_PARAMS],
			value: URLParam,
			defaultValue: URLParam,
		) => {
			const newSearchParams = new URLSearchParams(globalThis.location.search);

			if (Array.isArray(value)) {
				if (JSON.stringify(value) === JSON.stringify(defaultValue)) {
					newSearchParams.delete(key);
				} else {
					newSearchParams.set(key, JSON.stringify(value));
				}
			} else if (value === defaultValue) {
				newSearchParams.delete(key);
			} else {
				newSearchParams.set(key, value.toString());
			}

			setSearchParams(newSearchParams);
		},
		[setSearchParams],
	);

	// Set a query param.
	const setValue = useCallback(
		(
			key: (typeof URL_PARAMS)[keyof typeof URL_PARAMS],
			value: string | boolean | number[] | readonly number[] | string[],
		) => {
			switch (key) {
				case URL_PARAMS.SEARCH:
					updateURLParam(key, value, '');
					break;
				case URL_PARAMS.TAGS_INCLUDE:
				case URL_PARAMS.TAGS_EXCLUDE:
					updateURLParam(key, value, []);
					break;
				case URL_PARAMS.EDIT_MODE:
					updateURLParam(key, value, false);
					break;
				case URL_PARAMS.HIDE:
					updateURLParam(key, value, DEFAULT_HIDE_VALUE);
					break;
			}
		},
		[updateURLParam],
	);

	// Edit mode.
	const toggleEditMode = () => {
		const newValue = !values.editMode;
		updateURLParam(URL_PARAMS.EDIT_MODE, newValue, false);
		onEditModeToggle(newValue);
	};
	// Service search.
	const setSearch = (value: string) => setValue(URL_PARAMS.SEARCH, value);
	// Tag filtering.
	const setTags = (newTags: TagsTriType) => {
		setValue(URL_PARAMS.TAGS_INCLUDE, newTags.include);
		setValue(URL_PARAMS.TAGS_EXCLUDE, newTags.exclude);
	};
	// 'Hide' options.
	const setHide = (value: number[]) => setValue(URL_PARAMS.HIDE, value);

	return (
		<div className="mb-3 flex gap-2 md:gap-3">
			<TooltipProvider>
				<ToolbarProvider
					value={{
						hasOrderChanged,
						onSaveOrder,
						setHide,
						setSearch,
						setTags,
						toggleEditMode,
						values,
					}}
				>
					<SearchBar />
					<TagSelect />
					<ButtonGroup>
						<FilterDropdown />
						<EditModeToggle />
					</ButtonGroup>
				</ToolbarProvider>
			</TooltipProvider>
		</div>
	);
};

export default memo(ApprovalsToolbar);
