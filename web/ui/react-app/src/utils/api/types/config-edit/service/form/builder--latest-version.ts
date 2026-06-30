import { z } from 'zod';
import { urlCommandsTrimArray } from '@/components/modals/service-edit/util';
import type { NonNull } from '@/types/util';
import {
	type DockerFilter,
	type DockerFilterType,
	type DockerFilterUsername,
	LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE,
	LATEST_VERSION_LOOKUP_TYPE,
	type LatestVersionLookup,
	type LatestVersionLookupDefaults,
	type LatestVersionLookupGitHub,
	type LatestVersionLookupType,
	type LatestVersionLookupURL,
	type LatestVersionRequire,
	type LatestVersionRequireDefaults,
	type RequireDockerFilterDefaults,
	type URLCommand,
} from '@/utils/api/types/config/service/latest-version';
import { buildCommandSchemaWithFallbacks } from '@/utils/api/types/config-edit/command/form/builder';
import {
	dockerFilterSchema,
	dockerFilterSchemaDefaults,
	isLatestVersionType,
	latestVersionLookupRequireDockerTypeSchema,
	latestVersionLookupRequireDockerTypeSchemaDockerHub,
	type latestVersionLookupSchema,
	latestVersionLookupSchemaDefault,
	latestVersionLookupSchemaGitHub,
	latestVersionLookupSchemaURL,
	latestVersionRequireSchema,
	latestVersionRequireSchemaDefaults,
	urlCommandsSchema,
	urlCommandsSchemaWithValidation,
} from '@/utils/api/types/config-edit/service/types/latest-version';
import { addZodIssuesToContext } from '@/utils/api/types/config-edit/shared/add-issues';
import { buildHeadersSchemaWithFallbacks } from '@/utils/api/types/config-edit/shared/header/builder';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';
import { stringDefault } from '@/utils/api/types/config-edit/shared/preprocess';
import { safeParse } from '@/utils/api/types/config-edit/shared/safeparse';
import type { BuilderResponse } from '@/utils/api/types/config-edit/shared/types';
import { applyDefaultsRecursive } from '@/utils/api/types/config-edit/util';
import {
	CUSTOM_ISSUE_CODE,
	REQUIRED_MESSAGE,
	regexStringWithFallback,
	safeParseListWithSchemas,
	validateGitHubRepo,
	validateRequired,
	validateURL,
} from '@/utils/api/types/config-edit/validators';

/**
 * Builds a schema for 'latest version lookup' URL commands.
 *
 * @param data - The current value from the API.
 * @param defaults - Default values.
 * @param hardDefaults - Hard default values.
 */
export const buildURLCommandsSchemaWithFallbacks = (
	data?: URLCommand[],
	defaults?: URLCommand[],
	hardDefaults?: URLCommand[],
) => {
	const path = 'latest_version.url_commands';
	const defaultItem = defaults ?? hardDefaults ?? [];
	const defaultValue = urlCommandsTrimArray(defaultItem);

	// Schema.
	const schema = urlCommandsSchema;
	const schemaFinal = schema.superRefine((arg, ctx) => {
		if (arg.length === 0) return;

		const { result } = safeParseListWithSchemas({
			arg: arg,
			defaultSchema: urlCommandsSchema,
			defaultValue: defaultValue,
			matchingFields: ['type'],
			schema: urlCommandsSchemaWithValidation,
		});

		if (!result.success) {
			addZodIssuesToContext({ ctx: ctx, error: result.error });
		}
	});

	// Initial schema data.
	const schemaData = safeParse({
		data: data ?? defaultValue,
		fallback: [],
		path: path,
		schema: schema,
	});

	// Defaults for the schema.
	const schemaDataDefaults = safeParse({
		data: defaultValue,
		fallback: [],
		path: `${path} (defaults)`,
		schema: schema,
	});

	return {
		schema: schemaFinal,
		schemaData: schemaData,
		schemaDataDefaults: schemaDataDefaults,
	};
};

/* require.docker */
/**
 * Builds a schema for the 'latest version' require.docker.
 *
 * @param data - The current value from the API.
 * @param defaults - Default values.
 * @param hardDefaults - Hard default values.
 */
