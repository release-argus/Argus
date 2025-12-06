import { z } from 'zod';
import {
	commandsSchema,
	commandsSchemaOutgoing,
} from '@/utils/api/types/config-edit/command/schemas';
import {
	notifiersSchema,
	notifiersSchemaOutgoing,
} from '@/utils/api/types/config-edit/notify/schemas';
import { serviceDashboardOptionsSchema } from '@/utils/api/types/config-edit/service/types/dashboard';
import {
	deployedVersionLookupSchema,
	deployedVersionLookupSchemaDefault,
} from '@/utils/api/types/config-edit/service/types/deployed-version';
import {
	latestVersionLookupSchema,
	latestVersionLookupSchemaDefault,
	latestVersionLookupSchemaOutgoing,
} from '@/utils/api/types/config-edit/service/types/latest-version';
import {
	serviceOptionsSchema,
	serviceOptionsSchemaDefaults,
} from '@/utils/api/types/config-edit/service/types/options';
import { stringWithFallback } from '@/utils/api/types/config-edit/validators';
import {
	webhooksSchema,
	webhooksSchemaOutgoing,
} from '@/utils/api/types/config-edit/webhook/schemas';

export const serviceSchema = z.object({
	command: commandsSchema,
	comment: z.string().default(''),
	dashboard: serviceDashboardOptionsSchema,
	deployed_version: deployedVersionLookupSchema,
	id: stringWithFallback(),
	id_name_separator: z.boolean().default(false),
	latest_version: latestVersionLookupSchema,
	name: z.string().default(''),
	notify: notifiersSchema,
	options: serviceOptionsSchema,
	webhook: webhooksSchema,
});
export type ServiceSchema = z.infer<typeof serviceSchema>;

export const serviceSchemaDefault = serviceSchema.extend({
	dashboard: serviceDashboardOptionsSchema,
	deployed_version: deployedVersionLookupSchemaDefault,
	id: z.string().optional(),
	latest_version: latestVersionLookupSchemaDefault,
	notify: notifiersSchema,
	options: serviceOptionsSchemaDefaults,
	webhook: webhooksSchema,
});
export type ServiceSchemaDefault = z.infer<typeof serviceSchemaDefault>;

export const serviceSchemaOutgoing = serviceSchema
	.omit({ id_name_separator: true })
	.extend({
		command: commandsSchemaOutgoing,
		latest_version: latestVersionLookupSchemaOutgoing,
		notify: notifiersSchemaOutgoing,
		webhook: webhooksSchemaOutgoing,
	});
export type ServiceSchemaOutgoing = z.infer<typeof serviceSchemaOutgoing>;
