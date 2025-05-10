/**
 * Recursively trims the object, removing empty objects/values.
 *
 * @param obj - The object to remove empty values from.
 * @returns The object with all empty values removed.
 */

import { isEmptyArray, isEmptyObject } from './is-empty';

import isEmptyOrNull from './is-empty-or-null';

const removeEmptyValues = (
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	obj: { [x: string]: any },
	excludeKeys: string[] = [],
) => {
	for (const key in obj) {
		if (excludeKeys.find((k) => k === key)) {
			delete obj[key];
			continue
		}

		// [] Array.
		if (Array.isArray(obj[key])) {
			// Empty array - remove.
			if (isEmptyArray(obj[key])) delete obj[key];
			// {} Object.
		} else if (
			typeof obj[key] === 'object' &&
			!['notify', 'webhook'].includes(key) // Not notify/webhook as they may be empty to reference globals.
		) {
			const childExcludeKeys = excludeKeys
				.filter((k) => k.startsWith(`${key}.`))
				.map((k) => k.substring(key.length + 1));
			// Check object.
			removeEmptyValues(obj[key], childExcludeKeys);
			// Empty object - remove.
			if (isEmptyObject(obj[key])) {
				delete obj[key];
			}
			// "" Empty/undefined string - remove.
		} else if (isEmptyOrNull(obj[key])) delete obj[key];
	}
	return JSON.parse(JSON.stringify(obj));
};

export default removeEmptyValues;