export const buildDockerFilterSchemaWithFallbacks = (
	data?: DockerFilter,
	defaults?: RequireDockerFilterDefaults,
	hardDefaults?: RequireDockerFilterDefaults,
): BuilderResponse<
	typeof dockerFilterSchema,
	typeof dockerFilterSchemaDefaults
> => {
	const path = 'latest_version.require.docker';
	const defaultType = defaults?.type || hardDefaults?.type || undefined;

	const dockerHubValue =
		LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.DOCKER_HUB.value;
	const ghcrValue = LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.GHCR.value;
	const quayValue = LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.QUAY.value;

	const combinedDefaults = {
		registry: [dockerHubValue, ghcrValue, quayValue].reduce(
			(acc, type) => {
				acc[type] = applyDefaultsRecursive<Partial<DockerFilter>>(
					(defaults?.registry?.[type] as Partial<DockerFilter>) ?? null,
					{ tag: defaults?.tag },
					hardDefaults?.registry?.[type] as Partial<DockerFilter> | undefined,
					{ tag: hardDefaults?.tag },
				);
				return acc;
			},
			{} as Record<NonNullable<DockerFilterType>, Partial<DockerFilter>>,
		),
		type: defaultType,
	};

	// Docker registries that support username with tokens.
	const usernameTypes = new Set<DockerFilterType>([dockerHubValue]);

	// Docker schema.
	const schema = z.preprocess(
		(data) => {
			// Convert {type: null} to {type: nullString}.
			if (
				data &&
				typeof data === 'object' &&
				'type' in data &&
				data.type === null
			) {
				return { ...data, type: nullString };
			}
			return data;
		},
		z.discriminatedUnion('type', [
			latestVersionLookupRequireDockerTypeSchemaDockerHub.extend({
				type:
					defaultType === dockerHubValue
						? z.literal([dockerHubValue, nullString])
						: z.literal(dockerHubValue),
			}),
			latestVersionLookupRequireDockerTypeSchema.extend({
				type:
					defaultType === ghcrValue
						? z.literal([ghcrValue, nullString])
						: z.literal(ghcrValue),
			}),
			latestVersionLookupRequireDockerTypeSchema.extend({
				type:
					defaultType === quayValue
						? z.literal([quayValue, nullString])
						: z.literal(quayValue),
			}),
		]),
	);

	// Add validation for required fields.
	const schemaFinal = schema.superRefine((arg, ctx) => {
		const schemaType = arg.type === nullString ? defaultType : arg.type;
		const schemaDefaults = schemaType && combinedDefaults.registry[schemaType];
		const hasImage = !!arg.image?.trim();
		const hasTag = !!arg.tag?.trim();
		const hasTagDefaulted = hasTag || !!schemaDefaults?.tag?.trim();

		// If we have an image/tag defined for this instance, we must have values for both.
		if (hasImage !== hasTag && hasImage !== hasTagDefaulted) {
			ctx.addIssue({
				code: CUSTOM_ISSUE_CODE,
				message: REQUIRED_MESSAGE,
				path: hasImage ? ['tag'] : ['image'],
			});
		}

		// If we have an image:tag specified and have a username field.
		if (
			hasImage &&
			hasTagDefaulted &&
			schemaType &&
			usernameTypes.has(schemaType)
		) {
			type DockerUsernameTyped = z.infer<
				typeof latestVersionLookupRequireDockerTypeSchemaDockerHub
			>;
			const argTyped = arg as DockerUsernameTyped;
			const hasUsername = !!(
				argTyped.auth?.username ||
				(schemaDefaults as Partial<DockerFilterUsername>).auth?.username
			)?.trim();
			const hasToken = !!(
				arg.auth?.token || schemaDefaults?.auth?.token
			)?.trim();

			// We must have a username and token, or neither.
			if (hasUsername !== hasToken) {
				ctx.addIssue({
					code: CUSTOM_ISSUE_CODE,
					message: REQUIRED_MESSAGE,
					path: hasUsername ? ['auth', 'token'] : ['auth', 'username'],
				});
			}
		}
		// 	'unknown' since we can't have dynamic {type: nullString, X} in the discriminated union.
	}) as unknown as typeof dockerFilterSchema;

	// Initial schema data.
	const schemaData = safeParse({
		data: {
			...data,
			type: data?.type ?? null,
		},
		fallback: { type: defaultType as NonNull<DockerFilterType> },
		path: path,
		schema: schemaFinal,
	});

	// Defaults for the schema.
	const schemaDataDefaults = safeParse({
		data: combinedDefaults,
		fallback: { type: defaultType },
		path: `${path} (defaults)`,
		schema: dockerFilterSchemaDefaults,
	});

	return {
		schema: schemaFinal,
		schemaData: schemaData,
		schemaDataDefaults: schemaDataDefaults,
	};
};

