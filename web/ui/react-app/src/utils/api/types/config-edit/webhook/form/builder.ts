import { z } from 'zod';
import { isEmptyArray } from '@/utils';
import {
	isWebHookType,
	WEBHOOK_TYPE,
	type WebHook,
	type WebHookMap,
	webhookTypeOptions,
} from '@/utils/api/types/config/webhook';
import { buildSuperRefine } from '@/utils/api/types/config-edit/shared/builder--super-refine';
import { safeParse } from '@/utils/api/types/config-edit/shared/safeparse';
import { superRefineNameUnique } from '@/utils/api/types/config-edit/shared/unique-name--super-refine';
import { applyDefaultsRecursive } from '@/utils/api/types/config-edit/util';
import {
	validateListUniqueKeys,
	validateListWithSchemas,
	validateNumberInRange,
	validateNumberString,
	validateRequired,
} from '@/utils/api/types/config-edit/validators';
import {
	type WebHookSchema,
	webhookSchema,
	webhookSchemaDefault,
} from '@/utils/api/types/config-edit/webhook/schemas';

/**
 * Builds a schema for a webhook.
 *
 * @param defaults - Default values for a webhook.
 * @param mains - The 'main' webhook objects that may be referenced.
 */
const buildWebHookSchema = (
	defaults: WebHookSchema,
	mains: Record<string, WebHookSchema>,
) => {
	return buildSuperRefine(webhookSchema, mains, defaults, [
		{ path: ['desired_status_code'], validator: validateNumberString },
		{ path: ['max_tries'], validator: validateNumberString },
		{
			path: ['max_tries'],
			validator: validateNumberInRange({ max: 255, min: 0 }),
		},
		{ path: ['secret'], validator: validateRequired },
		{ path: ['url'], validator: validateRequired },
		{
			kind: 'array',
			path: ['custom_headers'],
			props: [
				{
					matchingFields: [],
				},
			],
			validator: [validateListWithSchemas, validateListUniqueKeys],
		},
	]);
};

/**
 * Builds a schema for an array of webhooks.
 *
 * @param data - The current data from the API.
 * @param defaultItems - The default webhooks to use.
 * @param mains - The 'main' webhook objects that may be referenced.
 * @param defaults - Default values for a webhook.
 * @param hardDefaults - Hard defaults for a webhook.
 */
export const buildWebHooksSchemaWithFallbacks = (
	data?: WebHook[],
	defaultItems?: Record<string, unknown>,
	mains?: WebHookMap,
	defaults?: WebHook,
	hardDefaults?: WebHook,
) => {
	const path = 'webhook';
	const defaultType =
		defaults?.type ??
		hardDefaults?.type ??
		Object.values(webhookTypeOptions)[0].value;
	const dataDefaulted = (data ?? []).map((item) => {
		const main = mains?.[item.name];
		const nameLower = item.name.toLowerCase();
		const itemType =
			main?.type ?? (isWebHookType(nameLower) ? nameLower : defaultType);

		return safeParse({
			data: {
				...item,
				type: itemType,
			},
			fallback: {
				desired_status_code: '',
				max_tries: '',
				type: defaultType,
			},
			path: `${path} (defaults-${item.name})`,
			schema: webhookSchema,
		});
	});
	const combinedDefaults = applyDefaultsRecursive<WebHook>(
		defaults ?? null,
		hardDefaults,
		{ custom_headers: [], type: WEBHOOK_TYPE.GITHUB.value },
	);
	const schemaDataTypeDefaults = safeParse({
		data: combinedDefaults,
		fallback: {
			desired_status_code: '',
			max_tries: '',
			type: defaultType,
		},
		path: `${path} (defaults)`,
		schema: webhookSchemaDefault,
	});
	const typeDefault = schemaDataTypeDefaults.type;

	// Default schema data.
	const schemaDataDefaults: WebHookSchema[] = Object.keys(
		defaultItems ?? {},
	).map((name) => {
		const main = mains?.[name];
		const nameLower = name.toLowerCase();
		const itemType =
			main?.type ?? (isWebHookType(nameLower) ? nameLower : typeDefault);
		// custom_headers.
		const customHeaders = isEmptyArray(main?.custom_headers)
			? schemaDataTypeDefaults.custom_headers
			: main?.custom_headers;
		const customHeadersHollow = customHeaders?.map(() => ({
			key: '',
			value: '',
		}));

		const data = {
			custom_headers: customHeadersHollow,
			name: name,
			old_index: null,
			type: itemType,
		};
		return safeParse({
			data: data,
			fallback: {
				desired_status_code: '',
				max_tries: '',
				type: itemType,
			},
			path: `${path} (defaults)`,
			schema: webhookSchema,
		});
	});

	// Defaults for each type.
	const schemaDataTypeDefaultsHollow = {
		custom_headers: schemaDataTypeDefaults.custom_headers.map(() => ({
			key: '',
			value: '',
		})),
	};

	// Defaults for each main.
	const schemaDataMains = Object.entries(mains ?? {}).reduce<
		Record<string, WebHookSchema>
	>((acc, [name, main]) => {
		acc[name] = safeParse({
			data: applyDefaultsRecursive({ ...main, name: name }, combinedDefaults),
			fallback: {
				desired_status_code: '',
				max_tries: '',
				name: name,
				type: main.type ?? typeDefault,
			},
			path: `${path} (mains-${name}`,
			schema: webhookSchema,
		});
		return acc;
	}, {});

	// Schemas.
	const schema = buildWebHookSchema(schemaDataTypeDefaults, schemaDataMains);
	const schemaFinal = z
		.array(schema)
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
		schemaDataTypeDefaultsHollow: schemaDataTypeDefaultsHollow,
	};
};
