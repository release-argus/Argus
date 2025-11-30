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
export const isEmptyOrNull = (
	value: unknown,
): value is null | undefined | '' => {
	return (value ?? '') === '';
};

/**
 * Whether the object is non-empty.
 *
 * @param arg - The object to check.
 * @returns `true` when object is non-empty. `false` otherwise.
 */
export const isNonEmptyObject = <T extends Record<string, unknown>>(
	arg: T | undefined,
): arg is T => arg !== undefined && Object.keys(arg).length > 0;
