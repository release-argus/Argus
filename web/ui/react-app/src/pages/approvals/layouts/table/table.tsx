import type { DragEndEvent, SensorOptions } from '@dnd-kit/core';
import type { SensorDescriptor } from '@dnd-kit/core/dist/sensors/types';
import type { ColumnVisibilityState, Row } from '@tanstack/react-table';
import { type FC, useCallback, useEffect } from 'react';
import { useToolbar } from '@/components/approvals/toolbar/toolbar-context';
import { DataTable, type DataTableFeatures } from '@/components/ui/data-table';
import { TABLE_COLUMNS_ORDER_STORAGE_KEY } from '@/constants/toolbar';
import { mergeColumnOrder } from '@/pages/approvals/layouts/table/column-order';
import {
	loadColumnVisibility,
	persistColumnVisibility,
	resetColumnVisibility,
	setAutoHideColumnVisibility,
} from '@/pages/approvals/layouts/table/column-visibility';
import { columns } from '@/pages/approvals/layouts/table/columns';
import type { ServiceSummary } from '@/utils/api/types/config/summary';

type TableLayoutProps = {
	/* The list of service summaries for the table. */
	services: ServiceSummary[];
	/* A flag indicating whether the layout is in edit mode. */
	editMode: boolean;
	/* The current order of rows. */
	order: string[];
	/* An array of sensor descriptors for drag-and-drop operations. */
	sensors: SensorDescriptor<SensorOptions>[];
	/* A function invoked after the drag-and-drop interaction on services is completed. */
	handleDragEnd: (event: DragEndEvent) => void;
	/* Apply the given order (without saving). */
	applyOrder: (ids: string[]) => void;
	/* Changes to this trigger the 'sorting reset' signal, resetting the table sorting. */
	resetSortingSignal?: number;
	/* Function to reset the table sorting (given to headers). */
	resetSorting?: () => void;
};

/**
 * A functional React component for rendering a table layout for services with support for drag-and-drop re-ordering,
 * column sorting, column visibility, and column ordering.
 *
 * @type {FC<TableLayoutProps>} TableLayout
 *
 * @param services - The list of service summaries for the table.
 * @param editMode - A flag indicating whether the layout is in edit mode.
 * @param order - An array of sensor descriptors for drag-and-drop operations.
 * @param sensors - A function invoked after the drag-and-drop interaction on services is completed.
 * @param handleDragEnd - Apply the given order (without saving).
 * @param applyOrder - When this numeric signal changes, the table sorting will be reset.
 * @param resetSortingSignal - Changes to this trigger the 'sorting reset' signal, resetting the table sorting.
 * @param resetSorting - Function to reset the table sorting (given to headers).
 * @returns The rendered table containing the given services in the order specified with drag-and-drop capabilities.
 */
export const TableLayout: FC<TableLayoutProps> = ({
	services,
	editMode,
	order,
	sensors,
	handleDragEnd,
	applyOrder,
	resetSortingSignal,
	resetSorting,
}) => {
	const {
		setTableInstance,
		setTableColumnVisibility,
		tableColumnVisibility,
		tableColumnOrder,
		setTableColumnOrder,
	} = useToolbar();
	const getItemId = useCallback((row: ServiceSummary) => row.id, []);

	// Inactive service colouring.
	const getRowClassName = useCallback(
		(row: Row<DataTableFeatures, ServiceSummary>) =>
			row.original?.active === false
				? 'border-2! border-[var(--muted-foreground)] italic line-through'
				: undefined,
		[],
	);

	// Merge a sorted subset of services (visible rows) into the full order.
	// biome-ignore lint/correctness/useExhaustiveDependencies: applyOrder stable.
	const onSortedOrderChange = useCallback(
		(sortedVisibleIDs: string[]) => {
			if (!applyOrder) return;
			if (!Array.isArray(sortedVisibleIDs) || sortedVisibleIDs.length === 0)
				return;

			// Remove any IDs present in the sorted subset from the current full order.
			const visibleSet = new Set(sortedVisibleIDs);
			const remaining = order.filter((id) => !visibleSet.has(id));
			const nextOrder = [...sortedVisibleIDs, ...remaining];

			applyOrder(nextOrder);
		},
		[order],
	);

	// Sets column visibility in the table, and persists to localStorage.
	// Resets visibility when all columns would be hidden.
	// biome-ignore lint/correctness/useExhaustiveDependencies: setColumnVisibility stable.
	const setColumnVisibilityMaster = useCallback(
		(
			updater:
				| ColumnVisibilityState
				| ((prev: ColumnVisibilityState) => ColumnVisibilityState),
		) => {
			setTableColumnVisibility((prev) => {
				const newValue =
					typeof updater === 'function' ? updater(prev) : updater;

				// Auto-hide empty columns based on current data (only on initial load).
				if (typeof updater !== 'function')
					setAutoHideColumnVisibility({ data: services, visibility: newValue });

				// Persist hidden columns.
				persistColumnVisibility(newValue);

				resetColumnVisibility({ data: services, visibility: newValue });

				return newValue;
			});
		},
		[services],
	);

	// Initialise the column visibility, and column order.
	// Remove the table instance when exiting table view.
	// biome-ignore lint/correctness/useExhaustiveDependencies: Startup dependency.
	useEffect(() => {
		// Column visibility.
		setTableColumnVisibility(loadColumnVisibility());

		// Column order (columns absent from the stored order keep their defined position).
		const storedOrder =
			localStorage.getItem(TABLE_COLUMNS_ORDER_STORAGE_KEY)?.split(',') ?? [];
		setTableColumnOrder(mergeColumnOrder(storedOrder));

		return () => setTableInstance(undefined);
	}, []);

	return (
		<div className="w-full">
			<DataTable
				columnOrder={tableColumnOrder}
				columns={columns}
				columnVisibility={tableColumnVisibility}
				data={services}
				dnd={{
					enabled: editMode,
					getItemId: getItemId,
					onDragEnd: handleDragEnd,
					order: order,
					sensors: sensors,
				}}
				getRowClassName={getRowClassName}
				noDataMessage="No services."
				onTableReady={setTableInstance}
				resetSorting={resetSorting}
				resetSortingSignal={resetSortingSignal}
				setColumnOrder={setTableColumnOrder}
				setColumnVisibility={setColumnVisibilityMaster}
				sortToOrder={{
					enabled: editMode,
					getItemId: getItemId,
					onOrderChange: onSortedOrderChange,
				}}
			/>
		</div>
	);
};
