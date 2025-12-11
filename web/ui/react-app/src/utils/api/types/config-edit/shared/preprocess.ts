import { type ZodEnum, z } from 'zod';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';
import { isUsingDefaults } from '@/utils/api/types/config-edit/validators';

/**
 * Parses a JSON string into an array and validates items with the given Zod schema.
 *
 * @template T - Schema for array items.
 * @param schema - Zod schema for each array item.
 * @returns Zod schema parsing strings as JSON arrays, defaulting to an empty array.
 */
export const preprocessArrayJSONFromString = <T extends z.ZodType>(schema: T) =>
	z.preprocess((arg) => {
		if (typeof arg === 'string') {
			try {
				return JSON.parse(arg) as unknown;
			} catch {
				return arg; // zod validation fail
			}
		}

		return arg;
	}, z.array(schema).default([]));

/**
 * Converts a string to boolean.
 * - 'true', 'yes', 'y' => true
 * - false otherwise.
 *
 * @returns Zod schema parsing strings to boolean/nullable, defaulting to null.
 */
export const preprocessBooleanFromString = z.preprocess((arg) => {
	if (typeof arg === 'string') {
		try {
			const s = arg.trim().toLowerCase();
			return ['true', 'yes', 'y'].includes(s);
		} catch {
			return arg; // zod validation fail
		}
	}

	return arg;
}, z.boolean().nullable().default(null));

export const preprocessNumberFromString = z.preprocess((arg) => {
	if (typeof arg === 'string') {
		if (arg.trim() === '') return null;
		try {
			return Number(arg);
		} catch {
			return arg;
		}
	}

	return arg;
}, z.number().nullable().default(null));

/**
 * Converts a boolean to a string.
 *
 * @returns Zod schema parsing booleans as strings, defaulting to ''.
 */
export const preprocessStringFromBoolean = z.preprocess((arg) => {
	if (typeof arg === 'boolean') return String(arg);

	return arg;
}, z.string().nullable().default(''));

/**
 * Converts a number to a string.
 *
 * @returns Zod schema parsing numbers as strings, defaulting to an empty string.
 */
export const preprocessStringFromNumber = z.preprocess((arg) => {
	if (typeof arg === 'number') return String(arg);

	return arg;
}, z.string().default(''));

/**
 * Validates that a value is a member of a Zod enum and converts it to a string.
 *
 * @param enumSchema - The ZodEnum schema to validate against.
 * @returns Zod schema parsing valid enum values as strings, optional.
 */
export const preprocessStringFromZodEnum = (enumSchema: ZodEnum) =>
	z.preprocess((val: unknown) => {
		if (val == null || val === nullString) return undefined;

		const enumValues = enumSchema.options;
		if (enumValues.includes(val as string)) return val;

		return undefined;
	}, z.string().optional());

/**
 * Factory to make a defaults-aware list preprocessor.
 * - If value is an empty array, returns null.
 * - If defaults exist and value matches defaults via matchingFields, returns null.
 * - Otherwise delegates to the given inner preprocessor.
 */
export const makeDefaultsAwareListPreprocessor = (
	inner: z.ZodType,
	opts: {
		matchingFields: string[];
		defaults?: Record<string, unknown>[];
	},
) =>
	z.preprocess((val: unknown) => {
		// Treat empty arrays as using defaults.
		if (Array.isArray(val) && val.length === 0) return null;

		const defaultsVal = opts.defaults;
		if (defaultsVal && Array.isArray(val)) {
			// Compare with defaults using provided matching fields.
			if (
				isUsingDefaults({
					arg: val,
					defaultValue: defaultsVal ?? [],
					matchingFieldsStartsWiths: opts.matchingFields,
				})
			) {
				return null;
			}
		}
		return val;
	}, inner);
