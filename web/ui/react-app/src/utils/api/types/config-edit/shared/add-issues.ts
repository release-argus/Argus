import type { RefinementCtx, ZodError } from 'zod';
import { CUSTOM_ISSUE_CODE } from '@/utils/api/types/config-edit/validators';

type AddZodIssuesToContextParams = {
	/* Context that has an `addIssue` method */
	ctx: Pick<RefinementCtx, 'addIssue'>;
	/* The Zod error from a failed parse */
	error: ZodError;
};

/**
 * Adds all Zod issues from a failed parse result to a context.
 *
 * @param ctx - The context that has an `addIssue` method.
 * @param error - The Zod error from a failed parse.
 */
export const addZodIssuesToContext = ({
	ctx,
	error,
}: AddZodIssuesToContextParams) => {
	for (const issue of error.issues) {
		ctx.addIssue({
			code: CUSTOM_ISSUE_CODE,
			message: issue.message,
			path: issue.path,
		});
	}
};
