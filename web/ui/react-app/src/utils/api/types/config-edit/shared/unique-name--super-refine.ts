import type { RefinementCtx } from 'zod';
import {
	CUSTOM_ISSUE_CODE,
	UNIQUE_MESSAGE,
} from '@/utils/api/types/config-edit/validators';

/**
 * Checks for duplicate names in a list of objects with a `name` property.
 *
 * @param arg - The list of objects to check for duplicate names.
 * @param ctx - The refinement context to report issues.
 */
export const superRefineNameUnique = <T extends { name: string }[]>(
	arg: T,
	ctx: RefinementCtx,
) => {
	const seen = new Map<string, number>(); // name -> first index found.

	// Check for duplicate names.
	for (let index = 0; index < arg.length; index++) {
		const name = arg[index].name;
		if (!name) continue;

		const firstIndex = seen.get(name);
		if (firstIndex === undefined) {
			// Record the first index of this name.
			seen.set(name, index);
		} else {
			// Add an issue for this duplicate name, and the first occurence.
			for (const i of [firstIndex, index]) {
				ctx.addIssue({
					code: CUSTOM_ISSUE_CODE,
					message: UNIQUE_MESSAGE,
					path: [i, 'name'],
				});
			}
		}
	}
};
