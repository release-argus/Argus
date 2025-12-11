import { removeEmptyValues } from '@/utils';
import {
	type NotifiersSchema,
	type NotifiersSchemaOutgoing,
	type NotifySchemaOutgoing,
	type NotifySchemaValues,
	type NotifyTypeSchema,
	notifySchemaMapOutgoingWithDefaults,
} from '@/utils/api/types/config-edit/notify/schemas';
import { applyDefaultsRecursive } from '@/utils/api/types/config-edit/util';
import diffLists from '@/utils/diff-lists';

/**
 * Converts a `NotifiersSchema` object to a `NotifiersSchemaOutgoing` object.
 *
 * @param data - The `NotifiersSchema` data to map.
 * @param defaultValue - The default values to compare against (and omit if all defaults used and unmodified).
 * @param mainDefaults - The 'notify' globals.
 * @param typeDefaults - Type-specific notify form data.
 * @returns A `NotifiersSchemaOutgoing` representing the `NotifiersSchema`.
 */
export const mapNotifiersSchemaToAPIPayload = (
	data: NotifiersSchema,
	defaultValue?: NotifiersSchema,
	mainDefaults?: Record<string, NotifySchemaValues>,
	typeDefaults?: NotifyTypeSchema,
): NotifiersSchemaOutgoing => {
	const dataMinimised = data.map((item) => {
		const defaultsForItem = applyDefaultsRecursive(
			mainDefaults?.[item.name] ?? null,
			typeDefaults?.[item.type],
		);
		const d = mapNotifySchemaToAPIPayload(item, defaultsForItem);
		return removeEmptyValues(d) as NotifySchemaOutgoing;
	});

	// Omit if all defaults used and unmodified.
	if (
		defaultValue &&
		data.length == defaultValue.length &&
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
 * Converts a `NotifySchemaValues` object to a `NotifySchemaOutgoing` object.
 *
 * @param item - The `NotifySchemaValues` to map.
 * @param defaults - The default values to compare against (and omit where used).
 * @returns A `NotifySchemaOutgoing` representing the `NotifySchemaValues`.
 */
export const mapNotifySchemaToAPIPayload = (
	item: NotifySchemaValues,
	defaults?: NotifySchemaValues,
): NotifySchemaOutgoing => {
	const itemType = item.type;
	const defaultsTyped = defaults ?? ({ type: itemType } as NotifySchemaValues);
	const schema = notifySchemaMapOutgoingWithDefaults(defaultsTyped);

	return removeEmptyValues(
		schema.parse(item) as NotifySchemaOutgoing,
	) as NotifySchemaOutgoing;
};
