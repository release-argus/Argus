/**
 * Whether the array is empty, null, or undefined.
 *
 * @param arg - The array to check.
 * @returns true when array is empty, null, or undefined. false otherwise.
 */
export const isEmptyArray = (arg?: unknown[]): arg is [] | undefined =>
	((arg as unknown[]) ?? []).length === 0;

export const isEmptyObject = <T extends Record<string, unknown> | undefined>(
	arg: T,
): boolean => Object.keys(arg ?? {}).length === 0;
