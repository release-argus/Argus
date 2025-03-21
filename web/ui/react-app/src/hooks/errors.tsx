import { FieldError, useFormState } from 'react-hook-form';
import { extractErrors, getNestedError } from 'utils';

import { StringStringMap } from 'types/config';

/**
 * Tracks an error of a field in a form.
 *
 * @param name - The name of the field in the form.
 * @param wanted - Whether the error is returned.
 * @returns The error of the field.
 */
export const useError = (
	name: string,
	wanted?: boolean,
): FieldError | undefined => {
	const { errors } = useFormState({ name: name, exact: true });
	if (!wanted) return undefined;
	return getNestedError(errors, name);
};

/**
 * Tracks all errors under a field in a form.
 *
 * @param name - The name of the field in the form.
 * @param wanted - Whether the errors are returned.
 * @returns The errors under the field.
 */
export const useErrors = (
	name: string,
	wanted?: boolean,
): StringStringMap | undefined => {
	const { errors } = useFormState({ name: name, exact: true });
	if (!wanted) return undefined;
	return extractErrors(errors, name);
};
