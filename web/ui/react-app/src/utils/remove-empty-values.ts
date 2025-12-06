import { isEmptyArray, isEmptyObject, isEmptyOrNull } from '@/utils/is-empty';

type Primitive = string | number | boolean | null | undefined;
interface NestedObject {
	[key: string]:
		| Primitive[][]
		| Primitive[]
		| Primitive
		| NestedObject
		| NestedObject[]
		| NestedObject[][];
}

/**
 * Removes empty values from an object while preserving specified keys.
 * Empty values include: empty arrays, empty objects, empty strings, null, and undefined.
 *
 * @param obj - The object to clean.
 * @param excludeKeys - Keys to exclude from removal.
 * @returns A new object with empty values removed.
 */
const removeEmptyValues = (
	obj: NestedObject,
	excludeKeys: string[] = [],
): NestedObject =>
	Object.entries(obj).reduce<NestedObject>((acc, [key, value]) => {
		const keyStr = String(key);

		if (value && typeof value === 'object' && !Array.isArray(value)) {
			// Nested object.
			const substringLength = keyStr.length + 1;
			const childExcludeKeys = excludeKeys
				.filter((k) => k.startsWith(`${keyStr}.`))
				.map((k) => k.substring(substringLength));

			const cleaned = removeEmptyValues(value, childExcludeKeys);

			if (!shouldRemoveValue(cleaned, keyStr, excludeKeys)) {
				acc[keyStr] = cleaned;
			}
		} else if (!shouldRemoveValue(value, keyStr, excludeKeys)) {
			acc[keyStr] = value;
		}

		return acc;
	}, {});

/**
 * Returns whether `value` is empty and contains no excluded keys.
 *
 * @param value - The value to check.
 * @param key - The key of the value in the object.
 * @param excludeKeys - Keys to exclude from removal.
 * @returns True if the value should be removed, false otherwise.
 */
const shouldRemoveValue = (
	value: unknown,
	key: string,
	excludeKeys: string[],
) => {
	if (excludeKeys.some((k) => k.split('.')[0] === key)) return false;
	if (Array.isArray(value)) return isEmptyArray(value);
	if (typeof value === 'object' && value !== null)
		return isEmptyObject(value as Record<string, unknown>);
	return isEmptyOrNull(value);
};

export default removeEmptyValues;