/* require */
/**
 * Builds a schema for the 'latest version' require.
 *
 * @param data - The current value from the API.
 * @param defaults - Default values.
 */
export const buildLatestVersionRequireSchemaWithFallbacks = (
	data?: LatestVersionRequire,
	defaults?: LatestVersionRequireDefaults,
): BuilderResponse<
	typeof latestVersionRequireSchema,
	typeof latestVersionRequireSchemaDefaults
> => {
	const path = 'latest_version.require';
	// command.
	const {
		schema: commandSchema,
		schemaData: commandSchemaData,
		schemaDataDefaultsHollow: commandDefaultsSchemaData,
	} = buildCommandSchemaWithFallbacks(data?.command, defaults?.command);
	// docker.
	const {
		schema: dockerSchema,
		schemaData: dockerSchemaData,
		schemaDataDefaults: dockerDefaultsSchemaData,
	} = buildDockerFilterSchemaWithFallbacks(data?.docker, defaults?.docker);

	// Latest version require schema.
	const schema = latestVersionRequireSchema.extend({
		command: commandSchema.default(commandSchemaData),
		docker: dockerSchema,
		regex_content: regexStringWithFallback(false, defaults?.regex_content),
		regex_version: regexStringWithFallback(false, defaults?.regex_version),
	});

	const invalidSchemaFallback = {
		command: [],
		docker: dockerFilterSchema.parse({
			type: LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.DOCKER_HUB.value,
		}),
		regex_content: '',
		regex_version: '',
	};

	// Initial schema data.
	const schemaData = safeParse({
		data: {
			command: commandSchemaData,
			docker: dockerSchemaData,
			regex_content: data?.regex_content ?? '',
			regex_version: data?.regex_version ?? '',
		},
		fallback: invalidSchemaFallback,
		path: path,
		schema: schema,
	});

	// Defaults for the schema.
	const schemaDataDefaults = safeParse({
		data: {
			...defaults,
			command: commandDefaultsSchemaData,
			docker: dockerDefaultsSchemaData,
		},
		fallback: invalidSchemaFallback,
		path: `${path} (defaults)`,
		schema: latestVersionRequireSchemaDefaults,
	});

	return {
		schema: schema,
		schemaData: schemaData,
		schemaDataDefaults: schemaDataDefaults,
	};
};

/**
 * Builds a schema for the 'latest version' lookup.
 *
 * @param data - The current value from the API.
 * @param defaults - Default values.
 * @param hardDefaults - Hard default values.
 */
export const buildLatestVersionLookupSchemaWithFallbacks = (
	data?: LatestVersionLookup,
	defaults?: LatestVersionLookupDefaults,
	hardDefaults?: LatestVersionLookupDefaults,
): BuilderResponse<
	typeof latestVersionLookupSchema,
	typeof latestVersionLookupSchemaDefault
