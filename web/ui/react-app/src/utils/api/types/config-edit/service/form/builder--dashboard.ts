import type { ServiceDashboardOptions } from '@/utils/api/types/config/service/dashboard';
import { serviceDashboardOptionsSchema } from '@/utils/api/types/config-edit/service/types/dashboard';
import { safeParse } from '@/utils/api/types/config-edit/shared/safeparse';
import type { BuilderResponse } from '@/utils/api/types/config-edit/shared/types';
import { applyDefaultsRecursive } from '@/utils/api/types/config-edit/util';

/**
 * Builds a schema for service dashboard options.
 *
 * @param data - The current value from the API.
 * @param defaults - Default values.
 * @param hardDefaults - Hard default values.
 */
export const buildServiceDashboardOptionsSchemaWithFallbacks = (
	data?: ServiceDashboardOptions,
	defaults?: ServiceDashboardOptions,
	hardDefaults?: ServiceDashboardOptions,
): BuilderResponse<typeof serviceDashboardOptionsSchema> => {
	const path = 'dashboard';
	// Service dashboard options schema.
	const schema = serviceDashboardOptionsSchema;

	const fallback = serviceDashboardOptionsSchema.parse({});

	// Initial schema data.
	const schemaData = safeParse({
		data: data,
		fallback: fallback,
		path: path,
		schema: schema,
	});

	// Defaults for the schema.
	const schemaDataDefaults = safeParse({
		data: applyDefaultsRecursive(defaults ?? null, hardDefaults ?? {}),
		fallback: fallback,
		path: `${path} (defaults)`,
		schema: schema,
	});

	return {
		schema: schema,
		schemaData: schemaData,
		schemaDataDefaults: schemaDataDefaults,
	};
};
