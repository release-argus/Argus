import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import {
	LATEST_VERSION__URL_COMMAND_TYPE,
	LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE,
	LATEST_VERSION_LOOKUP_TYPE,
	type LatestVersionLookupType,
	latestVersionLookupTypeOptions,
} from '@/utils/api/types/config/service/latest-version';
import { commandSchema } from '@/utils/api/types/config-edit/command/schemas';
import { zodStringToNumber } from '@/utils/api/types/config-edit/shared/number-string';
import {
	NUMBER_REQUIRED_MESSAGE,
	REQUIRED_MESSAGE,
	regexStringWithFallback,
} from '@/utils/api/types/config-edit/validators';

export const LatestVersionTypeEnum = z.enum(
	toZodEnumTuple(latestVersionLookupTypeOptions),
);

/* url_command: 'regex' | 'replace' | 'split' */

/* type: 'regex' */
export const urlCommandRegexSchema = z.object({
	index: zodStringToNumber(z.number().optional()),
	regex: regexStringWithFallback(true),
	template: z.string().default(''),
	template_toggle: z.boolean(),
	type: z.literal(LATEST_VERSION__URL_COMMAND_TYPE.REGEX.value),
});
export const urlCommandRegexSchemaDefault = urlCommandRegexSchema
	.extend({
		regex: z.string().default(''),
		template_toggle: z.boolean().default(false),
	})
	.transform((data) => ({
		...data,
		// template_toggle starts true if `template` not empty.
		template_toggle: !!data.template,
	}));
const urlCommandRegexSchemaOutgoing = z.object({
	index: zodStringToNumber(z.number().optional()),
	regex: regexStringWithFallback(true),
	template: z.string().default(''),
	type: z.literal(LATEST_VERSION__URL_COMMAND_TYPE.REGEX.value),
});

/* type: 'replace' */
export const urlCommandReplaceSchema = z.object({
	new: z.string().default(''),
	old: z.string().min(1, REQUIRED_MESSAGE),
	type: z.literal(LATEST_VERSION__URL_COMMAND_TYPE.REPLACE.value),
});
export const urlCommandReplaceSchemaDefault = urlCommandReplaceSchema.extend({
	old: z.string().default(''),
});

/* type: 'split' */
export const urlCommandSplitSchema = z.object({
	index: zodStringToNumber(
		z.number({
			error: (issue) =>
				issue.input ? NUMBER_REQUIRED_MESSAGE : REQUIRED_MESSAGE,
		}),
	),
	text: z.string().min(1, REQUIRED_MESSAGE),
	type: z.literal(LATEST_VERSION__URL_COMMAND_TYPE.SPLIT.value),
});
export const urlCommandSplitSchemaDefault = urlCommandSplitSchema.extend({
	index: z.union([z.string(), z.number()]).default(''),
	text: z.string().default(''),
});

/* url_command */
export const urlCommandSchemaDefaults = z.discriminatedUnion('type', [
	urlCommandRegexSchemaDefault,
	urlCommandReplaceSchemaDefault,
	urlCommandSplitSchemaDefault,
]);
export type URLCommand = z.infer<typeof urlCommandSchemaDefaults>;
export const urlCommandsSchemaDefaults = z
	.array(urlCommandSchemaDefaults)
	.default([]);

export const urlCommandSchema = z.discriminatedUnion('type', [
	urlCommandRegexSchema,
	urlCommandReplaceSchema,
	urlCommandSplitSchema,
]);
export const urlCommandsSchema = z.array(urlCommandSchema).default([]);
export type URLCommandsSchema = z.infer<typeof urlCommandsSchema>;

export const urlCommandMap = {
	regex: urlCommandRegexSchemaDefault,
	replace: urlCommandReplaceSchemaDefault,
	split: urlCommandSplitSchemaDefault,
};

const urlCommandSchemaOutgoing = z.discriminatedUnion('type', [
	urlCommandRegexSchemaOutgoing,
	urlCommandReplaceSchema,
	urlCommandSplitSchema,
]);
export const urlCommandsSchemaOutgoing = z
	.array(urlCommandSchemaOutgoing)
	.nullable()
	.default(null);

/* require.docker */
const dockerFilterSchemaBase = [
	LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.DOCKER_HUB.value,
	LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.GHCR.value,
	LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.QUAY.value,
] as const;

