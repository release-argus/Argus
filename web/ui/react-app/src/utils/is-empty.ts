/**
 * Whether the array is empty, null, or undefined.
 *
 * @param arg - The array to check.
 * @returns `true` when array is empty, null, or undefined. `false` otherwise.
 */
export const isEmptyArray = (arg?: unknown[]): arg is [] | undefined =>
	(arg ?? []).length === 0;

/**
 * Whether the object is empty, null, or undefined.
 *
 * @param arg - The object to check.
 * @returns `true` when object is empty, null, or undefined. `false` otherwise.
 */
export const isEmptyObject = (
	arg: Record<string, unknown> | undefined,
): boolean => Object.keys(arg ?? {}).length === 0;

/**
 * Whether the value is empty, null, or undefined.
 *
 * @param value - The value to check.
 * @returns `true` when value is empty, null, or undefined. `false` otherwise.
 */
export const isEmptyOrNull = (value: unknown): value is null | undefined | '' =>
	(value ?? '') === '';

/**
 * Whether the object is non-empty.
 *
 * @param arg - The object to check.
 * @returns `true` when object is non-empty. `false` otherwise.
 */
export const isNonEmptyObject = <T extends Record<string, unknown>>(
	arg: T | undefined,
): arg is T => arg !== undefined && Object.keys(arg).length > 0;

/**
 * Whether the arg is empty.
 *
 * @param arg - The arg to check.
 * @returns `true` when arg is empty. `false` otherwise.
 */
export const isEmpty = (
	arg: unknown,
): arg is undefined | '' | null | {} | [] => {
	if (typeof arg === 'object') {
		// Array.
		if (Array.isArray(arg)) return isEmptyArray(arg);

		// null.
		if (!arg) return true;
		// Dates, Functions, Class instances, Maps, Sets, etc.
		if (Object.getPrototypeOf(arg) !== Object.prototype) return false;

		// Record.
		return isEmptyObject(arg as Record<string, unknown>);
	}

	// String.
	return isEmptyOrNull(arg);
};
