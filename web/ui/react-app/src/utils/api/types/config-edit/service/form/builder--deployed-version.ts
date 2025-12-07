import { z } from 'zod';
import {
	DEPLOYED_VERSION_LOOKUP_TYPE,
	type DeployedVersionLookup,
	deployedVersionLookupTypeOptions,
} from '@/utils/api/types/config/service/deployed-version';
import {
	type deployedVersionLookupSchema,
	deployedVersionManualSchema,
	deployedVersionURLSchema,
	isDeployedVersionType,
} from '@/utils/api/types/config-edit/service/types/deployed-version';
import { buildHeadersSchemaWithFallbacks } from '@/utils/api/types/config-edit/shared/header/builder';
import { safeParse } from '@/utils/api/types/config-edit/shared/safeparse';
import type { BuilderResponse } from '@/utils/api/types/config-edit/shared/types';
import { applyDefaultsRecursive } from '@/utils/api/types/config-edit/util';
import {
	CUSTOM_ISSUE_CODE,
	REQUIRED_MESSAGE,
	stringWithFallback,
	urlPrefixValidator,
} from '@/utils/api/types/config-edit/validators';

/**
 * Builds a schema for deployed version lookup.
 *
 * @param data - The current value from the API.
 * @param defaults - Default values.
 * @param hardDefaults - Hard default values.
 */
export const buildDeployedVersionLookupSchemaWithFallbacks = (
	data?: DeployedVersionLookup,
	defaults?: DeployedVersionLookup,
	hardDefaults?: DeployedVersionLookup,
): BuilderResponse<typeof deployedVersionLookupSchema> => {
	const path = 'deployed_version';
	const combinedDefaults = applyDefaultsRecursive<DeployedVersionLookup>(
		defaults ?? null,
		hardDefaults,
	);

	// Manual schema.
	const dvManualSchema = deployedVersionManualSchema;

	// URL schema.
	const {
		schema: headersSchema,
		schemaData: headersSchemaData,
		schemaDataDefaults: headersSchemaDataDefaults,
	} = buildHeadersSchemaWithFallbacks(
		data && 'headers' in data ? data.headers : [],
		'headers' in combinedDefaults ? combinedDefaults.headers : [],
	);
	const dvURLSchema = deployedVersionURLSchema
		.extend({
			headers: headersSchema,
			url: stringWithFallback(
				urlPrefixValidator,
				false,
				'url' in combinedDefaults ? combinedDefaults.url : undefined,
			),
		})
		.superRefine((arg, ctx) => {
			// `regex` required if `template_toggle` is `true`.
			if (arg.template_toggle && !arg.regex) {
				ctx.addIssue({
					code: CUSTOM_ISSUE_CODE,
					message: REQUIRED_MESSAGE,
					path: ['regex'],
				});
			}
		});

	// Deployed version schema.
	const schemaRaw = z.discriminatedUnion('type', [
		deployedVersionManualSchema,
		deployedVersionURLSchema.transform((data) => ({
			...data,
			// template_toggle starts true if `regex_template` not empty.
			template_toggle: data.template_toggle || !!data.regex_template,
		})),
	]);
	const schema = z.discriminatedUnion('type', [dvManualSchema, dvURLSchema]);

	// Initial type/method.
	const fallbackType = Object.values(deployedVersionLookupTypeOptions)[1].value;
	const schemaDataType = isDeployedVersionType(data?.type)
		? data.type
		: fallbackType;
	// Initial schema data.
	const fallbackData = {
		headers: headersSchemaData,
		type: schemaDataType,
	};
	if (!isDeployedVersionType(fallbackData.type))
		fallbackData.type = fallbackType;
	const schemaData = safeParse({
		data: {
			allow_invalid_certs: null,
			...data,
			...fallbackData,
		},
		fallback: fallbackData,
		path: path,
		schema: schemaRaw,
	});

	// Defaults for the schema.
	const schemaDataDefaults = safeParse({
		data: {
			...combinedDefaults,
			headers: headersSchemaDataDefaults,
			type: DEPLOYED_VERSION_LOOKUP_TYPE.URL.value,
		},
		fallback: {
			headers: headersSchemaDataDefaults,
			type: fallbackData.type,
		},
		path: `${path} (defaults)`,
		schema: schemaRaw,
	});

	return {
		schema: schema,
		schemaData: schemaData,
		schemaDataDefaults: schemaDataDefaults,
	};
};
