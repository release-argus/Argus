import { removeEmptyValues } from '@/utils';
import {
	type NotifiersSchema,
	type NotifiersSchemaOutgoing,
	type NotifySchemaOutgoing,
	type NotifySchemaValues,
	notifySchemaMapOutgoing,
	notifySchemaMapOutgoingWithDefaults,
} from '@/utils/api/types/config-edit/notify/schemas';
import diffLists from '@/utils/diff-lists.ts';

/**
 * Converts a `NotifiersSchema` object to a `NotifiersSchemaOutgoing` object.
 *
 * @param data - The `NotifiersSchema` data to map.
 * @param defaultValue - The default values to compare against (and omit if all defaults used and unmodified).
 * @returns A `NotifiersSchemaOutgoing` representing the `NotifiersSchema`.
 */
export const mapNotifiersSchemaToAPIPayload = (
	data: NotifiersSchema,
	defaultValue?: NotifiersSchema,
): NotifiersSchemaOutgoing => {
	const dataMinimised = data.map((item, idx) => {
		const defaultsForItem = defaultValue?.[idx];
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
	if (defaults?.type === itemType) {
		const schema = notifySchemaMapOutgoingWithDefaults(defaults);
		return removeEmptyValues(
			schema.parse(item) as NotifySchemaOutgoing,
		) as NotifySchemaOutgoing;
	}
	return removeEmptyValues(
		notifySchemaMapOutgoing[itemType].parse(item),
	) as NotifySchemaOutgoing;
};
