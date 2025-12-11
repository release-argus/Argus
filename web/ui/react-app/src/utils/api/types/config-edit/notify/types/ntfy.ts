import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import {
	NTFY_ACTION_TYPE,
	ntfyPriorityOptions,
	ntfySchemeOptions,
} from '@/utils/api/types/config/notify/ntfy';
import {
	flattenHeaderArray,
	headersSchemaDefaults,
} from '@/utils/api/types/config-edit/shared/header/preprocess';
import {
	makeDefaultsAwareListPreprocessor,
	preprocessArrayJSONFromString,
} from '@/utils/api/types/config-edit/shared/preprocess';

export const NtfyPriorityZodEnum = z.enum(toZodEnumTuple(ntfyPriorityOptions));

export const NtfySchemeZodEnum = z.enum(toZodEnumTuple(ntfySchemeOptions));

/* Action */
export const ntfyActionBase = z.object({
	clear: z.boolean().default(false),
	label: z.string().default(''),
});
export const ntfyActionBroadcast = ntfyActionBase.extend({
	action: z.literal(NTFY_ACTION_TYPE.BROADCAST.value),
	extras: headersSchemaDefaults,
	intent: z.string().default(''),
});
export const ntfyActionHTTP = ntfyActionBase.extend({
	action: z.literal(NTFY_ACTION_TYPE.HTTP.value),
	body: z.string().default(''),
	headers: headersSchemaDefaults,
	method: z.string().default(''),
	url: z.string().default(''),
});
export const ntfyActionView = ntfyActionBase.extend({
	action: z.literal(NTFY_ACTION_TYPE.VIEW.value),
	url: z.string().default(''),
});

export const ntfyActionSchema = z.discriminatedUnion('action', [
	ntfyActionBroadcast,
	ntfyActionHTTP,
	ntfyActionView,
]);
export const ntfyActionsSchema =
	preprocessArrayJSONFromString(ntfyActionSchema);

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
