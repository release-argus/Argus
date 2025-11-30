/**
 * Whether `list` contains a string that `substring` starts with.
 *
 * @param substring - The string to check if it starts with any of the items in the list.
 * @param list - The list of strings to check against.
 * @param undefinedListDefault - The value to return if `list` undefined.
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
 * @param undefinedListDefault - The value to return if `list` undefined.
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
