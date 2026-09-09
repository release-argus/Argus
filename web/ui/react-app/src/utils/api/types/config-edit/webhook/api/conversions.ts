import { removeEmptyValues } from '@/utils';
import { applyDefaultsRecursive } from '@/utils/api/types/config-edit/util';
import {
	type WebhookSchema,
	type WebhookSchemaOutgoing,
	type WebhooksSchema,
	type WebhooksSchemaOutgoing,
	webhookSchemaMapOutgoingWithDefaults,
} from '@/utils/api/types/config-edit/webhook/schemas';
import diffLists from '@/utils/diff-lists';

/**
 * Maps the webhook form schema to an API payload.
 *
 * @param data - The form schema.
 * @param defaultValue - The default values to compare against (and omit if all defaults used and unmodified).
 * @param mainDefaults - The 'webhook' globals.
 * @param typeDefaults - Type-specific webhook form data.
 * @returns The API payload with matching defaults removed.
 */
export const mapWebhooksSchemaToAPIPayload = (
	data: WebhooksSchema,
	defaultValue?: WebhooksSchema,
	mainDefaults?: Record<string, WebhookSchema>,
	typeDefaults?: WebhookSchema,
): WebhooksSchemaOutgoing => {
	const dataMinimised = data.map((item) => {
		const defaultsForItem = applyDefaultsRecursive(
			mainDefaults?.[item.name] ?? null,
			typeDefaults,
		);
		const d = mapWebhookSchemaToAPIPayload(item, defaultsForItem);
		return removeEmptyValues(d) as WebhookSchemaOutgoing;
	});

	// Omit if all defaults used and unmodified.
	if (
		defaultValue &&
		data.length === defaultValue.length &&
		dataMinimised.every(
			(i) =>
				Object.keys(i).length === 2 && ['name', 'type'].every((k) => k in i),
		) &&
		!diffLists({
			key: 'name',
			listA: defaultValue,
			listB: dataMinimised,
		})
	) {
		return null;
	}

	return dataMinimised;
};

/**
 * Maps the webhook form schema to an API payload.
 *
 * @param item - The form schema.
 * @param defaults - The default values to compare against (and omit where used).
 * @returns A `WebhookSchemaOutgoing` representing the `WebhookSchema`.
 */
export const mapWebhookSchemaToAPIPayload = (
	item: WebhookSchema,
	defaults?: WebhookSchema,
): WebhookSchemaOutgoing => {
	const schema = webhookSchemaMapOutgoingWithDefaults(defaults);

	return removeEmptyValues(
		schema.parse(item) as WebhookSchemaOutgoing,
	) as WebhookSchemaOutgoing;
};
