import type { z } from 'zod';
import { isEmptyOrNull } from '@/utils';
import type { ServiceEditOtherData } from '@/utils/api/types/config/defaults';
import type { Service } from '@/utils/api/types/config/service';
import { DEPLOYED_VERSION_LOOKUP_TYPE } from '@/utils/api/types/config/service/deployed-version';
import { buildCommandsSchemaWithFallbacks } from '@/utils/api/types/config-edit/command/form/builder';
import { buildNotifySchemaWithFallbacks } from '@/utils/api/types/config-edit/notify/form/builder';
import { buildServiceDashboardOptionsSchemaWithFallbacks } from '@/utils/api/types/config-edit/service/form/builder--dashboard';
import { buildDeployedVersionLookupSchemaWithFallbacks } from '@/utils/api/types/config-edit/service/form/builder--deployed-version';
import { buildLatestVersionLookupSchemaWithFallbacks } from '@/utils/api/types/config-edit/service/form/builder--latest-version';
import { buildServiceOptionsSchemaWithFallbacks } from '@/utils/api/types/config-edit/service/form/builder--options';
import {
	serviceSchema,
	serviceSchemaDefault,
} from '@/utils/api/types/config-edit/service/schemas';
import { safeParse } from '@/utils/api/types/config-edit/shared/safeparse';
import {
	CUSTOM_ISSUE_CODE,
	REQUIRED_MESSAGE,
	UNIQUE_MESSAGE,
} from '@/utils/api/types/config-edit/validators';
import { buildWebHooksSchemaWithFallbacks } from '@/utils/api/types/config-edit/webhook/form/builder';

/**
 * Builds a schema for the service.
 *
 * @param names - The names of all services.
 * @param ids - The IDs of all services.
 * @param otherOptionsData - The other options data (defaults/hardDefaults and webhook/notify globals).
 * @param data - The current value from the API.
 */
