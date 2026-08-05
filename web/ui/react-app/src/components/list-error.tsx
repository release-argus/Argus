import type { ReactElement } from 'react';
import { APIError, getErrorMessage } from '@/utils/errors';

type ListErrorProps = {
	/* The failed list query's error. */
	error: unknown;
	/* What could not be listed (lower-case plural), e.g. 'users'. */
	resource: string;
};

/**
 * Renders a list query's failure in place of its table.
 */
export const ListError = ({
	error,
	resource,
}: ListErrorProps): ReactElement => (
	<p className="text-muted-foreground text-sm">
		{error instanceof APIError && error.status === 403
			? `You do not have permission to view ${resource}.`
			: getErrorMessage(error)}
	</p>
);
