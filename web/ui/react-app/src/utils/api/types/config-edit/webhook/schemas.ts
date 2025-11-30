import { z } from 'zod';
import { toZodEnumTuple } from '@/types/util';
import { WEBHOOK_TYPE } from '@/utils/api/types/config/webhook';
import { headersSchemaDefaults } from '@/utils/api/types/config-edit/shared/header/schemas';
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
	custom_headers: headersSchemaDefaults,
	delay: z.string().default(''),
	desired_status_code: preprocessStringFromNumber,
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
