import { z } from 'zod';
import { NUMBER_REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators';

type OptionalStringInput =
	| z.ZodNumber
	| z.ZodOptional<z.ZodNumber>
	| z.ZodPipe<z.ZodNumber, z.ZodTransform<number, number>>;

type ZodStringToNumber<P extends OptionalStringInput> = z.ZodPipe<
	z.ZodPipe<
		z.ZodOptional<
			z.ZodPipe<z.ZodString, z.ZodTransform<string | undefined, string>>
		>
	>,
	P
>;

/**
 * Converts a Zod schema that accepts strings or numbers into a number, handling empty strings and validation.
 *
 * @template P - A Zod schema to pipe the transformed number into.
 * @param zodPipe - The Zod schema to apply after converting the input to a number.
 * @returns A Zod schema that:
 *   - Accepts string or number inputs,
 *   - Converts empty strings to `undefined`,
 *   - Validates numeric values,
 *   - Transforms valid inputs to numbers,
 *   - Applies the given `zodPipe` schema.
 */
export const zodStringToNumber = <P extends OptionalStringInput>(zodPipe: P) =>
	z
		.union([z.string(), z.number()])
		.transform((value) => (value === '' ? undefined : value))
		.optional()
		.refine(
			(value) => value === undefined || !Number.isNaN(Number(value)),
			NUMBER_REQUIRED_MESSAGE,
		)
		.transform((value) => (value === undefined ? undefined : Number(value)))
		.pipe(zodPipe) as unknown as ZodStringToNumber<P>;
