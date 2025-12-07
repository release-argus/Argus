import { useQueryClient } from '@tanstack/react-query';
import type { Table, VisibilityState } from '@tanstack/react-table';
import { type ReactElement, useCallback, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import { ApprovalsToolbar } from '@/components/approvals';
import { ToolbarProvider } from '@/components/approvals/toolbar/toolbar-context';
import {
	APPROVALS_TOOLBAR_VIEW,
	type ApprovalsToolbarOptions,
	DEFAULT_HIDE_VALUE,
	DEFAULT_VIEW_VALUE,
	HideValue,
	type HideValueType,
	isToolbarViewOption,
	type ToolbarViewOption,
	URL_PARAMS,
} from '@/constants/toolbar';
import { useServices } from '@/hooks/use-services';
import { useSortableServices } from '@/hooks/use-sortable-services';
import { GridLayout } from '@/pages/approvals/layouts/grid';
import { TableLayout } from '@/pages/approvals/layouts/table/table';
import type { TagsTriType } from '@/types/util';
import type { ServiceSummary } from '@/utils/api/types/config/summary';

const toolbarDefaults: ApprovalsToolbarOptions = {
	editMode: false,
	hide: DEFAULT_HIDE_VALUE,
	search: '',
	tags: { exclude: [], include: [] },
	view: DEFAULT_VIEW_VALUE,
};

/**
 * @returns The 'approvals' page, including a toolbar, and a list of services.
 */
export const Approvals = (): ReactElement => {
	const queryClient = useQueryClient();
	const [searchParams, setSearchParams] = useSearchParams();
	// Signal for resetting table sorting when order is reset.
	const [resetSortingSignal, setResetSortingSignal] = useState(0);

	const toolbarOptions: ApprovalsToolbarOptions = useMemo(() => {
		const search =
			searchParams.get(URL_PARAMS.SEARCH) ?? toolbarDefaults.search;

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
					: toolbarDefaults.tags;
		} catch (e) {
			toast.error('Failed to parse tags from URL', {
				description: `Error: ${e instanceof Error ? e.message : String(e)}`,
			});
			tags = toolbarDefaults.tags;
		}

		const editMode = searchParams.has(URL_PARAMS.EDIT_MODE);

		// Extract hide options from URL.
		const hideQueryParam = searchParams.get(URL_PARAMS.HIDE);
		let hide: HideValueType[] = [];
		if (hideQueryParam === null) {
			hide = [...toolbarDefaults.hide];
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
			: toolbarDefaults.view;

		return { editMode, hide, search, tags, view };
	}, [searchParams]);

	const {
		order,
		sensors,
		handleDragEnd,
		handleSaveOrder,
		hasOrderChanged,
		resetOrder,
		applyOrder,
	} = useSortableServices();

	const services = useServices(order);

	// Filter the services based on the toolbar options.
	// biome-ignore lint/correctness/useExhaustiveDependencies: queryClient stable.
	const filteredServices = useMemo(() => {
		const {
			search = '',
			tags = toolbarDefaults.tags,
			hide,
		} = {
			...toolbarOptions,
			search: toolbarOptions.search.toLowerCase(),
		};
		const filterOnTags = tags.include.length > 0 || tags.exclude.length > 0;
		const excludeOnly = filterOnTags && tags.include.length === 0;

		return services
			.filter((svc) => {
				const service = svc.data;
				if (!service || service.loading) return true;

				const serviceID = service.id;

				// Filter on 'tags'.
				//     Have no tags to filter on,
				//   OR
				//     The service doesn't have any EXCLUDE tags
				//       AND
				//     We are only excluding tags, OR the service has all INCLUDE tags.
				const hasTags =
					!filterOnTags ||
					(!tags.exclude.some((tag) => service.tags?.includes(tag)) &&
						(excludeOnly ||
							tags.include.some((tag) => service.tags?.includes(tag))));
				if (!hasTags) return false;

				// Filter on 'name'.
				const name = (service.name ?? serviceID).toLowerCase();
				if (!name.includes(search)) return false;

				// Filter on 'hide' options.
				const skipped = service.status?.state === 'SKIPPED';
				const upToDate = service.status?.state === 'UP_TO_DATE';
				const hideInactiveServices = hide.includes(HideValue.Inactive);
				return (
					// hideUpToDate: deployed_version NOT latest_version.
					(!hide.includes(HideValue.UpToDate) || !upToDate) &&
					// hideUpdatable: deployed_version IS latest_version AND approved_version NOT a skip of latest_version.
					(!hide.includes(HideValue.Updatable) || upToDate || skipped) &&
					// hideSkipped: approved_version NOT a skip of latest_version OR NO approved_version.
					(!hide.includes(HideValue.Skipped) || !skipped) &&
					// hideInactive: active NOT false.
					(!hideInactiveServices || service.active !== false)
				);
			})
			.map((svc) => svc.data)
			.filter(Boolean) as ServiceSummary[];
	}, [toolbarOptions, queryClient, services]);

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
					updateURLParam(key, value, DEFAULT_HIDE_VALUE);
					break;
				case URL_PARAMS.VIEW:
					updateURLParam(key, value, DEFAULT_VIEW_VALUE);
					break;
				default:
					break;
			}
		},
		[updateURLParam],
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
		Table<ServiceSummary> | undefined
	>(undefined);

	// Column order for the 'table' view.
	const [tableColumnOrder, setTableColumnOrder] = useState<string[]>([]);
	// Column visibility for the 'table' view.
	const [tableColumnVisibility, setTableColumnVisibility] =
		useState<VisibilityState>({});

	return (
		<ToolbarProvider
			value={{
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
				toggleEditMode,

				values: toolbarOptions,
			}}
		>
			<ApprovalsToolbar />
			{toolbarOptions.view === APPROVALS_TOOLBAR_VIEW.GRID.value && (
				<GridLayout
					editMode={toolbarOptions.editMode}
					handleDragEnd={handleDragEnd}
					order={order}
					sensors={sensors}
					services={filteredServices}
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
					services={filteredServices}
				/>
			)}
		</ToolbarProvider>
	);
};
