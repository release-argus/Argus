import { removeEmptyValues } from '@/utils';
import {
	type WebHookSchema,
	type WebHookSchemaOutgoing,
	type WebHooksSchema,
	type WebHooksSchemaOutgoing,
	webhookSchemaOutgoing,
} from '@/utils/api/types/config-edit/webhook/schemas';

/**
 * Maps the webhook form schema to an API payload.
 *
 * @param data - The form schema.
 * @param defaultValue - Default values for the form schema.
 * @returns The API payload with matching defaults removed.
 */
export const mapWebHooksSchemaToAPIPayload = (
	data: WebHooksSchema,
	defaultValue?: WebHooksSchema,
): WebHooksSchemaOutgoing => {
	const dataMinimised = data.map((item) => {
		const d = mapWebHookSchemaToAPIPayload(item);
		return removeEmptyValues(d) as WebHookSchemaOutgoing;
	});

	// Omit if all defaults used and unmodified.
	if (
		defaultValue &&
		data.length == defaultValue.length &&
		dataMinimised.every(
			(i) =>
				Object.keys(i).length === 2 && ['name', 'type'].every((k) => k in i),
		)
	) {
		const sortedDefaultValue = defaultValue
			.map((i) => i.name)
			.toSorted((a, b) => a.localeCompare(b))
			.join('-_-');
		const sortedDataMinimised = dataMinimised
			.map((i) => i.name)
			.toSorted((a, b) => a.localeCompare(b))
			.join('-_-');

		if (sortedDefaultValue === sortedDataMinimised) return null;
	}

	return dataMinimised;
};

/**
 * Maps the webhook form schema to an API payload.
 *
 * @param item - The form schema.
 */
export const mapWebHookSchemaToAPIPayload = (
	item: WebHookSchema,
): WebHookSchemaOutgoing => webhookSchemaOutgoing.parse(item);
