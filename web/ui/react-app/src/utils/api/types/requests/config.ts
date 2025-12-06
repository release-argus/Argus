import { stringify } from 'yaml';
import {
	containsEndsWith,
	containsStartsWith,
	isEmptyArray,
	isEmptyObject,
} from '@/utils';
import type { ConfigType } from '@/utils/api/types/config';

export type ConfigGetResponse = Record<string, unknown>;

/**
 * Recursively trims the object.
 *
 * @param obj - The object to trim.
 * @param path - The path of the object.
 * @returns The object with empty objects removed.
 */
const trimConfig = (
	obj: Record<string, unknown>,
	path = '',
): Record<string, unknown> => {
	return Object.keys(obj).reduce<Record<string, unknown>>((acc, key) => {
		const value = obj[key];

		if (typeof value === 'object' && value !== null) {
			const isArray = Array.isArray(value);
			// Recurse on each key object.
			const trimmedValue = isArray
				? trimArray(value as unknown[], `${path}.${key}`)
				: trimConfig(value as Record<string, unknown>, `${path}.${key}`);

			// Notify/WebHook objects may be empty to reference mains.
			// e.g. .service.*.notify | .service.*.webhook
			// e.g. .defaults.service.*.notify | .defaults.service.*.webhook
			const isKeepPath =
				containsEndsWith(path, ['.notify', '.webhook']) &&
				containsStartsWith(path, ['.service', '.defaults.service']);

			if (
				!(isArray
					? isEmptyArray(trimmedValue as unknown[])
					: isEmptyObject(trimmedValue as Record<string, unknown>)) ||
				isKeepPath
			) {
				acc[key] = trimmedValue;
			}
		} else {
			acc[key] = value;
		}
		return acc;
	}, {});
};

const trimArray = <T>(array: T[], path = ''): T[] => {
	// Empty, or contains only empty objects.
	if (
		isEmptyArray(array) ||
		array.every(
			(item) =>
				typeof item === 'object' &&
				item !== null &&
				isEmptyObject(item as Record<string, unknown>),
		)
	)
		return [];

	// Reduce each element of the array.
	return array.map((item, index) => {
		if (typeof item !== 'object' || item == null || Array.isArray(item))
			return item;

		return trimConfig(item as Record<string, unknown>, `${path}.${index}`) as T;
	});
};

/**
 * Orders the services in the `object` according to the `order` array.
 *
 * @param object - The object to order.
 * @param order - The ordering to apply.
 * @returns The ordered object, using the ordering of the order array.
 */
const orderServices = (
	object: NonNullable<ConfigType['service']>,
	order?: NonNullable<ConfigType['order']>,
): NonNullable<ConfigType['service']> => {
	if (!order) return object;
	const orderArray: string[] = Array.isArray(order)
		? order
		: Object.values(order);

	return orderArray.reduce<NonNullable<ConfigType['service']>>((acc, key) => {
		if (Object.hasOwn(object, key)) {
			acc[key] = object[key];
		}
		return acc;
	}, {});
};

/**
 * Orders the services in the `config` according to the order key.
 *
 * @param config - The configuration object.
 * @returns The configuration object with the services ordered, and the order key removed.
 */
const updateConfig = (config: ConfigType) => {
	const trimmedConfig = trimConfig(config) as ConfigType;
	if (!trimmedConfig.service || !trimmedConfig.order) return config;
	trimmedConfig.service = orderServices(
		trimmedConfig.service,
		trimmedConfig.order,
	);
	delete trimmedConfig.order;

	return trimmedConfig;
};

/**
 * Reducer function for the config endpoint.
 * - Removes empty objects from the config.
 * - Orders the services in the config.
 * - Converts the config to YAML.
 *
 * @param data - The data from the config endpoint.
 * @returns The YAML string representation of the config.
 */
export const configReducer = (data?: ConfigGetResponse) =>
	data ? stringify(updateConfig(data)) : '';
