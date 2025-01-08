import { isEmptyArray } from './is-empty';
import isEmptyOrNull from './is-empty-or-null';

/**
 * Whether the fieldValues differ from the defaults.
 *
 * @param a - The fieldValues to compare.
 * @param b - The defaults to compare against.
 * @param allowDefined - The keys of `a` whose values may match `b` and still be considered unchanged.
 * @param key - The key path of the current fieldValues.
 * @returns Whether `a` is the same as `b` (values of keys in allowDefined are empty or match `b`, other values are empty).
 */
export function diffObjects<T>(
	a?: T,
	b?: T,
	allowDefined?: string[],
	key?: string,
): boolean {
	// No defaults.
	if (b === undefined && a !== undefined) return (a ?? '') !== (b ?? '');
	// If `a` is empty/undefined, treat as unchanged.
	if (isEmptyOrNull(a)) return false;
	// If no defaults, treat as changed.
	if (isEmptyOrNull(b)) return true;
	// If `a` is an array, check it is the same length as `b`.
	if (
		Array.isArray(a) &&
		(!Array.isArray(b) ||
			a.length !== b.length ||
			// If only one has an ID, check it is not a length difference of 1.
			(a.hasOwnProperty('id') != b.hasOwnProperty('id') &&
				Math.abs(a.length - b.length) !== 1))
	)
		// Non-empty means different as the lengths differ.
		return !isEmptyArray(a);

	if (typeof b === 'object') {
		const keys: Array<keyof T> = Object.keys(a as object) as Array<keyof T>;
		// Recursively check each key in the object.
		for (const k of keys) {
			if (
				diffObjects(
					a?.[k],
					b?.[k],
					allowDefined,
					key ? `${key}.${String(k)}` : String(k),
				)
			)
				// Difference!
				return true;
		}
		// No differences found.
		return false;
	} else if (typeof b === 'string') {
		// Check values match on allowed keys.
		if (containsEndsWith(key || '-', allowDefined)) return a !== b;
		// Else, we have got a difference.
		return true;
	}
	// Check values match on allowed keys. Different otherwise.
	else return containsEndsWith(key || '-', allowDefined) ? a !== b : true;
}

/**
 * Whether `list` contains a string that `substring` starts with.
 *
 * @param substring - The string to check if it starts with any of the items in the list.
 * @param list - The list of strings to check against.
 * @param undefinedListDefault - The value to return if list undefined.
 * @default undefinedListDefault=false
 * @returns Whether the substring starts with any of the items in the list.
 */
export const containsStartsWith = (
	substring: string,
	list?: string[],
	undefinedListDefault = false,
): boolean => {
	return list
		? list.some((item) => substring.startsWith(item))
		: undefinedListDefault;
};

/**
 * Whether `list` contains a string that `substring` ends with.
 *
 * @param substring - The string to check if it ends with any of the items in the list.
 * @param list - The list of strings to check against.
 * @param undefinedListDefault - The value to return if list undefined.
 * @default undefinedListDefault=false
 * @returns Whether the substring ends with any of the items in the list.
 */
export const containsEndsWith = (
	substring: string,
	list?: string[],
	undefinedListDefault = false,
): boolean => {
	return list
		? list.some((item) => substring.endsWith(item))
		: undefinedListDefault;
};
