import type { ZodType } from 'zod';

/**
 * Picks the first non-empty value for a given key from a list of objects.
 *
 * @template T - Type of the objects.
 * @param aVal - The current value to consider.
 * @param defaults - Array of default objects to pick from.
 * @param key - The key whose value should be selected.
 * @returns The first non-empty value found, or an empty array if no other value exists and `aVal` is missing.
 */
const pickDefault = <T>(
	aVal: unknown,
	defaults: Partial<T>[],
	key: keyof T,
) => {
	let chosen: unknown = undefined;

	for (const d of defaults) {
		const val = d[key];

		// 1. Arrays.
		if (Array.isArray(val)) {
			// First non-empty array.
			if (val.length > 0) {
				return val;
			}

			// Empty array as placeholder whilst we evaluate other defaults.
			if (aVal === undefined && chosen === undefined) {
				chosen = val;
			}
			continue;
		}

		// Empty string as placeholder whilst we evaluate other defaults.
		if (val === '') {
			chosen = val;
			continue;
		}

		// null as placeholder whilst we evaluate other defaults.
		if (val === null && chosen !== null) {
			chosen = val;
		}

		// Return the first non-null/emoty value.
		if (val != null) {
			// Other non-empty value.
			return val;
		}
	}
	return chosen;
};

/**
 * Merges an object `a` with a sequence of partial objects, applying default values recursively.
 * If a property exists in both `a` and the defaults, priority is given to the values in `a` unless
 * the value in `a` is an empty string, `null`, or `undefined`. In such cases, the default value is used.
 * If a property contains nested objects, the merging process is applied recursively.
 *
 * @template T - The type of the base object to merge defaults into.
 * @param  a - The base object that takes precedence, or `null` to use only the defaults.
 * @param  rest - One or more partial objects that provide default values.
 * @returns - A new object containing the merged values of `a` and the defaults.
 */
export const applyDefaultsRecursive = <T extends object>(
	a: T | null,
	...rest: (Partial<T> | undefined)[]
): T => {
	const isObject = (val: unknown): val is object =>
		typeof val === 'object' && !Array.isArray(val) && val != null;

	const result = {} as T;

	// Remove any undefined defaults.
	const defaults = [...rest].filter((d): d is Partial<T> => d !== undefined);

	const allKeys = new Set<keyof T>();
	if (a) {
		for (const k of Object.keys(a) as (keyof T)[]) {
			allKeys.add(k);
		}
	}
	for (const d of defaults) {
		for (const k of Object.keys(d) as (keyof T)[]) {
			allKeys.add(k);
		}
	}

	for (const key of allKeys) {
		const aVal = a?.[key];
		const bVal = pickDefault(aVal, defaults, key);

		if (isObject(aVal) && isObject(bVal)) {
			result[key] = applyDefaultsRecursive(aVal, bVal as Partial<typeof aVal>);
		} else if (aVal === '' || aVal == null) {
			result[key] = (bVal ??
				aVal ??
				(typeof bVal === 'string' ? '' : null)) as T[typeof key];
		} else {
			result[key] = aVal;
		}
	}

	return result;
};

/**
 * Ensures that an array has at least two items.
 *
 * @param arr - The array to check.
 * @returns The input array if it has at least two items.
 * @throws An error if the array has fewer than two items.
 */
export const atLeastTwo = <T extends ZodType>(
	arr: [T, ...T[]] | T[],
): [T, ...T[]] => {
	if (arr.length < 2) {
		throw new Error('Instantiated with fewer than two items');
	}

	return arr as [T, ...T[]];
};
