import type { z } from 'zod';
import type { NotifySchemaValues } from '@/utils/api/types/config-edit/notify/schemas';
import {
	type FieldValidation,
	validateFields,
	validateMainTypeMatch,
} from '@/utils/api/types/config-edit/validators';
import type { WebHookSchema } from '@/utils/api/types/config-edit/webhook/schemas';

/**
 * Builds a superRefine function for a schema.
 *
 * @param schema - The schema to build the superRefine function for.
 * @param mains - The 'main' objects the schema can reference.
 * @param defaults - Default values for the schema.
 * @param fieldValidations - Field validations for the schema.
 */
export const buildSuperRefine = <T extends NotifySchemaValues | WebHookSchema>(
	schema: z.ZodType<T>,
	mains: Record<string, T>,
	defaults: T,
	fieldValidations: FieldValidation[],
) =>
	schema.superRefine((arg, ctx) => {
		const mainValues = (mains[arg.name] as T | undefined) ?? null;
		validateMainTypeMatch(arg, mainValues, ctx);
		validateFields(arg, mainValues, defaults, fieldValidations, ctx);
	});
