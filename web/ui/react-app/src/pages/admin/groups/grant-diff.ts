import type { Grant } from '@/types/auth';

/** grantKey identifies a grant by value, for matching rows across renders. */
export const grantKey = (grant: Grant) =>
	`${grant.resource}:${grant.action}:${grant.scope.type}:${grant.scope.ref ?? ''}`;

export type GrantDiff = {
	/** modifiedRows[i] reports whether row i is new or edited. */
	modifiedRows: boolean[];
	/** duplicateKeys holds the keys of grants listed more than once. */
	duplicateKeys: Set<string>;
};

/**
 * diffGrants matches rows against the saved set by value. A saved grant is
 * claimed by the first row that matches it, leaving later duplicates marked
 * as modified.
 *
 * savedGrants is undefined while creating, where no row counts as modified.
 */
export const diffGrants = (
	grants: Grant[],
	savedGrants?: Grant[],
): GrantDiff => {
	// How many times each grant appears in the saved set, claimed as rows match it.
	const unclaimed = new Map<string, number>();
	for (const grant of savedGrants ?? []) {
		const key = grantKey(grant);
		unclaimed.set(key, (unclaimed.get(key) ?? 0) + 1);
	}

	const seen = new Set<string>();
	const duplicateKeys = new Set<string>();
	const modifiedRows = grants.map((grant) => {
		const key = grantKey(grant);
		if (seen.has(key)) duplicateKeys.add(key);
		seen.add(key);

		// A row is new/edited when no saved grant matches it.
		const left = unclaimed.get(key) ?? 0;
		if (left === 0) return savedGrants !== undefined;
		unclaimed.set(key, left - 1);
		return false;
	});

	return { duplicateKeys, modifiedRows };
};
