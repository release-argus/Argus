import { DndContext, type DragEndEvent } from '@dnd-kit/core';
import {
	arrayMove,
	SortableContext,
	verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { useQueryClient } from '@tanstack/react-query';
import { Eye } from 'lucide-react';
import { type FC, useCallback } from 'react';
import { useToolbar } from '@/components/approvals/toolbar/toolbar-context';
import { Button } from '@/components/ui/button';
import type { ExtraColumnMeta } from '@/components/ui/data-table';
import { DropdownMenuCheckboxItemSortable } from '@/components/ui/dropdown-checkbox-sortable.tsx';
import {
	DropdownMenu,
	DropdownMenuCheckboxItem,
	DropdownMenuContent,
	DropdownMenuGroup,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import Tip from '@/components/ui/tip';
import {
	ACTIVE_HIDE_VALUES,
	APPROVALS_TOOLBAR_VIEW,
	approvalsToolbarViewOptions,
	DEFAULT_HIDE_VALUE,
	HideValue,
	type HideValueType,
	TABLE_COLUMNS_ORDER_STORAGE_KEY,
	TABLE_COLUMNS_VISIBLE_STORAGE_KEY,
	toolbarHideOptions,
} from '@/constants/toolbar';
import { getServiceSummaries } from '@/hooks/use-services';
import { resetColumnVisibility } from '@/pages/approvals/layouts/table/column-visibility';
import { columns } from '@/pages/approvals/layouts/table/columns.tsx';

type HideOptionKey = (typeof toolbarHideOptions)[number]['key'];

/**
 * FilterDropdown
 *
 * A dropdown component for toggling visibility filters on services.
 * It lists all `HIDE_OPTIONS`, allowing users to show or hide specific statuses,
 * and includes a reset option to restore default visibility (`DEFAULT_HIDE_VALUE`).
 */
const FilterDropdown: FC = () => {
	const queryClient = useQueryClient();
	const {
		values,
		setHide,
		setView,
		tableInstance,
		tableColumnVisibility,
		setTableColumnVisibility,
		tableColumnOrder,
		setTableColumnOrder,
	} = useToolbar();
	const currentHideValues = values.hide;

	const onDragEnd = (event: DragEndEvent) => {
		const { active, over } = event;
		// If it hasn't moved, exit.
		if (active.id === over?.id) return;

		// Swap the indexes.
		const oldIndex = tableColumnOrder.indexOf(String(active.id));
		const newIndex = over ? tableColumnOrder.indexOf(String(over.id)) : -1;

		const newOrder = arrayMove(tableColumnOrder, oldIndex, newIndex);

		setTableColumnOrder(newOrder);
		tableInstance?.setColumnOrder(newOrder);
	};

	// Sortable columns and visibility.
	const tableColumnOptions = tableInstance ? (
		<DndContext onDragEnd={onDragEnd}>
			<SortableContext
				items={tableColumnOrder}
				strategy={verticalListSortingStrategy}
			>
				{tableColumnOrder.map((colID) => {
					const col = tableInstance
						?.getAllColumns()
						.find((c) => c.id === colID);
					if (!col) return null;
					const meta = col.columnDef.meta as ExtraColumnMeta | undefined;

					return (
						<DropdownMenuCheckboxItemSortable
							checked={
								tableColumnVisibility ? !!tableColumnVisibility[col.id] : true
							}
							id={col.id}
							key={col.id}
							label={meta?.label ?? col.id}
							onCheckedChange={(value) => col.toggleVisibility(!!value)}
						/>
					);
				})}
			</SortableContext>
		</DndContext>
	) : null;

	const handleHideOptionClick = useCallback(
		(key: HideOptionKey) => {
			const option = toolbarHideOptions.find((opt) => opt.key === key);
			if (!option) return;
			const { value: clickedValue } = option;

			const toggle = (val: number) => {
				const newValues = currentHideValues.includes(val as HideValueType)
					? currentHideValues.filter((v) => v !== val)
					: [...currentHideValues, val];
				setHide(newValues);
			};

			const flipActiveValues = () => {
				const newActiveHidden = ACTIVE_HIDE_VALUES.filter(
					(v) => !currentHideValues.includes(v),
				);
				const inactiveIsHidden = currentHideValues.includes(HideValue.Inactive);
				const newValues = inactiveIsHidden
					? [...newActiveHidden, HideValue.Inactive]
					: newActiveHidden;
				setHide(newValues);
			};

			if (ACTIVE_HIDE_VALUES.includes(clickedValue)) {
				const otherActiveValues = ACTIVE_HIDE_VALUES.filter(
					(v) => v !== clickedValue,
				);
				// If all other active statuses hidden, flip them all.
				if (otherActiveValues.every((v) => currentHideValues.includes(v))) {
					flipActiveValues();
					return;
				}
			}

			toggle(clickedValue);
		},
		[currentHideValues, setHide],
	);

	const handleResetHideFilters = useCallback(() => {
		setHide(DEFAULT_HIDE_VALUE);
	}, [setHide]);

	// biome-ignore lint/correctness/useExhaustiveDependencies: queryClient stable.
	const handleResetColumns = useCallback(() => {
		const serviceSummaries = getServiceSummaries(queryClient);

		const visibility = {};
		resetColumnVisibility({
			data: serviceSummaries,
			visibility: visibility,
		});
		setTableColumnVisibility(visibility);

		// Persist column visibility.
		localStorage.setItem(
			TABLE_COLUMNS_VISIBLE_STORAGE_KEY,
			Object.entries(visibility)
				.filter(([_, isVisible]) => isVisible)
				.map(([columnID]) => columnID)
				.join(','),
		);

		// Reset order.
		const fallbackOrder = columns
			.map((c) => c.id)
			.filter((id): id is string => typeof id === 'string');
		setTableColumnOrder(fallbackOrder);

		// Persist column order.
		localStorage.setItem(
			TABLE_COLUMNS_ORDER_STORAGE_KEY,
			fallbackOrder.join(','),
		);
	}, []);

	const filterButtonTooltip = 'Filter shown services';

	return (
		<DropdownMenu>
			<DropdownMenuTrigger asChild>
				<div className="h-full cursor-pointer">
					<Tip
						content={filterButtonTooltip}
						delayDuration={500}
						touchDelayDuration={250}
					>
						<Button
							aria-label={filterButtonTooltip}
							className="rounded-e-none"
							size="icon-md"
							variant="outline"
						>
							<Eye />
						</Button>
					</Tip>
				</div>
			</DropdownMenuTrigger>
			<DropdownMenuContent className="w-max">
				<DropdownMenuGroup>
					<DropdownMenuLabel>Filters:</DropdownMenuLabel>
					{toolbarHideOptions.map(({ key, label, value }) => {
						const isSelected = currentHideValues.includes(value);
						return (
							<DropdownMenuCheckboxItem
								checked={isSelected}
								key={key}
								onClick={() => handleHideOptionClick(key)}
							>
								{label}
							</DropdownMenuCheckboxItem>
						);
					})}
					<DropdownMenuItem
						className="cursor-pointer"
						onClick={handleResetHideFilters}
					>
						Reset
					</DropdownMenuItem>
				</DropdownMenuGroup>
				<DropdownMenuSeparator className="sm:hidden" />
				<DropdownMenuGroup className="sm:hidden">
					<DropdownMenuLabel>Layout:</DropdownMenuLabel>
					{Object.values(approvalsToolbarViewOptions).map((option) => (
						<DropdownMenuCheckboxItem
							checked={
								option.value === APPROVALS_TOOLBAR_VIEW.GRID.value
									? values.view === APPROVALS_TOOLBAR_VIEW.GRID.value
									: values.view === APPROVALS_TOOLBAR_VIEW.TABLE.value
							}
							key={option.value}
							onClick={() => setView(option.value)}
						>
							{option.label}
						</DropdownMenuCheckboxItem>
					))}
				</DropdownMenuGroup>
				{tableInstance && (
					<>
						<DropdownMenuSeparator />
						<DropdownMenuGroup>
							<DropdownMenuLabel>Columns:</DropdownMenuLabel>
							{tableColumnOptions}
							<DropdownMenuItem
								className="cursor-pointer"
								onClick={handleResetColumns}
							>
								Reset
							</DropdownMenuItem>
						</DropdownMenuGroup>
					</>
				)}
			</DropdownMenuContent>
		</DropdownMenu>
	);
};

export default FilterDropdown;