export const buildServiceSchemaWithFallbacks = (
	names: Set<string>,
	ids: string[],
	otherOptionsData?: ServiceEditOtherData,
	data?: Service,
) => {
	// Wait for the 'other options' data.
	if (!otherOptionsData)
		return {
			schema: serviceSchema,
			schemaData: null,
			schemaDataDefaults: null,
		};

	const {
		defaults,
		hard_defaults: hardDefaults,
		webhook,
		notify,
	} = otherOptionsData;

	/* Options */
	const {
		schema: optionsSchema,
		schemaData: optionsSchemaData,
		schemaDataDefaults: optionsSchemaDataDefaults,
	} = buildServiceOptionsSchemaWithFallbacks(
		data?.options,
		defaults?.service?.options,
		hardDefaults?.service?.options,
	);

	/* Latest version */
	const {
		schema: latestVersionSchema,
		schemaData: latestVersionSchemaData,
		schemaDataDefaults: latestVersionSchemaDataDefaults,
	} = buildLatestVersionLookupSchemaWithFallbacks(
		data?.latest_version,
		defaults?.service.latest_version,
		hardDefaults?.service?.latest_version,
	);

	/* Deployed version */
	const {
		schema: deployedVersionSchema,
		schemaData: deployedVersionSchemaData,
		schemaDataDefaults: deployedVersionSchemaDataDefaults,
	} = buildDeployedVersionLookupSchemaWithFallbacks(
		data?.deployed_version,
		defaults?.service?.deployed_version,
		hardDefaults?.service?.deployed_version,
	);

	/* Command */
	const {
		schema: commandSchema,
		schemaData: commandSchemaData,
		schemaDataDefaults: commandSchemaDataDefaults,
		schemaDataDefaultsHollow: commandSchemaDataDefaultsHollow,
	} = buildCommandsSchemaWithFallbacks(
		data?.command,
		defaults?.service?.command,
		hardDefaults?.service?.command,
	);

	/* WebHook */
	const {
		schema: webhookSchema,
		schemaData: webhookSchemaData,
		schemaDataDefaults: webhookSchemaDataDefaults,
		schemaDataMains: webhookMainDataDefaults,
		schemaDataTypeDefaults: webhookTypeDataDefaults,
		schemaDataTypeDefaultsHollow: webhookTypeDataDefaultsHollow,
	} = buildWebHooksSchemaWithFallbacks(
		data?.webhook,
		defaults?.service?.webhook,
		webhook,
		defaults?.webhook,
		hardDefaults?.webhook,
	);

	/* Notify */
	const {
		schema: notifySchema,
		schemaData: notifySchemaData,
		schemaDataDefaults: notifySchemaDataDefaults,
		schemaDataMains: notifyMainDataDefaults,
		schemaDataTypeDefaults: notifyTypeDataDefaults,
	} = buildNotifySchemaWithFallbacks(
		data?.notify,
		defaults?.service?.notify,
		notify,
		defaults?.notify,
		hardDefaults?.notify,
	);

	/* Dashboard */
	const {
		schema: dashboardSchema,
		schemaData: dashboardSchemaData,
		schemaDataDefaults: dashboardSchemaDataDefaults,
	} = buildServiceDashboardOptionsSchemaWithFallbacks(
		data?.dashboard,
		defaults?.service?.dashboard,
		hardDefaults?.service?.dashboard,
	);

	// Build the initial schema data.
	const schemaData = {
		command: commandSchemaData,
		comment: data?.comment ?? '',
		dashboard: dashboardSchemaData,
		deployed_version: deployedVersionSchemaData,
		id: data?.id ?? '',
		id_name_separator: !isEmptyOrNull(data?.name) && !isEmptyOrNull(data?.id),
		latest_version: latestVersionSchemaData,
		name: data?.name ?? '',
		notify: notifySchemaData,

		options: optionsSchemaData,
		webhook: webhookSchemaData,
	};

	// Combine the schemas.
	const schema = serviceSchema
		.extend({
			command: commandSchema,
			dashboard: dashboardSchema,
			deployed_version: deployedVersionSchema,
			latest_version: latestVersionSchema,
			notify: notifySchema,
			options: optionsSchema,
			webhook: webhookSchema,
		})
		.superRefine((arg, ctx) => {
			// Name required if ID name separator in use.
			if (arg.id_name_separator && isEmptyOrNull(arg.name)) {
				ctx.addIssue({
					code: CUSTOM_ISSUE_CODE,
					message: REQUIRED_MESSAGE,
					path: ['name'],
				});
			}

			// If the ID has changed, check not already in use.
			if (
				schemaData.id !== arg.id &&
				schemaData.name !== arg.id &&
				(names.has(arg.id) || ids.includes(arg.id))
			) {
				ctx.addIssue({
					code: CUSTOM_ISSUE_CODE,
					message: UNIQUE_MESSAGE,
					path: ['id'],
				});
			}
			// If the name has changed, check not already in use.
			if (
				arg.name &&
				arg.name !== arg.id &&
				schemaData.id !== arg.name &&
				schemaData.name !== arg.name &&
				(names.has(arg.name) || ids.includes(arg.name))
			) {
				ctx.addIssue({
					code: CUSTOM_ISSUE_CODE,
					message: UNIQUE_MESSAGE,
					path: ['name'],
				});
			}
		});

	// Defaults for the schema.
	const schemaDataDefaults = safeParse({
		data: {
			command: commandSchemaDataDefaults,
			dashboard: dashboardSchemaDataDefaults,
			deployed_version: deployedVersionSchemaDataDefaults,
			latest_version: latestVersionSchemaDataDefaults,
			notify: notifySchemaDataDefaults,
			options: optionsSchemaDataDefaults,
			webhook: webhookSchemaDataDefaults,
		},
		fallback: {
			dashboard: {},
			deployed_version: {
				type: DEPLOYED_VERSION_LOOKUP_TYPE.URL.value,
			},
			notify: [],
		},
		path: 'service (defaults)',
		schema: serviceSchemaDefault,
	});

	// Hollow defaults.
	const schemaDataDefaultsHollow = {
		command: commandSchemaDataDefaultsHollow,
	};

	// Type-specific defaults.
	const typeDataDefaults = {
		notify: notifyTypeDataDefaults,
		webhook: webhookTypeDataDefaults,
	};
	// Hollow type-specific defaults.
	const typeDataDefaultsHollow = {
		webhook: webhookTypeDataDefaultsHollow,
	};
	// 'Main' defaults.
	const mainDataDefaults = {
		notify: notifyMainDataDefaults,
		webhook: webhookMainDataDefaults,
	};

	return {
		mainDataDefaults: mainDataDefaults,
		schema: schema,
		schemaData: schemaData,
		schemaDataDefaults: schemaDataDefaults,
		schemaDataDefaultsHollow: schemaDataDefaultsHollow,
		typeDataDefaults: typeDataDefaults,
		typeDataDefaultsHollow: typeDataDefaultsHollow,
	} satisfies {
		schema: typeof schema;
		schemaData: z.infer<typeof schema> | null;
		schemaDataDefaults: z.infer<typeof serviceSchemaDefault> | null;
		schemaDataDefaultsHollow: typeof schemaDataDefaultsHollow;
		typeDataDefaults: typeof typeDataDefaults;
		typeDataDefaultsHollow: typeof typeDataDefaultsHollow;
		mainDataDefaults: typeof mainDataDefaults;
	};
};
