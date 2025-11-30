import { strToBool } from '@/utils';
import type { ServiceOptions } from '@/utils/api/types/config/service/options';
import { serviceOptionsSchema } from '@/utils/api/types/config-edit/service/types/options';
import { addZodIssuesToContext } from '@/utils/api/types/config-edit/shared/add-issues';
import { safeParse } from '@/utils/api/types/config-edit/shared/safeparse';
import type { BuilderResponse } from '@/utils/api/types/config-edit/shared/types';
import { applyDefaultsRecursive } from '@/utils/api/types/config-edit/util';

/**
 * Builds a schema for service options.
 *
 * @param data - The current value from the API.
 * @param defaults - Default values.
 * @param hardDefaults - Hard default values.
 */
export const buildServiceOptionsSchemaWithFallbacks = (
	data?: ServiceOptions,
	defaults?: ServiceOptions,
	hardDefaults?: ServiceOptions,
): BuilderResponse<typeof serviceOptionsSchema> => {
	const path = 'options';
	const combinedDefaults = applyDefaultsRecursive(
		defaults ?? null,
		hardDefaults ?? {},
	);

	// Service options schema.
	const schema = serviceOptionsSchema.superRefine((arg, ctx) => {
		const merged = applyDefaultsRecursive(arg, combinedDefaults);
		const result = serviceOptionsSchema.safeParse(merged);

		if (!result.success) {
			addZodIssuesToContext({ ctx, error: result.error });
		}
	});

	// Initial schema data.
	const schemaData = safeParse({
		data: {
			...data,
			active: data?.active !== false,
			semantic_versioning: strToBool(data?.semantic_versioning),
		},
		fallback: {},
		path: path,
		schema: schema,
	});

	// Defaults for the schema.
	const schemaDataDefaults = safeParse({
		data: combinedDefaults,
		fallback: {},
		path: `${path} (defaults)`,
		schema: schema,
	});

	return {
		schema: schema,
		schemaData: schemaData,
		schemaDataDefaults: schemaDataDefaults,
	};
};
