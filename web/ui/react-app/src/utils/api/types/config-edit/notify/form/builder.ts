import { z } from 'zod';
import type {
	NotifyMap,
	NotifyTypesMap,
	NotifyTypesValues,
} from '@/utils/api/types/config/notify';
import { NOTIFY_TYPE_MAP } from '@/utils/api/types/config/notify/all-types';
import {
	isNotifyType,
	type NotifySchemaKeys,
	type NotifySchemaValues,
	type NotifyTypeSchema,
	notifiersSchema,
	notifyBarkSchema,
	notifyDiscordSchema,
	notifyGenericSchema,
	notifyGoogleChatSchema,
	notifyGotifySchema,
	notifyIFTTTSchema,
	notifyJoinSchema,
	notifyMatrixSchema,
	notifyMatterMostSchema,
	notifyNtfySchema,
	notifyOpsGenieSchema,
	notifyPushbulletSchema,
	notifyPushoverSchema,
	notifyRocketChatSchema,
	notifySchemaMap,
	notifySlackSchema,
	notifySMTPSchema,
	notifyTeamsSchema,
	notifyTelegramSchema,
	notifyZulipSchema,
} from '@/utils/api/types/config-edit/notify/schemas';
import { buildSuperRefine } from '@/utils/api/types/config-edit/shared/builder--super-refine';
import { safeParse } from '@/utils/api/types/config-edit/shared/safeparse';
import { superRefineNameUnique } from '@/utils/api/types/config-edit/shared/unique-name--super-refine';
import {
	applyDefaultsRecursive,
	atLeastTwo,
} from '@/utils/api/types/config-edit/util';
import {
	validateHexString,
	validateListUniqueKeys,
	validateListWithSchemas,
	validateNumberInRange,
	validateNumberString,
	validateRequired,
} from '@/utils/api/types/config-edit/validators';

/* Validators shared by all notify types. */
const defaultValidators = [
	{
		path: ['options', 'max_tries'],
		validator: validateNumberString,
	},
	{
		path: ['options', 'max_tries'],
		validator: validateNumberInRange({ max: 255, min: 0 }),
	},
];

/**
 * Builds a schema for a specific notify type.
 *
 * @param defaults - The default values for the notify type.
 * @param mains - All 'main' objects for the notify type.
 */
