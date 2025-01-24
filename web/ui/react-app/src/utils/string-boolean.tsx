import isEmptyOrNull from './is-empty-or-null';

/**
 * The boolean value of a string.
 *
 * @param str - The string to convert to boolean.
 * @returns The boolean value of the string, or null if string empty.
 */
export const strToBool = (str?: string | boolean | null): boolean | null => {
	if (typeof str === 'boolean') return str;
	if (isEmptyOrNull(str)) return null;
	return ['true', 'yes'].includes(str.toLowerCase());
};

export const boolToStr = (bool?: boolean) =>
	bool === undefined ? '' : bool.toString();
