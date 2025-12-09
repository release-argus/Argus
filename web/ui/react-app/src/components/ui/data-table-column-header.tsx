import { Button } from '@/components/ui/button';
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { cn } from '@/lib/utils';
import type { Column, SortDirection } from '@tanstack/react-table';
import { ArrowDown, ArrowUp, ChevronsUpDown, EyeOff } from 'lucide-react';
import { type HTMLAttributes, useMemo } from 'react';

type SortArrowProps = {
	/* The direction to sort in. */
	sortDirection: SortDirection | false;
	/* Whether the column can be sorted. */
	canSort: boolean;
};

/**
 * Renders a sort arrow based on the sort direction and whether the column can be sorted.
 *
 * @param sortDirection - The direction being sorted in, or false if not sorted.
 * @param canSort - Whether the column can be sorted.
 * @returns An arrow indicating the sort direction, up-down arrow if sorting possible, null otherwise.
 */
const SortArrow = ({ sortDirection, canSort }: SortArrowProps) => {
	if (!canSort) return null;

	if (sortDirection === 'asc') return <ArrowUp/>;
	if (sortDirection === 'desc') return <ArrowDown/>;

	return <ChevronsUpDown/>;
};

type DataTableColumnHeaderProps<TData, TValue> =
	HTMLAttributes<HTMLDivElement> & {
	/* The column to render the header for. */
	column: Column<TData, TValue>;
	/* The title of the column. */
	title: string;
	/* Triggers the 'sorting reset' signal to a new value. */
	resetSorting?: () => void;
};

/**
 * Renders a column header with sorting and hiding options.
 *
 * @param column - The column to render the header for.
 * @param title - The title of the column.
 * @param className - Additional class names to apply to the header.
 * @param resetSorting - Triggers the 'sorting reset' signal to a new value, resetting the sorting state of the form.
 * @returns The rendered column header, with sorting and hiding options if applicable.
 */
export const DataTableColumnHeader = <TData, TValue>({
	column,
	title,
	className,
	resetSorting,
}: DataTableColumnHeaderProps<TData, TValue>) => {
	const canSort = column.getCanSort();
	const canHide = column.getCanHide();

	// SortArrow wasn't updating without this memo.
	// biome-ignore lint/correctness/useExhaustiveDependencies: canSort stable.
	const sortArrow = useMemo(
		() => <SortArrow canSort={canSort} sortDirection={column.getIsSorted()}/>,
		[ column.getIsSorted() ],
	);

	if (!canSort && !canHide) {
		return <div className={className}>{title}</div>;
	}

	return (
		<div className={cn('flex items-center gap-2', className)}>
			<DropdownMenu>
				<DropdownMenuTrigger asChild>
					<Button
						className="-ml-3 h-8 data-[state=open]:bg-accent"
						size="sm"
						variant="ghost"
					>
						<span>{title}</span>
						{/*<SortArrow canSort={canSort} sortDirection={column.getIsSorted()} />*/}
						{sortArrow}
					</Button>
				</DropdownMenuTrigger>
				<DropdownMenuContent align="start">
					{canSort &&
					 [ 'asc', 'desc' ].map((dir) => (
						 <DropdownMenuItem
							 className="capitalize"
							 key={dir}
							 onClick={() =>
								 column.getIsSorted() === dir
								 ? resetSorting?.()
								 : column.toggleSorting(dir === 'desc')
							 }
						 >
							 {dir === 'asc' ? <ArrowUp/> : <ArrowDown/>}
							 {dir}
						 </DropdownMenuItem>
					 ))}

					{canSort && canHide && <DropdownMenuSeparator/>}

					{canHide && (
						<DropdownMenuItem onClick={() => column.toggleVisibility(false)}>
							<EyeOff/>
							Hide
						</DropdownMenuItem>
					)}
				</DropdownMenuContent>
			</DropdownMenu>
		</div>
	);
};
