import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import {
	DEPLOYED_VERSION_LOOKUP__URL_METHOD,
	DEPLOYED_VERSION_LOOKUP_TYPE,
	type DeployedVersionLookupType,
	type DeployedVersionLookupURLMethod,
	deployedVersionLookupTypeOptions,
} from '@/utils/api/types/config/service/deployed-version';
import { headersSchema } from '@/utils/api/types/config-edit/shared/header/preprocess';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';
import { regexStringWithFallback } from '@/utils/api/types/config-edit/validators';

/* Type: manual */

export const deployedVersionManualSchema = z.object({
	type: z.literal(DEPLOYED_VERSION_LOOKUP_TYPE.MANUAL.value),
	version: z.string().default(''),
});

const deployedVersionManualSchemaDefault = deployedVersionManualSchema.extend({
	version: z.string().default(''),
});

/* Type: url */

/* basic_auth */
const basicAuthSchema = z
	.object({
		password: z.string().default(''),
		username: z.string().default(''),
	})
	.default({ password: '', username: '' });

/* deployed_version.method */
export const DeployedVersionURLMethodEnum = z.enum(
	toZodEnumTuple(Object.values(DEPLOYED_VERSION_LOOKUP__URL_METHOD)),
);
export type DeployedVersionURLMethod = z.infer<
	typeof DeployedVersionURLMethodEnum
>;

/* deployed_version */
export const deployedVersionURLSchema = z.object({
	allow_invalid_certs: z.boolean().nullable().default(null),
	basic_auth: basicAuthSchema,
	body: z.string().default(''),
	headers: headersSchema,
	json: z.string().default(''),
	method: DeployedVersionURLMethodEnum.or(z.literal(nullString)).default(
		nullString,
	),
	regex: regexStringWithFallback(false),
	regex_template: z.string().default(''),
	target_header: z.string().default(''),
	template_toggle: z.boolean().default(false),
	type: z.literal(DEPLOYED_VERSION_LOOKUP_TYPE.URL.value),
	url: z.string().default(''),
});
export type DeployedVersionURLSchema = z.infer<typeof deployedVersionURLSchema>;

export const isDeployedVersionURLMethod = (
	value?: string | null,
): value is DeployedVersionLookupURLMethod =>
	value != null &&
	DeployedVersionURLMethodEnum.options.includes(
		value as DeployedVersionLookupURLMethod,
	);

const deployedVersionURLSchemaDefault = deployedVersionURLSchema.extend({
	method: DeployedVersionURLMethodEnum.nullable().default(null),
	url: z.string().default(''),
});

/* All */
export const deployedVersionLookupSchema = z.discriminatedUnion('type', [
	deployedVersionManualSchema,
	deployedVersionURLSchema,
]);

export type DeployedVersionLookupSchema = z.infer<
	typeof deployedVersionLookupSchema
>;

export const isDeployedVersionType = (
	value?: string | null,
): value is DeployedVersionLookupType =>
	value != null &&
	deployedVersionLookupTypeOptions.some((v) => v.value === value);

export const deployedVersionLookupSchemaDefault = z.discriminatedUnion('type', [
	deployedVersionManualSchemaDefault,
	deployedVersionURLSchemaDefault,
]);
export type DeployedVersionLookupSchemaDefault = z.infer<
	typeof deployedVersionLookupSchemaDefault
>;
