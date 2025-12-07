import type { VisibilityState } from '@tanstack/react-table';
import { columns } from '@/pages/approvals/layouts/table/columns';
import { isEmptyOrNull } from '@/utils';
import type { ServiceSummary } from '@/utils/api/types/config/summary';

type ColumnVisibilityProps = {
	/* The current visibility state. */
	visibility: VisibilityState;
	/* The table data. */
	data: ServiceSummary[];
};

/**
 * Updates the given `visibility` map by automatically hiding columns whose values
 * are empty across all rows.
 *
 * For each column configured with `meta.hideWhenAllValuesEmpty`, this function checks
 * whether *all* values for that column in `data` are empty.
 * If so, the column is marked as hidden in `visibility`. Columns already hidden are
 * left unchanged.
 *
 * Note: This function mutates the `visibility` object in place.
 *
 * @param visibility - A map of column IDs to their visibility state.
 * @param data - The dataset for the table.
 */
export const setAutoHideColumnVisibility = ({
	visibility,
	data,
}: ColumnVisibilityProps) => {
	for (const col of columns) {
		if (col.meta?.hideWhenAllValuesEmpty && 'accessorKey' in col) {
			// Skip if already hidden.
			if (!visibility[col.accessorKey]) continue;

			const key = col.accessorKey as keyof (typeof data)[0];
			const allEmpty = data.every((row) => isEmptyOrNull(row[key]));
			visibility[col.accessorKey] = !allEmpty;
		}
	}
};

/**
 * Resets the column visibility state when no columns are currently visible.
 *
 * If at least one column is already visible, the existing `visibility` map is
 * returned unchanged. Otherwise, this function:
 *   1. Enables all columns by setting every column's visibility to `true`.
 *   2. Applies automatic hiding rules, hiding certain columns whose values are empty across all rows.
 *
 * Note: This function mutates the `visibility` object in place.
 *
 * @param visibility - A map of column IDs to their visibility state.
 * @param data - The dataset for the table.
 */
export const resetColumnVisibility = ({
	visibility,
	data,
}: ColumnVisibilityProps) => {
	const trueCount = Object.values(visibility).filter(Boolean).length;
	if (trueCount > 0) return;

	// Enable all columns when no columns are visible.
	for (const col of columns) {
		const id = col.id ?? ('accessorKey' in col ? col.accessorKey : '');
		visibility[id] = true;
	}
	setAutoHideColumnVisibility({ data, visibility });
};
