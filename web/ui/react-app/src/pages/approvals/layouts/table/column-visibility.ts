import type { ColumnVisibilityState } from '@tanstack/react-table';
import {
	TABLE_COLUMNS_HIDDEN_STORAGE_KEY,
	TABLE_COLUMNS_VISIBLE_STORAGE_KEY_LEGACY,
} from '@/constants/toolbar';
import { COLUMN_IDS, columns } from '@/pages/approvals/layouts/table/columns';
import { isEmptyOrNull } from '@/utils';
import type { ServiceSummary } from '@/utils/api/types/config/summary';

type ColumnVisibilityProps = {
	/* The current visibility state. */
	visibility: ColumnVisibilityState;
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
	for (const id of COLUMN_IDS) {
		visibility[id] = true;
	}
	setAutoHideColumnVisibility({ data, visibility });
};

/**
 * Reads the persisted column visibility.
 *
 * Hidden columns are what get persisted, so a column that the user has never seen
 * (i.e. one added in a later release) starts visible rather than hidden.
 * The superseded 'visible columns' value is discarded on the first read.
 *
 * @returns The visibility state of every known column.
 */
export const loadColumnVisibility = (): ColumnVisibilityState => {
	localStorage.removeItem(TABLE_COLUMNS_VISIBLE_STORAGE_KEY_LEGACY);

	const hidden = new Set(
		(localStorage.getItem(TABLE_COLUMNS_HIDDEN_STORAGE_KEY) ?? '')
			.split(',')
			.filter(Boolean),
	);
	// Ignore a state that would leave nothing to show.
	const hideNone = COLUMN_IDS.every((id) => hidden.has(id));

	return COLUMN_IDS.reduce<ColumnVisibilityState>((acc, id) => {
		acc[id] = hideNone || !hidden.has(id);
		return acc;
	}, {});
};

/**
 * Persists the hidden columns of the given visibility state.
 *
 * @param visibility - A map of column IDs to their visibility state.
 */
export const persistColumnVisibility = (visibility: ColumnVisibilityState) => {
	localStorage.setItem(
		TABLE_COLUMNS_HIDDEN_STORAGE_KEY,
		Object.entries(visibility)
			.filter(([_, isVisible]) => !isVisible)
			.map(([columnID]) => columnID)
			.join(','),
	);
};