export const latestVersionLookupRequireDockerTypeSchemaBase = z.object({
	image: z.string().default(''),
	tag: z.string().default(''),
	token: z.string().default(''),
});
export const latestVersionLookupRequireDockerTypeSchema =
	latestVersionLookupRequireDockerTypeSchemaBase.extend({
		type: z.literal([
			LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.GHCR.value,
			LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.QUAY.value,
		]),
	});

export const latestVersionLookupRequireDockerTypeSchemaDockerHubBase =
	latestVersionLookupRequireDockerTypeSchemaBase.extend({
		username: z.string().default(''),
	});
export const latestVersionLookupRequireDockerTypeSchemaDockerHub =
	latestVersionLookupRequireDockerTypeSchemaDockerHubBase.extend({
		type: z.literal(
			LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.DOCKER_HUB.value,
		),
	});

export const dockerFilterSchema = z.discriminatedUnion('type', [
	latestVersionLookupRequireDockerTypeSchema,
	latestVersionLookupRequireDockerTypeSchemaDockerHub,
]);
export type DockerTypeDockerHub = z.infer<
	typeof latestVersionLookupRequireDockerTypeSchemaDockerHub
>;
export const dockerFilterSchemaDefaults = z.object({
	ghcr: latestVersionLookupRequireDockerTypeSchemaBase.optional(),

	hub: latestVersionLookupRequireDockerTypeSchemaDockerHubBase.optional(),
	quay: latestVersionLookupRequireDockerTypeSchemaBase.optional(),
	type: z.literal(dockerFilterSchemaBase).optional(),
});

/* require */
export const latestVersionRequireSchema = z.object({
	command: commandSchema.default([]),
	docker: dockerFilterSchema,
	regex_content: z.string().default(''),
	regex_version: z.string().default(''),
});
export type LatestVersionRequire = z.infer<typeof latestVersionRequireSchema>;
export const latestVersionRequireSchemaDefaults =
	latestVersionRequireSchema.extend({
		docker: dockerFilterSchemaDefaults.optional(),
	});

export const latestVersionLookupSchemaBase = z.object({
	require: latestVersionRequireSchema,
	url_commands: urlCommandsSchemaDefaults,
});

/* Require assets from GitHub. */
export const latestVersionLookupSchemaGitHub =
	latestVersionLookupSchemaBase.extend({
		access_token: z.string().default(''),
		allow_invalid_certs: z.boolean().nullable().default(null),
		type: z.literal(LATEST_VERSION_LOOKUP_TYPE.GITHUB.value),
		url: z.string().default(''),
		use_prerelease: z.boolean().nullable().default(null),
	});
/* Require assets from the web page. */
export const latestVersionLookupSchemaURL =
	latestVersionLookupSchemaBase.extend({
		allow_invalid_certs: z.boolean().nullable().default(null),
		type: z.literal(LATEST_VERSION_LOOKUP_TYPE.URL.value),
		url: z.string(),
	});

export const latestVersionLookupSchema = z.discriminatedUnion('type', [
	latestVersionLookupSchemaGitHub,
	latestVersionLookupSchemaURL,
]);

export type LatestVersionLookupSchema = z.infer<
	typeof latestVersionLookupSchema
>;

export const isLatestVersionType = (
	value?: string | null,
): value is LatestVersionLookupType =>
	value != null &&
	latestVersionLookupTypeOptions.some((v) => v.value === value);

export const latestVersionLookupSchemaDefault = z
	.object({
		access_token: z.string().optional(),
		allow_invalid_certs: z.boolean().nullable().optional(),
		require: latestVersionRequireSchemaDefaults.optional(),
		type: LatestVersionTypeEnum.nullable().optional(),
		url: z.string().optional(),
		url_commands: urlCommandsSchemaDefaults.optional(),
		use_prerelease: z.boolean().nullable().optional(),
	})
	.optional();
export type LatestVersionLookupSchemaDefault = z.infer<
	typeof latestVersionLookupSchema
>;

export const latestVersionLookupSchemaOutgoing = z.discriminatedUnion('type', [
	latestVersionLookupSchemaGitHub.extend({
		url_commands: urlCommandsSchemaOutgoing,
	}),
	latestVersionLookupSchemaURL,
]);
