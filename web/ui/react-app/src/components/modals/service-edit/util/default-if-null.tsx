/**
 * Null value defaulting.
 *
 * @template T - The type of the variable.
 * @param value - The variable to check.
 * @param defaultValue - The value to return if `value` is `null`.
 * @returns The `defaultValue` if value===null, `value` otherwise.
 */
export const defaultIfNull = <T,>(
	value: T | null | undefined,
	defaultValue: T,
): T | undefined => {
	return value === null ? defaultValue : value;
};
