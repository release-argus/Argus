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
	stringDefault,
} from '@/utils/api/types/config-edit/shared/preprocess';
import { stringWithFallback } from '@/utils/api/types/config-edit/validators';

export const WebhookTypeEnum = z.enum(
	toZodEnumTuple(Object.values(WEBHOOK_TYPE)),
);

export const webhookSchema = z.object({
	allow_invalid_certs: z.boolean().nullable().default(null),
	delay: stringDefault,
	desired_status_code: preprocessStringFromNumber,
	headers: headersSchema,
	max_tries: preprocessStringFromNumber,
	name: stringWithFallback(), // Required.
	old_index: z.string().nullable().default(null),
	secret: stringDefault, // Required.
	silent_fails: z.boolean().nullable().default(null),
	type: WebhookTypeEnum,
	url: stringDefault, // Required.
});
export type WebhookSchema = z.infer<typeof webhookSchema>;
export const webhooksSchema = z.array(webhookSchema).default([]);
export type WebhooksSchema = z.infer<typeof webhooksSchema>;

export const webhookSchemaDefault = webhookSchema.extend({
	name: stringDefault,
});
export type WebhookSchemaDefault = z.infer<typeof webhookSchemaDefault>;
export const webhooksSchemaDefault = z.array(webhookSchemaDefault).default([]);
export type WebhooksSchemaDefault = z.infer<typeof webhooksSchemaDefault>;

/* API Outgoing requests */

export const webhookSchemaOutgoing = webhookSchema.extend({
	desired_status_code: preprocessNumberFromString,
	max_tries: preprocessNumberFromString,
});
export type WebhookSchemaOutgoing = z.infer<typeof webhookSchemaOutgoing>;

export const webhooksSchemaOutgoing = z
	.array(webhookSchemaOutgoing)
	.nullable()
	.default(null);
export type WebhooksSchemaOutgoing = z.infer<typeof webhooksSchemaOutgoing>;

/**
 * Outgoing schemas that are defaults-aware for list-like fields.
 *
 * @returns a per-type schema with the provided defaults where
 * preprocessors can null fields that match the defaults.
 */
export const webhookSchemaMapOutgoingWithDefaults = (
	defaults?: WebhookSchema,
) => {
	return webhookSchema.extend({
		desired_status_code: preprocessNumberFromString,
		headers: preprocessHeadersToHeadersSchema(defaults?.headers),
		max_tries: preprocessNumberFromString,
	});
};