> => {
	const path = 'latest_version';
	const combinedDefaults = applyDefaultsRecursive<LatestVersionLookupDefaults>(
		defaults ?? null,
		hardDefaults,
	);
	const defaultType = isLatestVersionType(combinedDefaults.type)
		? combinedDefaults.type
		: undefined;

	// url_commands.
	const {
		schema: urlCommandsSchema,
		schemaData: urlCommandsSchemaData,
		schemaDataDefaults: urlCommandsSchemaDataDefaults,
	} = buildURLCommandsSchemaWithFallbacks(
		data?.url_commands,
		combinedDefaults?.url_commands,
	);
	// require.
	const {
		schema: requireSchema,
		schemaData: requireSchemaData,
		schemaDataDefaults: requireSchemaDataDefaults,
	} = buildLatestVersionRequireSchemaWithFallbacks(
		data?.require,
		combinedDefaults?.require,
	);
	// headers (URL-type only).
	const {
		schema: headersSchema,
		schemaData: headersSchemaData,
		schemaDataDefaults: headersSchemaDataDefaults,
	} = buildHeadersSchemaWithFallbacks(
		data && 'headers' in data ? data.headers : [],
		'headers' in combinedDefaults ? combinedDefaults.headers : [],
	);

	// Schemas shared between multiple types.
	const sharedSchemas = {
		require: requireSchema,
		url_commands: urlCommandsSchema,
	};

	// Latest version schema.
	const schemaRaw = z.discriminatedUnion('type', [
		latestVersionLookupSchemaGitHub.extend(sharedSchemas),
		latestVersionLookupSchemaURL.extend({
			...sharedSchemas,
			headers: headersSchema,
		}),
	]);
	const schema = z.discriminatedUnion('type', [
		latestVersionLookupSchemaGitHub.extend({
			...sharedSchemas,
			url: stringDefault.superRefine((arg, ctx) => {
				const url = arg || (defaults?.url ?? hardDefaults?.url ?? '');

				validateRequired({ arg: url, ctx: ctx });
				validateGitHubRepo({ arg: url, ctx: ctx });
			}),
		}),
		latestVersionLookupSchemaURL.extend({
			...sharedSchemas,
			headers: headersSchema,
			url: z.string().superRefine((arg, ctx) => {
				const url = arg || (defaults?.url ?? hardDefaults?.url ?? '');

				validateRequired({ arg: url, ctx: ctx });
				validateURL({ arg: url, ctx: ctx });
			}),
		}),
	]);

	// Initial type.
	const schemaDataType = isLatestVersionType(data?.type)
		? data.type
		: defaultType;
	// Initial schema data.
	const fallbackData: Partial<z.infer<typeof schemaRaw>> = {
		require: requireSchemaData,
		type: schemaDataType,
		url_commands: urlCommandsSchemaData,
	};
	// Type-specific schema data.
	if (schemaDataType === LATEST_VERSION_LOOKUP_TYPE.GITHUB.value) {
		const typedLatestVersion = (data ?? {}) as LatestVersionLookupGitHub;
		(fallbackData as LatestVersionLookupGitHub).use_prerelease =
			typedLatestVersion.use_prerelease;
	} else {
		// URL.
		const typedLatestVersion = (data ?? {}) as LatestVersionLookupURL;
		(fallbackData as LatestVersionLookupURL).allow_invalid_certs =
			typedLatestVersion.allow_invalid_certs;
		(fallbackData as LatestVersionLookupURL).headers = headersSchemaData;
	}
	const schemaData = safeParse({
		data: {
			url: '',
			...data,
			...fallbackData,
		},
		// biome-ignore lint/suspicious/noExplicitAny: couldn't get the discriminated union type to work as fallback.
		fallback: fallbackData as any,
		path: path,
		schema: schemaRaw,
	});

	// Defaults for the schema.
	const schemaDataDefaults = safeParse({
		data: {
			...combinedDefaults,
			headers: headersSchemaDataDefaults,
			require: requireSchemaDataDefaults,
			type: defaultType,
			url_commands: urlCommandsSchemaDataDefaults,
		},
		fallback: {
			headers: headersSchemaDataDefaults,
			require: requireSchemaDataDefaults,
			type: fallbackData.type as LatestVersionLookupType,
		},
		path: path,
		schema: latestVersionLookupSchemaDefault,
	});

	return {
		schema: schema,
		schemaData: schemaData,
		schemaDataDefaults: schemaDataDefaults,
	};
};
