import {
	type ZodArray,
	type ZodDefault,
	type ZodPipe,
	type ZodTransform,
	z,
} from 'zod';
import type { CustomHeaders } from '@/utils/api/types/config/shared';
import {
	headerSchema,
	headerSchemaDefaults,
} from '@/utils/api/types/config-edit/shared/header/schemas';
import { makeDefaultsAwareListPreprocessor } from '@/utils/api/types/config-edit/shared/preprocess';

/**
 * Flattens the headers for the API.
 *
 * @param headers - The CustomHeaders[] to flatten {key: "KEY", value: "VAL"}[].
 * @returns The flattened object {KEY: VAL, ...}.
 */
export const flattenHeaderArray = (headers?: CustomHeaders) => {
	if (!headers) return undefined;
	return headers.reduce<Record<string, string>>((obj, header) => {
		obj[header.key] = header.value;
		return obj;
	}, {});
};

/**
 * Converts 'headers' from an array of objects to a JSON string.
 *
 * @param val - The `CustomHeaders[]` to convert.
 * @returns A JSON string of the headers.
 */
export const preprocessStringFromHeaderArray = z.preprocess((val: unknown) => {
	if (!val || !Array.isArray(val) || val.length === 0) return '';

	const flattened = flattenHeaderArray(val);
	// Using defaults if any key empty.
	if (flattened && Object.keys(flattened).some((arg) => !arg.trim())) return '';

	return JSON.stringify(flattened);
}, z.string());

/**
 * Defaults-aware variant of headers -> string preprocessor.
 * - Empty array -> null
 * - Matches defaults (deep, via matchingFields=[]) -> null
 */
export const preprocessStringFromHeaderArrayWithDefaults = (
	defaults?: Record<string, unknown>[],
) =>
	makeDefaultsAwareListPreprocessor(
		preprocessStringFromHeaderArray.nullable(),
		{
			defaults: defaults,
			matchingFields: [],
		},
	);

/**
 * Converts various header representations into an array of header objects.
 * Accepts a Zod schema and returns a Zod array schema with preprocessing logic for strings and objects.
 *
 * @template T - A ZodType representing the header schema.
 * @param schema - The Zod schema representing the headers.
 */
export const preprocessToHeadersArray = <T extends z.ZodType>(
	schema: T,
): ZodDefault<ZodPipe<ZodTransform, ZodArray<T>>> =>
	z
		.preprocess((arg) => {
			// If string, try to parse as JSON.
			if (typeof arg === 'string') {
				const parsed = JSON.parse(arg) as unknown;
				if (parsed && typeof parsed === 'object') {
					return Object.entries(parsed).map(([key, value]) => ({
						key,
						value: typeof value === 'string' ? value : JSON.stringify(value),
					}));
				}
				return [];
			}

			// If object, try to parse each key-value pair.
			if (
				arg &&
				typeof arg === 'object' &&
				!Array.isArray(arg) &&
				!(Object.keys(arg).length === 2 && 'key' in arg && 'value' in arg)
			) {
				return Object.entries(arg).map(([key, value], i) => ({
					key: key,
					old_index: i,
					value: value,
				}));
			}

			// Unknown type, return as is.
			return arg;
		}, z.array(schema))
		.default([]);

/* Array of Header objects (min length 1 on key and value) */
export const headersSchema = preprocessToHeadersArray(headerSchema);
/* Array of Header objects (no validation) */
export const headersSchemaDefaults =
	preprocessToHeadersArray(headerSchemaDefaults);

export const preprocessHeaderArrayWithDefaults = (defaults?: CustomHeaders) =>
	makeDefaultsAwareListPreprocessor(headersSchemaDefaults.nullable(), {
		defaults: defaults,
		matchingFields: ['key', 'value'],
	});
