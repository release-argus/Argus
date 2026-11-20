import { type ZodArray, type ZodDefault, type ZodPreprocess, z } from 'zod';
import type { Headers } from '@/utils/api/types/config/shared';
import {
	headerSchema,
	headerSchemaWithValidation,
	mapEntrySchema,
	mapEntrySchemaWithValidation,
} from '@/utils/api/types/config-edit/shared/header/schemas';
import { makeDefaultsAwareListPreprocessor } from '@/utils/api/types/config-edit/shared/preprocess';
import { SECRET_VALUE } from '@/utils/api/types/config-edit/validators';

type MapEntry = { key: string; old_key?: string | null; value: string };
type RenamedSecret = { old_key: string; value: string };

/**
 * Returns the value to send for a JSON object entry.
 *
 * @param key - The entry's key.
 * @param oldKey - The key the entry loaded with.
 * @param value - The entry's value.
 * @returns `{old_key, value}` when renamed with its secret masked, else the value.
 */
export const renamedSecret = (
	key: string,
	oldKey: string | null | undefined,
	value: string,
): string | RenamedSecret =>
	value === SECRET_VALUE && oldKey && oldKey !== key
		? { old_key: oldKey, value }
		: value;

/**
 * Flattens JSON object entries for the API.
 * A renamed entry still holding the secret placeholder carries its old key, so the API restores it.
 *
 * @param entries - The entries to flatten {key: "KEY", value: "VAL", old_key: "OLD"}[].
 * @returns The flattened object {KEY: VAL, ...}, or undefined without entries.
 */
export const flattenMapEntries = (entries?: MapEntry[]) =>
	entries?.reduce<Record<string, string | RenamedSecret>>(
		(obj, { key, old_key: oldKey, value }) => {
			obj[key] = renamedSecret(key, oldKey, value);
			return obj;
		},
		{},
	);

/**
 * Converts map entries to a JSON object string.
 * An entry renamed while its value is masked becomes `{old_key, value}`, so the API can restore it.
 *
 * @param val - The map entries to convert.
 * @returns A JSON string of the entries.
 */
export const preprocessStringFromMapEntries = z.preprocess((val: unknown) => {
	if (!val || !Array.isArray(val) || val.length === 0) return '';

	const flattened = flattenMapEntries(val);
	// Using defaults if any key empty.
	if (flattened && Object.keys(flattened).some((arg) => !arg.trim())) return '';

	return JSON.stringify(flattened);
}, z.string());

/**
 * Defaults-aware variant of the map entries -> string preprocessor.
 * - Empty array -> null
 * - Matches defaults (deep, via matchingFields=[]) -> null
 */
export const preprocessStringFromMapEntriesWithDefaults = (
	defaults?: Record<string, unknown>[],
) =>
	makeDefaultsAwareListPreprocessor(preprocessStringFromMapEntries.nullable(), {
		defaults: defaults,
		matchingFields: [],
	});

/**
 * Converts various header representations into an array of header objects.
 * Accepts a Zod schema and returns a Zod array schema with preprocessing logic for strings and objects.
 *
 * @template T - A ZodType representing the header schema.
 * @param schema - The Zod schema representing the headers.
 */
export const preprocessToHeadersArray = <T extends z.ZodType>(
	schema: T,
): ZodDefault<ZodPreprocess<ZodArray<T>>> =>
	z
		.preprocess((arg) => {
			// If string, try to parse as JSON.
			if (typeof arg === 'string') {
				const parsed = JSON.parse(arg) as unknown;
				if (parsed && typeof parsed === 'object') {
					return Object.entries(parsed).map(([key, value]) => ({
						key,
						old_key: key,
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
					old_key: key,
					value: value,
				}));
			}

			// Unknown type, return as is.
			return arg;
		}, z.array(schema))
		.default([]);

/* Array of Header objects (min length 1 on key and value) */
export const headersSchemaWithValidation = preprocessToHeadersArray(
	headerSchemaWithValidation,
);
/* Array of Header objects (no validation) */
export const headersSchema = preprocessToHeadersArray(headerSchema);
/* Array of JSON object entries (min length 1 on key and value) */
export const mapEntriesSchemaWithValidation = preprocessToHeadersArray(
	mapEntrySchemaWithValidation,
);
/* Array of JSON object entries (no validation) */
export const mapEntriesSchema = preprocessToHeadersArray(mapEntrySchema);

export const preprocessHeadersToHeadersSchema = (defaults?: Headers) =>
	makeDefaultsAwareListPreprocessor(headersSchema.nullable(), {
		defaults: defaults,
		matchingFields: ['key', 'value'],
	});
