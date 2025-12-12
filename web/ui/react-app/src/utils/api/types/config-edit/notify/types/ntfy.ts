import { z } from 'zod';
import { normaliseForSelect } from '@/components/modals/service-edit/util';
import { toZodEnumTuple } from '@/types/util';
import {
	NTFY_ACTION_TYPE,
	ntfyActionHTTPMethodOptions,
	ntfyActionTypeOptions,
	ntfyPriorityOptions,
	ntfySchemeOptions,
} from '@/utils/api/types/config/notify/ntfy';
import {
	flattenHeaderArray,
	headersSchema,
	headersSchemaWithValidation,
} from '@/utils/api/types/config-edit/shared/header/preprocess';
import {
	makeDefaultsAwareListPreprocessor,
	preprocessArrayJSONFromString,
} from '@/utils/api/types/config-edit/shared/preprocess';
import { REQUIRED_MESSAGE } from '@/utils/api/types/config-edit/validators.ts';

export const NtfyPriorityZodEnum = z.enum(toZodEnumTuple(ntfyPriorityOptions));

export const NtfySchemeZodEnum = z.enum(toZodEnumTuple(ntfySchemeOptions));

/* Action */
export const ntfyActionBase = z.object({
	clear: z.boolean().default(false),
	label: z.string().default(''),
});
export const ntfyActionBroadcast = ntfyActionBase.extend({
	action: z.literal(NTFY_ACTION_TYPE.BROADCAST.value),
	extras: headersSchema,
	intent: z.string().default(''),
});
export const ntfyActionBroadcastWithValidation = ntfyActionBroadcast.extend({
	extras: headersSchemaWithValidation,
	label: z.string().min(1, REQUIRED_MESSAGE).default(''),
});
export const ntfyActionHTTP = ntfyActionBase.extend({
	action: z.literal(NTFY_ACTION_TYPE.HTTP.value),
	body: z.string().default(''),
	headers: headersSchema,
	method: z.string().default(''),
	url: z.string().default(''),
});
export const ntfyActionHTTPWithValidation = ntfyActionHTTP.extend({
	headers: headersSchemaWithValidation,
	method: z.string().min(1, REQUIRED_MESSAGE).default(''),
	url: z.string().min(1, REQUIRED_MESSAGE).default(''),
});
export const ntfyActionView = ntfyActionBase.extend({
	action: z.literal(NTFY_ACTION_TYPE.VIEW.value),
	url: z.string().default(''),
});
export const ntfyActionViewWithValidation = ntfyActionView.extend({
	url: z.string().min(1, REQUIRED_MESSAGE).default(''),
});

export const ntfyActionSchema = z.discriminatedUnion('action', [
	ntfyActionBroadcast,
	ntfyActionHTTP,
	ntfyActionView,
]);
export const ntfyActionSchemaWithValidation = z.discriminatedUnion('action', [
	ntfyActionBroadcastWithValidation,
	ntfyActionHTTPWithValidation,
	ntfyActionViewWithValidation,
]);

const normaliseNtfyActionsSchema = (
	actions: NtfyActionSchema[],
): NtfyActionSchema[] => {
	return actions.map((action) => {
		const actionNormalised = normaliseForSelect(
			ntfyActionTypeOptions,
			action.action,
		)?.value;
		if (actionNormalised) action.action = actionNormalised;

		if (action.action === NTFY_ACTION_TYPE.HTTP.value) {
			const method = normaliseForSelect(
				ntfyActionHTTPMethodOptions,
				action.method,
			)?.value;
			if (method) {
				action.method = method;
			}
			return action;
		} else {
			return action;
		}
	});
};

export const ntfyActionsSchema = preprocessArrayJSONFromString({
	jsonProcessor: normaliseNtfyActionsSchema,
	schema: ntfyActionSchema,
});
export const ntfyActionsSchemaWithValidation = preprocessArrayJSONFromString({
	schema: ntfyActionSchemaWithValidation,
});

export type NtfyActionSchema = z.infer<typeof ntfyActionSchema>;
export type NtfyActionsSchema = NtfyActionSchema[];

/**
 * Converts the Ntfy actions from an array of objects to a JSON string.
 *
 * @param obj - The `NotifyNtfyAction[]` to convert.
 * @returns A JSON string of the actions.
 */
export const preprocessStringFromNtfyActions = z.preprocess((val: unknown) => {
	if (!Array.isArray(val)) return JSON.stringify([]); // fallback

	return JSON.stringify(
		val.map((item) => {
			switch (item.action) {
				case 'view':
					return item;
				// http - headers as {KEY:VAL}, not {key:KEY, val:VAL}.
				case 'http':
					return {
						...item,
						headers: flattenHeaderArray(item.headers),
					};
				// broadcast - extras as {KEY:VAL}, not {key:KEY, val:VAL}.
				case 'broadcast':
					return {
						...item,
						extras: flattenHeaderArray(item.extras),
					};
				default:
					return item;
			}
		}),
	);
}, z.string());

/**
 * Defaults-aware variant of Ntfy actions -> string preprocessor.
 * - Empty array -> null.
 * - Matches defaults -> null.
 *
 * @param defaults - The default values for the actions.
 */
export const preprocessStringFromNtfyActionsWithDefaults = (
	defaults?: NtfyActionsSchema,
) =>
	makeDefaultsAwareListPreprocessor(
		preprocessStringFromNtfyActions.nullable(),
		{
			defaults: defaults,
			matchingFields: ['action', 'method', 'clear'],
		},
	);
