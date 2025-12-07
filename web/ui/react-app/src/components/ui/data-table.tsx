'use client';

import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow, } from '@/components/ui/table';
import { cn } from '@/lib/utils';
import { closestCenter, DndContext, type DragEndEvent, type SensorOptions, } from '@dnd-kit/core';
import type { SensorDescriptor } from '@dnd-kit/core/dist/sensors/types';
import { SortableContext, useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import {
	type ColumnDef,
	flexRender,
	getCoreRowModel,
	getSortedRowModel,
	type Row,
	type SortingState,
	useReactTable,
	type VisibilityState,
} from '@tanstack/react-table';
import { GripVertical } from 'lucide-react';
import { type CSSProperties, type Dispatch, type Ref, type SetStateAction, useEffect, useRef, useState, } from 'react';

export type ExtraColumnMeta = {
	/* Hide the column when all values are empty. */
	hideWhenAllValuesEmpty?: boolean;
	/* Hidden column state. */
	hidden?: boolean;
	/* Column label. */
	label?: string;
};
export type ColumnDefWithMeta<TData> = ColumnDef<TData> & {
	/* Extra column metadata. */
	meta?: ExtraColumnMeta;
};

type SortableRowProps<TData> = {
	/* React Table row. */
	row: Row<TData>;
	/* Function to get the unique ID of a row. */
	getItemId: (row: TData) => string;
	/* Function to get extra row CSS class names. */
	getRowClassName?: (row: Row<TData>) => string | undefined;
};

/**
 * A drag-and-drop sortable table row.
 *
 * @param row - React Table row.
 * @param getItemId - Function to get the unique ID of a row.
 * @param enabled - Whether drag-and-drop sorting is enabled.
 * @param getRowClassName - Function to get extra row CSS class names.
 */
const SortableRow = <TData, >({
	row,
	getItemId,
	getRowClassName,
}: SortableRowProps<TData>) => {
	const id = getItemId(row.original);
	const {
		attributes,
		listeners,
		setNodeRef,
		transform,
		transition,
		isDragging,
	} = useSortable({ id: id });

	const style: CSSProperties = {
		transform: CSS.Transform.toString(transform),
		transition,
	};

	return (
		<TableRow
			className={cn(
				'odd:bg-muted/30',
				isDragging && 'z-100 bg-secondary!',
				getRowClassName?.(row),
			)}
			data-state={row.getIsSelected() && 'selected'}
			ref={setNodeRef as Ref<HTMLTableRowElement>}
			style={style}
		>
			<TableCell className="h-full w-8 py-0 align-middle">
				<Button
					{...listeners}
					{...attributes}
					aria-label="Drag handle"
					className="cursor-grab touch-none px-0! py-2 text-muted-foreground"
					size="sm"
					variant="ghost"
				>
					<GripVertical/>
				</Button>
			</TableCell>
			{row.getVisibleCells().map((cell) => (
				<TableCell key={cell.id}>
					{flexRender(cell.column.columnDef.cell, cell.getContext())}
				</TableCell>
			))}
		</TableRow>
	);
};

/* Drag and drop properties. */
type DataTableDndProps<TData> = {
	dnd:
		| {
		/* Flag to indicate whether drag-and-drop sorting is enabled. */
		enabled: false;
	}
		| {
		/* Flag to indicate whether drag-and-drop sorting is enabled. */
		enabled: true;
		/* The current order of rows. */
		order: string[];
		/* An array of sensor descriptors for drag-and-drop operations. */
		sensors: SensorDescriptor<SensorOptions>[];
		/* A function invoked after the drag-and-drop interaction on services is completed. */
		onDragEnd: (event: DragEndEvent) => void;
		/* Function to get the unique ID of a row. */
		getItemId: (row: TData) => string;
	};
};

/* Sorting properties that can optionally push sorting upwards. */
type DataTableSortPushProps<TData> = {
	sortToOrder:
		| {
		/* Flag to indicate whether sorting is enabled */
		enabled: false;
	}
		| {
		/* Flag to indicate whether sorting is enabled */
		enabled: true;
	}
		| {
		/* Flag to indicate whether sorting is enabled */
		enabled: true;
		/* Function to get the unique ID of a row. */
		getItemId: (row: TData) => string;
		/* Function to run when the row order changes. */
		onOrderChange: (ids: string[]) => void;
	};
};

/* Sorting reset properties. */
type DataTableResetProps = {
	/* Resets the table sorting. */
	resetSorting?: () => void;
	/* When this numeric signal changes, the table sorting will be reset. */
	resetSortingSignal?: number;
};

/* Column order properties. */
type DataTableColumnOrderProps = {
	/* Order of columns */
	columnOrder: string[];
	/* Sets the visibility of a column */
	setColumnOrder: Dispatch<SetStateAction<string[]>>;
};

/* Column visibility properties. */
type DataTableColumnVisibilityProps = {
	/* Visible columns */
	columnVisibility: VisibilityState;
	/* Sets the visibility of a column */
	setColumnVisibility: Dispatch<SetStateAction<VisibilityState>>;
};

type DataTableProps<TData> = {
	/* React Table columns. */
	columns: ColumnDefWithMeta<TData>[];
	/* React Table data. */
	data: TData[];
	/* Function to get extra row CSS class names. */
	getRowClassName?: (row: Row<TData>) => string | undefined;
	/* Message to display when there is no data. */
	noDataMessage?: string;
	/* Callback invoked when the table is ready. */
	onTableReady?: (table: ReturnType<typeof useReactTable<TData>>) => void;
} & DataTableDndProps<TData> &
	DataTableSortPushProps<TData> &
	DataTableResetProps &
	DataTableColumnOrderProps &
	DataTableColumnVisibilityProps;

/**
 * A table component that supports optional row drag-and-drop, column sorting, and column reordering.
 *
 * @param columns - React Table columns.
 * @param data - React Table data.
 * @param dnd - Drag and drop properties.
 * @param sortToOrder - Sorting properties that can optionally push sorting upwards.
 * @param resetSortingSignal - When this numeric signal changes, the table sorting will be reset.
 * @param resetSorting - Resets the table sorting.
 * @param getRowClassName - Function to get extra row CSS class names.
 * @param noDataMessage - Message to display when there is no data.
 * @param onTableReady - Callback invoked when the table is ready.
 * @param columnOrder - Order of columns.
 * @param setColumnOrder - Sets the visibility of a column.
 * @param columnVisibility - Visible columns.
 * @param setColumnVisibility - Sets the visibility of a column.
 */
export const DataTable = <TData, >({
	columns,
	data,
	dnd = { enabled: false },
	sortToOrder = { enabled: false },
	resetSortingSignal,
	resetSorting,
	getRowClassName,
	noDataMessage,
	onTableReady,
	columnOrder,
	setColumnOrder,
	columnVisibility,
	setColumnVisibility,
}: DataTableProps<TData>) => {
	const [ sorting, setSorting ] = useState<SortingState>([]);

	const orderRef = useRef<{
		/**
		 * The serialised string of row IDs from the last time `sortToOrder.onOrderChange` was called.
		 * Used to avoid emitting duplicate orders when rows or sorting haven't changed.
		 * */
		lastEmittedOrder: string | null;
		/**
		 * When a reset is requested (e.g. edit mode toggled off),
		 * suppress the push of the current sorted row order upwards.
		 */
		skipNextPush: boolean;
	}>({
		   lastEmittedOrder: null,
		   skipNextPush: false,
	   });

	const table = useReactTable({
		                            columns: columns,
		                            data: data,
		                            enableSorting: sortToOrder.enabled,
		                            getCoreRowModel: getCoreRowModel(),
		                            getSortedRowModel: getSortedRowModel(),
		                            onColumnOrderChange: setColumnOrder,
		                            onColumnVisibilityChange: setColumnVisibility,
		                            onSortingChange: setSorting,
		                            state: {
			                            columnOrder,
			                            columnVisibility,
			                            sorting,
		                            },
	                            });
	// biome-ignore lint/correctness/useExhaustiveDependencies: table stable.
	useEffect(() => {
		if (onTableReady) onTableReady(table);
	}, [ onTableReady ]);

	// Reset sorting when asked by parent (e.g. when order reset is triggered).
	useEffect(() => {
		if (resetSortingSignal === undefined) return;
		setSorting([]);
		orderRef.current.lastEmittedOrder = null;
		// Suppress the next sort-to-order emission that may be caused by this reset.
		orderRef.current.skipNextPush = true;
	}, [ resetSortingSignal ]);

	// Optionally, push the current sorted row order upwards.
	// biome-ignore lint/correctness/useExhaustiveDependencies: sortToOrder.getItemId stable.
	useEffect(() => {
		if (!sortToOrder.enabled || !(
			'getItemId' in sortToOrder
		)) return;

		// If a reset just happened, skip this cycle once.
		if (orderRef.current.skipNextPush) {
			orderRef.current.skipNextPush = false;
			return;
		}

		// Compute current visible row order according to sorting.
		const ids = table
			.getRowModel()
			.rows.map((r) => sortToOrder.getItemId(r.original));
		const key = ids.join('\u0001');
		if (orderRef.current.lastEmittedOrder !== key) {
			orderRef.current.lastEmittedOrder = key;
			sortToOrder.onOrderChange(ids);
		}
	}, [ sorting, data ]);

	const rows = table.getRowModel().rows;

	const tableElement = (
		<Table>
			<TableHeader>
				{table.getHeaderGroups().map((headerGroup) => (
					<TableRow key={headerGroup.id}>
						{dnd.enabled && <TableHead className="w-8"/>}
						{headerGroup.headers
							.filter((header) => header.column.getIsVisible())
							.map((header) => {
								return (
									<TableHead key={header.id}>
										{header.isPlaceholder
										 ? null
										 : flexRender(header.column.columnDef.header, {
												...header.getContext(),
												resetSorting: resetSorting,
											})}
									</TableHead>
								);
							})}
					</TableRow>
				))}
			</TableHeader>
			<TableBody>
				{rows?.length ? (
					rows.map((row) =>
						         dnd.enabled ? (
							         <SortableRow
								         getItemId={dnd.getItemId}
								         getRowClassName={getRowClassName}
								         key={dnd.getItemId(row.original)}
								         row={row}
							         />
						         ) : (
							         <TableRow
								         className={cn(
									         'odd:bg-muted/30',
									         getRowClassName?.(row),
									         row.id,
								         )}
								         data-state={row.getIsSelected() && 'selected'}
								         key={row.id}
							         >
								         {row.getVisibleCells().map((cell) => (
									         <TableCell key={cell.id}>
										         {flexRender(cell.column.columnDef.cell, cell.getContext())}
									         </TableCell>
								         ))}
							         </TableRow>
						         ),
					)
				) : (
					 <TableRow>
						 <TableCell
							 className="h-24 text-center"
							 colSpan={
								 (
									 dnd.enabled ? 1 : 0
								 ) + table.getVisibleLeafColumns().length
							 }
						 >
							 {noDataMessage ?? 'No results.'}
						 </TableCell>
					 </TableRow>
				 )}
			</TableBody>
		</Table>
	);

	return (
		<div className="overflow-hidden rounded-md border">
			{dnd.enabled ? (
				<DndContext
					autoScroll={{
						acceleration: 100,
						enabled: true,
						interval: 5,
						threshold: { x: 0.2, y: 0.15 },
					}}
					collisionDetection={closestCenter}
					onDragEnd={(e) => {
						setSorting([]);
						dnd.onDragEnd(e);
					}}
					sensors={dnd.sensors}
				>
					<SortableContext
						items={table
							.getRowModel()
							.rows.map((r) => dnd.getItemId(r.original))}
					>
						{tableElement}
					</SortableContext>
				</DndContext>
			) : (
				 tableElement
			 )}
		</div>
	);
};
