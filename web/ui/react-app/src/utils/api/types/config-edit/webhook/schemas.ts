import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import { WEBHOOK_TYPE } from '@/utils/api/types/config/webhook';
import {
	headersSchema,
	preprocessHeadersToHeadersSchema,
} from '@/utils/api/types/config-edit/shared/header/preprocess';
import {
	preprocessNumberFromString,
	preprocessStringFromNumber,
} from '@/utils/api/types/config-edit/shared/preprocess';
import { stringWithFallback } from '@/utils/api/types/config-edit/validators';

export const WebHookTypeEnum = z.enum(
	toZodEnumTuple(Object.values(WEBHOOK_TYPE)),
);

export const webhookSchema = z.object({
	allow_invalid_certs: z.boolean().nullable().default(null),
	delay: z.string().default(''),
	desired_status_code: preprocessStringFromNumber,
	headers: headersSchema,
	max_tries: preprocessStringFromNumber,
	name: stringWithFallback(), // Required.
	old_index: z.string().nullable().default(null),
	secret: z.string().default(''), // Required.
	silent_fails: z.boolean().nullable().default(null),
	type: WebHookTypeEnum,
	url: z.string().default(''), // Required.
});
export type WebHookSchema = z.infer<typeof webhookSchema>;
export const webhooksSchema = z.array(webhookSchema).default([]);
export type WebHooksSchema = z.infer<typeof webhooksSchema>;

export const webhookSchemaDefault = webhookSchema.extend({
	name: z.string().default(''),
});
export type WebHookSchemaDefault = z.infer<typeof webhookSchemaDefault>;
export const webhooksSchemaDefault = z.array(webhookSchemaDefault).default([]);
export type WebHooksSchemaDefault = z.infer<typeof webhooksSchemaDefault>;

/* API Outgoing requests */

export const webhookSchemaOutgoing = webhookSchema.extend({
	desired_status_code: preprocessNumberFromString,
	max_tries: preprocessNumberFromString,
});
export type WebHookSchemaOutgoing = z.infer<typeof webhookSchemaOutgoing>;

export const webhooksSchemaOutgoing = z
	.array(webhookSchemaOutgoing)
	.nullable()
	.default(null);
export type WebHooksSchemaOutgoing = z.infer<typeof webhooksSchemaOutgoing>;

/**
 * Outgoing schemas that are defaults-aware for list-like fields.
 *
 * @returns a per-type schema with the provided defaults where
 * preprocessors can null fields that match the defaults.
 */
export const webhookSchemaMapOutgoingWithDefaults = (
	defaults?: WebHookSchema,
) => {
	return webhookSchema.extend({
		desired_status_code: preprocessNumberFromString,
		headers: preprocessHeadersToHeadersSchema(defaults?.headers),
		max_tries: preprocessNumberFromString,
	});
};
