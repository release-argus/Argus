import {
	convertUIDeployedVersionDataEditToAPI,
	convertUILatestVersionDataEditToAPI,
} from 'components/modals/service-edit/util';

import { ArgType } from 'types/service-edit';
import { isEmptyObject } from './is-empty';
import isEmptyOrNull from './is-empty-or-null';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type DiffObject = { [key: string]: any };

/**
 * The differences between the two objects.
 *
 * @param newObj - The new object to compare.
 * @param oldObj - The old object to compare.
 * @returns The differences between the two objects.
 */
export const deepDiff = (
	newObj: DiffObject,
	oldObj?: DiffObject,
): DiffObject => {
	const diff: DiffObject = {};

	// If oldObj empty/undefined/null, return newObj,
	// e.g. DeployedVersion has no defaults.
	if (isEmptyOrNull(oldObj)) return newObj;

	// get all keys from both objects.
	const keys = Object.keys(oldObj).concat(Object.keys(newObj));

	keys.forEach((key) => {
		// Skip same values.
		if ((newObj[key] ?? '') === (oldObj[key] ?? '')) return;

		// Diff arrays.
		if (Array.isArray(oldObj[key]) || Array.isArray(newObj[key])) {
			// If array lengths differ, include all elements in difference.
			if ((oldObj[key] ?? []).length !== (newObj[key] ?? []).length) {
				diff[key] = newObj[key];
				// Else, recurse on each element.
			} else {
				const subDiff = (oldObj[key] ?? []).map(
					(elem: DiffObject, i: string | number) =>
						deepDiff(newObj[key][i], elem),
				);
				// Add to diff if any element has changed.
				if (subDiff.some((diffElem: DiffObject) => !isEmptyObject(diffElem)))
					diff[key] = newObj[key];
			}
			// Diff objects.
		} else if (
			typeof oldObj[key] === 'object' &&
			typeof newObj[key] === 'object'
		) {
			// recurse on nested objects.
			const subDiff = deepDiff(newObj[key], oldObj[key]);
			// add to diff if any nested object has changed.
			if (!isEmptyObject(subDiff)) diff[key] = subDiff;
			// diff scalars.
		} else if (oldObj[key] !== newObj[key]) diff[key] = newObj[key];
	});

	return diff;
};

/**
 * Encode the given key/value pair into a query param string.
 *
 * @param key - The key of the query param.
 * @param value - The value of the query param.
 * @param omitUndefined - Whether to omit undefined values.
 * @returns An encoded query param string for a given key/value pair.
 */
export const stringifyQueryParam = (
	key: string,
	value?: string | number | boolean | null,
	omitUndefined?: boolean,
) =>
	omitUndefined && value == null
		? ''
		: `${key}=${encodeURIComponent(value ?? '')}`;

/**
 * Converts an object to query params
 *
 * @param params - The object to convert to query params.
 * @returns A string of query params joined with '&' and prefixed with '?' (if not empty).
 */
export const convertToQueryParams = (
	params: Record<string, string | number | boolean | null | undefined>,
) => {
	const queryParams = Object.keys(params)
		.filter((key) => (params[key] ?? '') !== '') // Filter out empty strings.
		.map((key) => {
			return stringifyQueryParam(key, params[key]);
		})
		.join('&');
	return queryParams ? `?${queryParams}` : '';
};

/**
 * Get the differences between the two objects as query params.
 *
 * @param params - The new object.
 * @param defaults - The old object.
 * @param target - The target object to compare.
 * @returns The query params of any changed values between the two objects.
 */
export const getChanges = ({
	params,
	defaults,
	target,
}: {
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	params: any;
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	defaults?: any;
	target?: 'latest_version' | 'deployed_version';
}): string => {
	const convertedParams =
		target === 'latest_version'
			? convertUILatestVersionDataEditToAPI(params)
			: convertUIDeployedVersionDataEditToAPI(params);
	if (!convertedParams) return '';

	const changedParams = deepDiff(convertedParams, defaults);
	if (
		target === 'deployed_version' &&
		changedParams.hasOwnProperty('template_toggle')
	) {
		delete changedParams.template_toggle;
	}

	if (changedParams?.url_commands) {
		changedParams.url_commands = isEmptyObject(
			changedParams['url_commands']?.command,
		)
			? changedParams['url_commands']
			: {
					...changedParams['url_commands'],
					command: Object.values<ArgType>(changedParams.require.command).map(
						(value) => value.arg,
					),
			  };
	}

	// if no changes, return empty string.
	if (isEmptyObject(changedParams)) return '';

	// otherwise, stringify the changes.
	return JSON.stringify(changedParams, (_key, value) => {
		if (value === undefined) {
			return null;
		}
		return value;
	});
};
