/**
 * A replacer function for JSON.stringify that converts `undefined` values to `null`.
 * @param _key - The key of the property being stringified.
 * @param value - The value of the property being stringified.
 */
export const replaceUndefinedWithNull = (_key: string, value: unknown) =>
	value ?? null;
