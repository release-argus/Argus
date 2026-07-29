import { COLUMN_IDS } from '@/pages/approvals/layouts/table/columns';

/**
 * Merges a persisted column order with the known columns.
 *
 * Unknown (and duplicate) IDs are dropped, and any column missing from `storedOrder`
 * (i.e. one added in a later release) is inserted at its position in the column
 * definitions rather than appended to the end.
 *
 * @param storedOrder - The persisted order of column IDs.
 * @returns The order of every known column ID.
 */
export const mergeColumnOrder = (storedOrder: string[]): string[] => {
	const known = new Set(COLUMN_IDS);
	const order = [...new Set(storedOrder.filter((id) => known.has(id)))];

	// Insert missing columns after the column they follow in the definitions.
	let insertAt = 0;
	for (const id of COLUMN_IDS) {
		const index = order.indexOf(id);
		if (index !== -1) {
			insertAt = index + 1;
			continue;
		}

		order.splice(insertAt, 0, id);
		insertAt++;
	}

	return order;
};
