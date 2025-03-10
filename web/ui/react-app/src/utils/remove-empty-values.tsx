/**
 * Recursively trims the object, removing empty objects/values.
 *
 * @param obj - The object to remove empty values from.
 * @returns The object with all empty values removed.
 */

import { isEmptyArray, isEmptyObject } from './is-empty';

import isEmptyOrNull from './is-empty-or-null';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const removeEmptyValues = (obj: { [x: string]: any }) => {
	for (const key in obj) {
		// [] Array.
		if (Array.isArray(obj[key])) {
			// Empty array - remove.
			if (isEmptyArray(obj[key])) delete obj[key];
			// {} Object.
		} else if (
			typeof obj[key] === 'object' &&
			!['notify', 'webhook'].includes(key) // Not notify/webhook as they may be empty to reference globals.
		) {
			// Check object.
			removeEmptyValues(obj[key]);
			// Empty object - remove.
			if (isEmptyObject(obj[key])) {
				delete obj[key];
			}
			// "" Empty/undefined string - remove.
		} else if (isEmptyOrNull(obj[key])) delete obj[key];
	}
	return obj;
};

export default removeEmptyValues;
