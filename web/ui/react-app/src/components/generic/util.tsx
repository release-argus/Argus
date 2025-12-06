/**
 * Pluralise a string.
 *
 * @param str - The string to pluralise.
 * @param count - The count of the string.
 * @returns The pluralised string (if count not 1).
 */
export const pluralise = (str: string, count?: number): string => {
	if (count !== 1) return str + 's';
	return str;
};
