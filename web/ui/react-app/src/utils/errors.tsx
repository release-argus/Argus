import { FieldError, FieldErrors } from 'react-hook-form';

import { StringStringMap } from 'types/config';
import { isEmptyObject } from 'utils';

/**
 * getNestedError gets the error for a potentially nested key in a react-hook-form errors object.
 *
 * @param errors - The errors object from react-hook-form.
 * @param key - The key to get the error for.
 * @returns The error for the provided key.
 */
export const getNestedError = (
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	errors: any,
	key: string,
): FieldError | undefined =>
	key.split('.').reduce((acc, key) => acc?.[key], errors);

/**
 * Extracts and flattens errors from a react-hook-form errors object
 *
 * e.g.
 *
 * { first: { second: [ {item1: {message: "reason"}}, {item2: {message: "otherReason"}} ] } }
 *
 * becomes
 *
 * { first.second.0.item1: "reason", first.second.1.item2: "otherReason"}.
 *
 * @param errors - The react-hook-form errors object.
 * @param path - The path to limit the errors to.
 * @returns The flattened errors object for the provided path.
 */
export const extractErrors = (
	errors: FieldErrors,
	path = '',
): StringStringMap | undefined => {
	const flatErrors: StringStringMap = {};
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const traverse = (prefix: string, obj: any) => {
		for (const key in obj) {
			const value = obj[key];
			const fullPath = prefix ? `${prefix}.${key}` : key;
			if (!fullPath.startsWith(path) || value === null) continue;

			if (typeof value === 'object') {
				if ('message' in value && 'ref' in value) {
					const trimmedPath = path ? fullPath.substring(path.length + 1) : fullPath;
					flatErrors[trimmedPath] = value.message;
				} else {
					traverse(fullPath, value);
				}
			}
		}
	};

	traverse('', errors);
	return isEmptyObject(flatErrors) ? undefined : flatErrors;
};

