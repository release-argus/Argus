import { useMemo } from 'react';
import { useFormState } from 'react-hook-form';
import type { StringStringMap } from '@/types/util';
import { extractErrors } from '@/utils';

/**
 * Tracks all errors under a field in a form.
 *
 * @param name - Form field name.
 * @param wanted - Whether to return the error.
 * @returns The errors under the field.
 */
export const useErrors = (
	name: string,
	wanted?: boolean,
): StringStringMap | undefined => {
	const { errors } = useFormState({ name: name });

	return useMemo(() => {
		if (!wanted) return undefined;
		return extractErrors(errors, name);
	}, [errors, name, wanted]);
};