const buildNotifySchema = (
	defaults: NotifySchemaValues,
	mains: Record<string, NotifySchemaValues>,
) => {
	const itemType = defaults.type;

	let schema;
	switch (itemType) {
		case NOTIFY_TYPE_MAP.BARK.value:
			schema = buildSuperRefine(notifyBarkSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'host'], validator: validateRequired },
				{ path: ['url_fields', 'port'], validator: validateNumberString },
				{ path: ['params', 'badge'], validator: validateNumberString },
			]);
			break;
		case NOTIFY_TYPE_MAP.DISCORD.value:
			schema = buildSuperRefine(notifyDiscordSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'token'], validator: validateRequired },
				{ path: ['url_fields', 'webhookid'], validator: validateRequired },
			]);
			break;
		case NOTIFY_TYPE_MAP.SMTP.value:
			schema = buildSuperRefine(notifySMTPSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'host'], validator: validateRequired },
				{ path: ['url_fields', 'port'], validator: validateNumberString },
				{ path: ['params', 'fromaddress'], validator: validateRequired },
				{ path: ['params', 'toaddresses'], validator: validateRequired },
			]);
			break;
		case NOTIFY_TYPE_MAP.GOOGLE_CHAT.value:
			schema = buildSuperRefine(notifyGoogleChatSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'raw'], validator: validateRequired },
			]);
			break;
		case NOTIFY_TYPE_MAP.GOTIFY.value:
			schema = buildSuperRefine(notifyGotifySchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'host'], validator: validateRequired },
				{ path: ['url_fields', 'port'], validator: validateNumberString },
				{ path: ['url_fields', 'token'], validator: validateRequired },
				{ path: ['params', 'priority'], validator: validateNumberString },
			]);
			break;
		case NOTIFY_TYPE_MAP.IFTTT.value:
			schema = buildSuperRefine(notifyIFTTTSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'webhookid'], validator: validateRequired },
				{ path: ['params', 'events'], validator: validateRequired },
				{
					path: ['params', 'usemessageasvalue'],
					validator: validateNumberString,
				},
				{ path: ['params', 'usetitlevalue'], validator: validateNumberString },
			]);
			break;
		case NOTIFY_TYPE_MAP.JOIN.value:
			schema = buildSuperRefine(notifyJoinSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'apikey'], validator: validateRequired },
				{ path: ['params', 'devices'], validator: validateRequired },
			]);
			break;
		case NOTIFY_TYPE_MAP.MATTERMOST.value:
			schema = buildSuperRefine(notifyMatterMostSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'host'], validator: validateRequired },
				{ path: ['url_fields', 'port'], validator: validateNumberString },
				{ path: ['url_fields', 'token'], validator: validateRequired },
			]);
			break;
		case NOTIFY_TYPE_MAP.MATRIX.value:
			schema = buildSuperRefine(notifyMatrixSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'host'], validator: validateRequired },
				{ path: ['url_fields', 'password'], validator: validateRequired },
				{ path: ['url_fields', 'port'], validator: validateNumberString },
			]);
			break;
		case NOTIFY_TYPE_MAP.NTFY.value:
			schema = buildSuperRefine(notifyNtfySchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'host'], validator: validateRequired },
				{ path: ['url_fields', 'port'], validator: validateNumberString },
				{ path: ['url_fields', 'topic'], validator: validateRequired },
				{
					kind: 'array',
					path: ['params', 'actions'],
					props: [
						{
							matchingFields: ['action', 'method'],
							notRequired: ['body', 'intent'],
						},
					],
					validator: [validateListWithSchemas],
				},
			]);
			break;
		case NOTIFY_TYPE_MAP.OPSGENIE.value:
			schema = buildSuperRefine(notifyOpsGenieSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'apikey'], validator: validateRequired },
				{
					kind: 'array',
					path: ['params', 'actions'],
					validator: [validateListWithSchemas],
				},
				{
					kind: 'array',
					path: ['params', 'details'],
					validator: [validateListWithSchemas, validateListUniqueKeys],
				},
				{
					kind: 'array',
					path: ['params', 'responders'],
					props: [
						{
							matchingFields: ['type', 'sub_type'],
						},
					],
					validator: [validateListWithSchemas],
				},
				{
					kind: 'array',
					path: ['params', 'visibleto'],
					props: [
						{
							matchingFields: ['type', 'sub_type'],
						},
					],
					validator: [validateListWithSchemas],
				},
			]);
			break;
		case NOTIFY_TYPE_MAP.PUSHBULLET.value:
			schema = buildSuperRefine(notifyPushbulletSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'targets'], validator: validateRequired },
				{ path: ['url_fields', 'token'], validator: validateRequired },
			]);
			break;
		case NOTIFY_TYPE_MAP.PUSHOVER.value:
			schema = buildSuperRefine(notifyPushoverSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'token'], validator: validateRequired },
				{ path: ['url_fields', 'user'], validator: validateRequired },
				{ path: ['params', 'priority'], validator: validateNumberString },
			]);
			break;
		case NOTIFY_TYPE_MAP.ROCKET_CHAT.value:
			schema = buildSuperRefine(notifyRocketChatSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'channel'], validator: validateRequired },
				{ path: ['url_fields', 'host'], validator: validateRequired },
				{ path: ['url_fields', 'port'], validator: validateNumberString },
				{ path: ['params', 'path'], validator: validateRequired },
				{ path: ['params', 'tokena'], validator: validateRequired },
				{ path: ['params', 'tokenb'], validator: validateRequired },
			]);
			break;
		case NOTIFY_TYPE_MAP.SLACK.value:
			schema = buildSuperRefine(notifySlackSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'token'], validator: validateRequired },
				{ path: ['params', 'color'], validator: validateHexString },
			]);
			break;
		case NOTIFY_TYPE_MAP.TEAMS.value:
			schema = buildSuperRefine(notifyTeamsSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['params', 'color'], validator: validateHexString },
			]);
			break;
		case NOTIFY_TYPE_MAP.TELEGRAM.value:
			schema = buildSuperRefine(notifyTelegramSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'token'], validator: validateRequired },
			]);
			break;
		case NOTIFY_TYPE_MAP.ZULIP.value:
			schema = buildSuperRefine(notifyZulipSchema, mains, defaults, [
				...defaultValidators,
				{ path: ['url_fields', 'botmail'], validator: validateRequired },
				{ path: ['url_fields', 'botkey'], validator: validateRequired },
				{ path: ['url_fields', 'host'], validator: validateRequired },
				{ path: ['url_fields', 'port'], validator: validateNumberString },
			]);
			break;
		case NOTIFY_TYPE_MAP.GENERIC.value:
			schema = buildSuperRefine(notifyGenericSchema, mains, defaults, [
				...defaultValidators,
				{
					kind: 'array',
					path: ['url_fields', 'custom_headers'],
					validator: [validateListWithSchemas, validateListUniqueKeys],
				},
				{
					path: ['url_fields', 'host'],
					validator: validateRequired,
				},
				{
					kind: 'array',
					path: ['url_fields', 'json_payload_vars'],
					validator: [validateListWithSchemas, validateListUniqueKeys],
				},
				{
					path: ['url_fields', 'port'],
					validator: validateNumberString,
				},
				{
					kind: 'array',
					path: ['url_fields', 'query_vars'],
					validator: [validateListWithSchemas, validateListUniqueKeys],
				},
			]);
			break;
		default:
			schema = null;
			break;
	}

	return schema;
};

