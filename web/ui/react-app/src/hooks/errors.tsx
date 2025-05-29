import { FieldError, useFormState } from 'react-hook-form';
import { extractErrors, getNestedError } from 'utils';

import { StringStringMap } from 'types/config';
import { useMemo } from 'react';

/**
 * Tracks an error of a field in a form.
 *
 * @param name - Form field name.
 * @param wanted - Whether to return the error.
 * @returns The error of the field.
 */
export const useError = (
	name: string,
	wanted?: boolean,
): FieldError | undefined => {
	const { errors } = useFormState({ name: name, exact: true });
	return wanted ? getNestedError(errors, name) : undefined
};

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

	const extracted = useMemo(() => {
		if (!wanted) return undefined;
		return extractErrors(errors, name);
	}, [errors, name, wanted]);

	return extracted;
};
