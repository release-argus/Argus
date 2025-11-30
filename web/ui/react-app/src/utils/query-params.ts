import type {
	DeployedVersionLookupSchema,
	DeployedVersionLookupSchemaDefault,
} from '@/utils/api/types/config-edit/service/types/deployed-version';
import type {
	LatestVersionLookupSchema,
	LatestVersionLookupSchemaDefault,
} from '@/utils/api/types/config-edit/service/types/latest-version';
import { isEmptyObject, isEmptyOrNull } from '@/utils/is-empty';
import { replaceUndefinedWithNull } from '@/utils/json-stringify-helpers';

// Helper to turn an object's shape into a "null" overlay (all leaf values null).
const nullifyObject = (obj: unknown): unknown => {
	if (obj == null || Array.isArray(obj) || typeof obj !== 'object') return null;

	const result: Record<string, unknown> = {};
	for (const k of Object.keys(obj as Record<string, unknown>)) {
		const v = (obj as Record<string, unknown>)[k];
		// Recurse objects, set scalars/arrays to null.
		result[k] =
			v && typeof v === 'object' && !Array.isArray(v)
				? (nullifyObject(v) as Record<string, unknown>)
				: null;
	}
	return result;
};

/**
 * Get the differences in key/value pairs between two objects recursively.
 *
 * @param newObj - The new object to compare.
 * @param oldObj - The old object to compare.
 * @returns The differences between the two objects.
 */
export const deepDiff = <T extends Record<string, unknown>>(
	newObj: T,
	oldObj?: T | null,
): Partial<T> => {
	// If `oldObj` empty/undefined/null, return the newObj.
	// e.g. DeployedVersion has no defaults.
	if (isEmptyOrNull(oldObj)) return newObj;

	// get all keys from both objects.
	const allKeys = [
		...new Set([...Object.keys(oldObj), ...Object.keys(newObj)]),
	];

	return allKeys.reduce<Partial<T>>((acc, key) => {
		const newValue = newObj[key];
		const oldValue = oldObj[key];

		// Skip unchanged.
		if ((newValue ?? '') === (oldValue ?? '')) return acc;

		// If newValue missing and oldValue defined, represent deletion as null (or nested nulls for objects).
		if (newValue === undefined && oldValue !== undefined) {
			const valueToSet =
				oldValue && typeof oldValue === 'object' && !Array.isArray(oldValue)
					? (nullifyObject(oldValue) as T[keyof T])
					: (null as T[keyof T]);
			return { ...acc, [key]: valueToSet };
		}

		const isNewObj =
			newValue && typeof newValue === 'object' && !Array.isArray(newValue);
		const isOldObj =
			oldValue && typeof oldValue === 'object' && !Array.isArray(oldValue);

		let diffValue: T[keyof T] | undefined = undefined;

		if (isNewObj && isOldObj) {
			const subDiff = deepDiff(
				newValue as Record<string, unknown>,
				oldValue as Record<string, unknown>,
			);
			if (!isEmptyObject(subDiff)) diffValue = subDiff as T[keyof T];
		} else if (Array.isArray(newValue) || Array.isArray(oldValue)) {
			const newArray = Array.isArray(newValue) ? newValue : [];
			const oldArray = Array.isArray(oldValue) ? oldValue : [];

			const hasChanges =
				newArray.length !== oldArray.length ||
				newArray.some((item, index) => {
					const oldItem = oldArray[index];
					if (
						item != null &&
						typeof item === 'object' &&
						oldItem != null &&
						typeof oldItem === 'object'
					) {
						return Object.keys(deepDiff(item, oldItem)).length > 0;
					}
					return item !== oldItem;
				});

			if (hasChanges) diffValue = newValue as T[keyof T];
		} else {
			// Handle scalars and type mismatches.
			diffValue = newValue as T[keyof T];
		}

		if (diffValue !== undefined) {
			return { ...acc, [key]: diffValue };
		}
		return acc;
	}, {});
};

/**
 * Encode the given key/value pair into a query param string.
 *
 * @param key - The key of the query param.
 * @param value - The value of the query param.
 * @param omitUndefined - Whether to omit undefined values.
 * @returns An encoded query param string for a given key/value pair.
 */
export const stringifyQueryParam = (
	key: string,
	value?: string | number | boolean | null,
	omitUndefined?: boolean,
) =>
	omitUndefined && value == null
		? ''
		: `${key}=${encodeURIComponent(value ?? '')}`;

/**
 * Converts an object to query params
 *
 * @param params - The object to convert to query params.
 * @returns A string of query params joined with '&' and prefixed with '?' (if not empty).
 */
export const convertToQueryParams = (
	params: Record<string, string | number | boolean | null | undefined>,
) => {
	const queryParams = Object.keys(params)
		.filter((key) => (params[key] ?? '') !== '') // Filter out empty strings.
		.map((key) => {
			return stringifyQueryParam(key, params[key]);
		})
		.join('&');
	return queryParams ? `?${queryParams}` : '';
};

export type GetChangesProps =
	| {
			target: 'latest_version';
			params: LatestVersionLookupSchema;
			defaults?: LatestVersionLookupSchema | LatestVersionLookupSchemaDefault;
	  }
	| {
			target: 'deployed_version';
			params: DeployedVersionLookupSchema;
			defaults?:
				| DeployedVersionLookupSchema
				| DeployedVersionLookupSchemaDefault;
	  };
/**
 * Get the differences between the two objects as query params.
 *
 * @param params - The new object.
 * @param defaults - The old object.
 * @param target - The target object to compare.
 * @returns The query params of any changed values between the two objects.
 */
export const getChanges = ({
	params,
	defaults,
	target,
}: GetChangesProps): string => {
	const changedParams = deepDiff(params, defaults);
	if (target === 'deployed_version' && 'template_toggle' in changedParams) {
		delete changedParams.template_toggle;
	}

	// If no changes, return an empty string.
	if (isEmptyObject(changedParams)) return '';

	// Otherwise, stringify the changes.
	return JSON.stringify(changedParams, replaceUndefinedWithNull);
};
