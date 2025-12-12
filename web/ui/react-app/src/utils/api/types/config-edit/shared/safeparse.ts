import { toast } from 'sonner';
import type { ZodType, z } from 'zod';

type SafeParseParams<T extends ZodType> = {
	/* Schema to parse with. */
	schema: T;
	/* Data to parse. */
	data: unknown;
	/* Path to use in error messages. */
	path: string;
	/* Fallback value if parsing fails. */
	fallback: z.input<T>;
	/* Whether to show toasts/errors. */
	showErrors?: boolean;
};

/**
 * SafeParse a schema with a fallback value.
 *
 * @param schema - The schema to parse.
 * @param data - The data to parse.
 * @param path - The path to use in error messages.
 * @param fallback - The fallback value if parsing fails.
 * @param showErrors - Whether to show toasts/errors.
 */
export const safeParse = <T extends ZodType>({
	schema,
	data,
	path,
	fallback,
	showErrors = true,
}: SafeParseParams<T>): z.output<T> => {
	const parsedSchema = schema.safeParse(data);
	if (parsedSchema.success) {
		return parsedSchema.data;
	}

	if (showErrors) {
		console.error(
			`Failed to parse schema data for ${path}.`,
			'data=',
			data,
			'error=',
			parsedSchema.error,
		);
		toast.error(`Failed to parse ${path}.`, {
			description: parsedSchema.error.message,
			duration: 30000, // 30 seconds.
		});
	}
	return fallback as z.output<T>;
};
