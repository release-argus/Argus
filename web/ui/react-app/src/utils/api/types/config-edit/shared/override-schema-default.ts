import type { output, ZodDefault, ZodType } from 'zod';

type NonUndefinedOutput<T extends ZodType> = Exclude<output<T>, undefined>;

/**
 * Overrides the default value of a Zod schema.
 *
 * @param schema - The schema to override the default value of.
 * @param newDefault - The new default value to set.
 * @returns A new schema with the overridden default value.
 */
export const overrideSchemaDefault = <T extends ZodType>(
	schema: ZodDefault<T> | T,
	newDefault: NonUndefinedOutput<T>,
): ZodDefault<T> => {
	// Unwrap if this schema already has a default.
	const baseSchema = 'unwrap' in schema ? schema.unwrap() : schema;

	return baseSchema.default(() => newDefault);
};