/**
 * Builds a schema that covers an array of any/all notify types.
 *
 * @param data - The current data from the API.
 * @param defaultItems - The default notifiers to use.
 * @param mains - The 'main' notifier objects that may be referenced.
 * @param defaults - Defaults for notify types.
 * @param hardDefaults - Hard defaults for notify types.
 */
export const buildNotifySchemaWithFallbacks = (
	data?: NotifyTypesValues[],
	defaultItems?: Record<string, unknown>,
	mains?: NotifyMap,
	defaults?: Partial<NotifyTypesMap>,
	hardDefaults?: NotifyTypesMap,
) => {
	const path = 'notify';
	const defaultType = NOTIFY_TYPE_MAP.DISCORD.value;
	const dataDefaulted = (data ?? []).map((item) => {
		const main = mains?.[item.name];
		const nameLower = item.name.toLowerCase();
		const itemType =
			item?.type ??
			main?.type ??
			(isNotifyType(nameLower) ? nameLower : defaultType);
		// Schema for this type
		const schema = notifySchemaMap[itemType];

		return safeParse({
			data: {
				...item,
				old_index: item.name,
				type: itemType,
			},
			fallback: {
				old_index: item.name,
				type: defaultType,
			},
			path: path,
			schema: schema,
		});
	});
	const combinedDefaults = Object.fromEntries(
		Object.entries(
			applyDefaultsRecursive(defaults ?? null, hardDefaults),
		).filter(([, n]) => isNotifyType(n.type)),
	) as NotifyTypesMap;
	const fallbackHardDefaults =
		hardDefaults ??
		Object.fromEntries(
			Object.values(NOTIFY_TYPE_MAP).map((n) => [
				n.value,
				notifySchemaMap[n.value].parse({ type: n.value }),
			]),
		);

	const schemaDataTypeDefaultsArray = safeParse({
		data: Object.values(combinedDefaults),
		fallback: Object.values(fallbackHardDefaults),
		path: `${path} (defaults)`,
		schema: notifiersSchema,
	});
	const schemaDataTypeDefaults = schemaDataTypeDefaultsArray.reduce(
		(acc, item) => {
			// biome-ignore lint/suspicious/noExplicitAny: type matches.
			acc[item.type] = item as any;
			return acc;
		},
		{} as Partial<NotifyTypeSchema>,
	) as NotifyTypeSchema;

	// Defaults for the schema.
	const schemaDataDefaults: NotifySchemaValues[] = Object.keys(
		defaultItems ?? {},
	)
		.map((name) => {
			const main = mains?.[name];
			const nameLower = name.toLowerCase();
			const itemType =
				main?.type ?? (isNotifyType(nameLower) ? nameLower : null);
			if (itemType === null) return null;
			// Schema for this type
			const schema = notifySchemaMap[itemType];

			return schema.parse({
				name: name,
				old_index: null,
				type: itemType,
			});
		})
		.filter((item) => item !== null);

	// Defaults for each main.
	const schemaDataMains = Object.entries(mains ?? {}).reduce<
		Record<string, NotifySchemaValues>
	>((acc, [key, value]) => {
		const itemType = value.type;
		const data = applyDefaultsRecursive(value, combinedDefaults[itemType]);

		acc[key] = safeParse({
			data: data,
			fallback: { type: itemType },
			path: `${path} (mains-${key})`,
			schema: notifySchemaMap[itemType],
		});
		return acc;
	}, {});

	// Schemas for each notify type.
	const schemas = Object.values(schemaDataTypeDefaultsArray).map((item) =>
		buildNotifySchema(item, schemaDataMains),
	) as {
		[K in NotifySchemaKeys]: (typeof notifySchemaMap)[K];
	}[NotifySchemaKeys][];

	const schemaFinal = z
		.array(z.discriminatedUnion('type', atLeastTwo(schemas)))
		.default(schemaDataDefaults)
		.superRefine(superRefineNameUnique);

	// Initial schema data.
	let schemaData;
	if (data) {
		schemaData = safeParse({
			data: dataDefaulted,
			fallback: schemaDataDefaults,
			path: path,
			schema: schemaFinal,
		});
	} else {
		schemaData = schemaDataDefaults;
	}

	return {
		schema: schemaFinal,
		schemaData: schemaData,
		schemaDataDefaults: schemaDataDefaults,
		schemaDataMains: schemaDataMains,
		schemaDataTypeDefaults: schemaDataTypeDefaults,
	};
};
