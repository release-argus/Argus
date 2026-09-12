import type { ColumnVisibilityState, Table } from '@tanstack/react-table';
import { type ReactElement, useCallback, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router';
import { toast } from 'sonner';
import { ApprovalsToolbar } from '@/components/approvals';
import { ToolbarProvider } from '@/components/approvals/toolbar/toolbar-context';
import type { DataTableFeatures } from '@/components/ui/data-table';
import {
	APPROVALS_TOOLBAR_VIEW,
	type ApprovalsToolbarOptions,
	type CardTimestampType,
	DEFAULT_CARD_TIMESTAMPS,
	DEFAULT_HIDE_VALUE,
	DEFAULT_VIEW_VALUE,
	HideValue,
	type HideValueType,
	hideValuesFromNames,
	isToolbarViewOption,
	sortHideValues,
	type ToolbarViewOption,
	URL_PARAMS,
} from '@/constants/toolbar';
import { useAuth } from '@/contexts/auth';
import { useServices } from '@/hooks/use-services';
import { useSortableServices } from '@/hooks/use-sortable-services';
import { GridLayout } from '@/pages/approvals/layouts/grid';
import { TableLayout } from '@/pages/approvals/layouts/table/table';
import type { TagsTriType } from '@/types/util';
import type { ServiceSummary } from '@/utils/api/types/config/summary';
import {
	loadCardTimestamps,
	persistCardTimestamps,
} from '@/utils/card-timestamps';
import { visibleServices as getVisibleServices } from '@/utils/visible-services';

const toolbarFixedDefaults = {
	search: '',
	tags: { exclude: [], include: [] } as TagsTriType,
};

type DisplayDefaults = {
	hide: HideValueType[];
	timestamps: CardTimestampType[];
	view: ToolbarViewOption;
};

/**
 * @returns The 'approvals' page, including a toolbar, and a list of services.
 */
export const Approvals = (): ReactElement => {
	const [searchParams, setSearchParams] = useSearchParams();
	// Signal for resetting table sorting when order is reset.
	const [resetSortingSignal, setResetSortingSignal] = useState(0);

	const { preferences } = useAuth();
	const [defaults] = useState<DisplayDefaults>(() => {
		const savedView = preferences?.view ?? null;

		return {
			hide: sortHideValues(
				preferences?.hide
					? hideValuesFromNames(preferences.hide)
					: DEFAULT_HIDE_VALUE,
			),
			timestamps: preferences?.timestamps ?? [...DEFAULT_CARD_TIMESTAMPS],
			view: isToolbarViewOption(savedView) ? savedView : DEFAULT_VIEW_VALUE,
		};
	});

	const toolbarOptions: ApprovalsToolbarOptions = useMemo(() => {
		const search =
			searchParams.get(URL_PARAMS.SEARCH) ?? toolbarFixedDefaults.search;

		// Extract tags from URL.
		const tagsIncludeQueryParam = searchParams.get(URL_PARAMS.TAGS_INCLUDE);
		const tagsExcludeQueryParam = searchParams.get(URL_PARAMS.TAGS_EXCLUDE);
		let tags: TagsTriType;
		try {
			tags =
				tagsIncludeQueryParam || tagsExcludeQueryParam
					? {
							exclude: JSON.parse(tagsExcludeQueryParam ?? '[]') as string[],
							include: JSON.parse(tagsIncludeQueryParam ?? '[]') as string[],
						}
					: toolbarFixedDefaults.tags;
		} catch (e) {
			toast.error('Failed to parse tags from URL', {
				description: `Error: ${e instanceof Error ? e.message : String(e)}`,
			});
			tags = toolbarFixedDefaults.tags;
		}

		const editMode = searchParams.has(URL_PARAMS.EDIT_MODE);

		// Extract hide options from URL.
		const hideQueryParam = searchParams.get(URL_PARAMS.HIDE);
		let hide: HideValueType[] = [];
		if (hideQueryParam === null) {
			hide = [...defaults.hide];
		} else if (hideQueryParam) {
			try {
				const parsedHide: unknown = JSON.parse(hideQueryParam);
				if (Array.isArray(parsedHide)) {
					const validValues = Object.values(HideValue) as HideValueType[];
					hide = parsedHide
						.map(Number)
						.filter(
							(num: unknown) =>
								Number.isFinite(num) &&
								validValues.includes(num as HideValueType),
						) as HideValueType[];
				}
			} catch (e) {
				toast.error('Failed to parse hide options from URL', {
					description: `Error: ${e instanceof Error ? e.message : String(e)}`,
				});
				hide = [];
			}
		}

		const rawView = searchParams.get(URL_PARAMS.VIEW);
		const view: ApprovalsToolbarOptions['view'] = isToolbarViewOption(rawView)
			? rawView
			: defaults.view;

		return { editMode, hide, search, tags, view };
	}, [defaults, searchParams]);

	const {
		order,
		sensors,
		handleDragEnd,
		handleSaveOrder,
		hasOrderChanged,
		resetOrder,
		applyOrder,
	} = useSortableServices();

	const { services } = useServices(order);

	const visibleServices = useMemo(
		() => getVisibleServices(services, toolbarOptions),
		[services, toolbarOptions],
	);

	// URL param helpers and context value
	type URLParam = boolean | number[] | readonly number[] | string | string[];

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
					updateURLParam(
						key,
						sortHideValues(value as HideValueType[]),
						defaults.hide,
					);
					break;
				case URL_PARAMS.VIEW:
					updateURLParam(key, value, defaults.view);
					break;
				default:
					break;
			}
		},
		[defaults, updateURLParam],
	);

	// Reset service order and table sorting.
	const resetSorting = useCallback(() => {
		resetOrder();
		setResetSortingSignal((x) => x + 1);
	}, [resetOrder]);

	const toggleEditMode = useCallback(() => {
		const newValue = !toolbarOptions.editMode;
		updateURLParam(URL_PARAMS.EDIT_MODE, newValue, false);
		if (!newValue) {
			// Turning edit mode off: reset order and table sorting
			resetSorting();
		}
	}, [toolbarOptions.editMode, updateURLParam, resetSorting]);

	// Set the 'search' filter.
	const setSearch = useCallback(
		(value: string) => setValue(URL_PARAMS.SEARCH, value),
		[setValue],
	);

	// Set the 'tags' filter.
	const setTags = useCallback(
		(newTags: TagsTriType) => {
			setValue(URL_PARAMS.TAGS_INCLUDE, newTags.include);
			setValue(URL_PARAMS.TAGS_EXCLUDE, newTags.exclude);
		},
		[setValue],
	);

	// Set that layout.
	const setView = useCallback(
		(value: ToolbarViewOption) => setValue(URL_PARAMS.VIEW, value),
		[setValue],
	);

	// Set the 'hide' filter options.
	const setHide = useCallback(
		(value: number[]) => setValue(URL_PARAMS.HIDE, value),
		[setValue],
	);

	// Table instance for the 'table' view.
	const [tableInstance, setTableInstance] = useState<
		Table<DataTableFeatures, ServiceSummary> | undefined
	>(undefined);

	// Column order for the 'table' view.
	const [tableColumnOrder, setTableColumnOrder] = useState<string[]>([]);
	// Column visibility for the 'table' view.
	const [tableColumnVisibility, setTableColumnVisibility] =
		useState<ColumnVisibilityState>({});

	// Timestamps shown at the bottom of the service card..
	const [cardTimestamps, setCardTimestamps] = useState<CardTimestampType[]>(
		() => loadCardTimestamps(defaults.timestamps),
	);
	const toggleCardTimestamp = useCallback((value: CardTimestampType) => {
		setCardTimestamps((current) => {
			const next = current.includes(value)
				? current.filter((timestamp) => timestamp !== value)
				: [...current, value];
			persistCardTimestamps(next);
			return next;
		});
	}, []);

	return (
		<ToolbarProvider
			value={{
				cardTimestamps,
				defaultHide: defaults.hide,
				hasOrderChanged,
				onSaveOrder: handleSaveOrder,
				setHide,
				setSearch,
				setTableColumnOrder,
				setTableColumnVisibility,
				setTableInstance,
				setTags,
				setView,
				tableColumnOrder,
				tableColumnVisibility,

				tableInstance,
				toggleCardTimestamp,
				toggleEditMode,

				values: toolbarOptions,
				visibleServices,
			}}
		>
			<ApprovalsToolbar />
			{toolbarOptions.view === APPROVALS_TOOLBAR_VIEW.GRID.value && (
				<GridLayout
					editMode={toolbarOptions.editMode}
					handleDragEnd={handleDragEnd}
					order={order}
					sensors={sensors}
					services={visibleServices}
				/>
			)}
			{toolbarOptions.view === APPROVALS_TOOLBAR_VIEW.TABLE.value && (
				<TableLayout
					applyOrder={applyOrder}
					editMode={toolbarOptions.editMode}
					handleDragEnd={handleDragEnd}
					order={order}
					resetSorting={resetSorting}
					resetSortingSignal={resetSortingSignal}
					sensors={sensors}
					services={visibleServices}
				/>
			)}
		</ToolbarProvider>
	);
};
