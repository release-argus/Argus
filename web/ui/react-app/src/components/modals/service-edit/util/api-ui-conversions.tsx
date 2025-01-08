import {
	DeployedVersionLookupEditType,
	LatestVersionLookupEditType,
	NotifyEditAPIType,
	NotifyEditType,
	ServiceEditAPIType,
	ServiceEditOtherData,
	ServiceEditType,
	WebHookEditType,
} from 'types/service-edit';
import {
	HeaderType,
	NonNullable,
	NotifyNtfyAction,
	NotifyOpsGenieTarget,
	NotifyTypes,
	NotifyTypesKeys,
	StringFieldArray,
	StringStringMap,
	WebHookType,
} from 'types/config';
import {
	firstNonDefault,
	firstNonEmpty,
	isEmptyArray,
	isEmptyOrNull,
	strToBool,
} from 'utils';

import { urlCommandsTrimArray } from './url-command-trim';

/**
 * The converted service data for the UI.
 *
 * @param id - The ID of the service.
 * @param serviceData - The service data from the API.
 * @param otherOptionsData - The other options data, containing globals/defaults/hardDefaults.
 * @returns The converted service data for use in the UI.
 */
export const convertAPIServiceDataEditToUI = (
	id: string,
	serviceData?: ServiceEditAPIType,
	otherOptionsData?: ServiceEditOtherData,
): ServiceEditType => {
	if (!serviceData || !id)
		// New service defaults.
		return {
			id: '',
			name: '',
			comment: '',
			options: { active: true },
			latest_version: {
				type: 'github',
				require: { docker: { type: '' } },
			},
			deployed_version: {
				headers: [],
			},
			command: [],
			webhook: [],
			notify: [],
			dashboard: {
				icon: '',
				icon_link_to: '',
				web_url: '',
			},
		};

	// Edit service defaults.
	return {
		...serviceData,
		id: id,
		name: serviceData.name ?? '',
		options: convertAPIOptionsDataEditToUI(serviceData?.options),
		latest_version: convertAPILatestVersionDataEditToUI(
			serviceData?.latest_version,
		),
		deployed_version: convertAPIDeployedVersionDataEditToUI(
			serviceData?.deployed_version,
		),
		command: convertAPICommandDataEditToUI(serviceData?.command),
		webhook: convertAPIWebhookDataEditToUI(
			serviceData?.webhook,
			otherOptionsData,
		),
		notify: convertAPINotifyDataEditToUI(serviceData?.notify, otherOptionsData),
		dashboard: convertAPIDashboardDataEditToUI(serviceData?.dashboard),
	};
};

/**
 * The converted latest_version data for the UI.
 *
 * @param latest_version - The service.latest_version data from the API.
 * @returns The converted latest_version data for use in the UI.
 */
const convertAPILatestVersionDataEditToUI = (
	latest_version?: LatestVersionLookupEditType,
) => {
	const typeSpecific =
		latest_version?.type === 'github'
			? {
					use_prerelease: strToBool(latest_version?.use_prerelease),
			  }
			: {
					allow_invalid_certs: strToBool(latest_version?.allow_invalid_certs),
			  };

	return {
		...latest_version,
		...typeSpecific,
		url_commands:
			latest_version?.url_commands &&
			urlCommandsTrimArray(latest_version.url_commands),
		require: {
			...latest_version?.require,
			command: latest_version?.require?.command?.map((arg) => ({
				arg: arg as string,
			})),
			docker: {
				...latest_version?.require?.docker,
				type: latest_version?.require?.docker?.type ?? '',
			},
		},
	};
};

/**
 * The converted field array for the UI.
 *
 * (If defaults provided and str undefined/empty, it will return only empty fields).
 *
 * @param str - JSON list or string to convert.
 * @param defaults - The defaults.
 * @param key - key to use for the object.
 * @returns The converted object for use in the UI.
 */
export const convertStringToFieldArray = (
	str?: string,
	defaults?: string,
	key = 'arg',
): StringFieldArray | undefined => {
	// already converted.
	if (typeof str === 'object') return str;
	if (!str && typeof defaults === 'object') return defaults;

	// undefined/empty.
	const s = str || defaults || '';
	if (s === '') return [];

	let list: string[];
	try {
		list = JSON.parse(s);
		list = Array.isArray(list) ? list : [s];
	} catch (error) {
		list = [s];
	}

	// map the []string to {arg: string} for the form.
	if (!str) return list.map(() => ({ [key]: '' }));
	return list.map((arg: string) => ({ [key]: arg }));
};

/**
 * The converted notify.X.headers for the UI.
 *
 * (If defaults provided and str undefined/empty, it will return only empty fields).
 *
 * @param str - JSON to convert.
 * @param defaults - The defaults.
 * @returns The converted object for use in the UI.
 */
export const convertHeadersFromString = (
	str?: string | HeaderType[],
	defaults?: string | HeaderType[],
): HeaderType[] => {
	// already converted.
	if (typeof str === 'object') return str;
	if (!str && typeof defaults === 'object') return defaults;

	// undefined/empty.
	const s = (str || defaults || '') as string;
	if (s === '') return [];

	const usingStr = !!str;

	// convert from a JSON string.
	try {
		return Object.entries(JSON.parse(s)).map(([key, value], i) => {
			const id = usingStr ? { id: i } : {};
			return {
				...id,
				key: usingStr ? key : '',
				value: usingStr ? value : '',
			};
		}) as HeaderType[];
	} catch (error) {
		return [];
	}
};

/**
 * The converted notify.X.params.(responders|visibleto) for the UI.
 *
 * (If defaults provided and str undefined/empty, it will return the values in select fields).
 *
 * @param str - JSON to convert.
 * @param defaults - The defaults.
 * @returns The converted object for use in the UI.
 */
export const convertOpsGenieTargetFromString = (
	str?: string | NotifyOpsGenieTarget[],
	defaults?: string | NotifyOpsGenieTarget[],
): NotifyOpsGenieTarget[] => {
	// already converted.
	if (typeof str === 'object') return str;
	if (!str && typeof defaults === 'object') return defaults;

	// undefined/empty.
	const s = (str || defaults || '') as string;
	if (s === '') return [];

	const usingStr = !!str;

	// convert from a JSON string.
	try {
		return JSON.parse(s).map(
			(obj: { id?: string; type: string; name: string; username: string }) => {
				// team/user - id.
				if (obj.id) {
					return {
						type: obj.type,
						sub_type: 'id',
						value: usingStr ? obj.id : '',
					};
				} else {
					// team/user - username/name.
					return {
						type: obj.type,
						sub_type: obj.type === 'user' ? 'username' : 'name',
						value: usingStr ? obj.name || obj.username : '',
					};
				}
			},
		) as NotifyOpsGenieTarget[];
	} catch (error) {
		return [];
	}
};

/**
 * The converted notify.X.actions for the UI.
 *
 * (If defaults provided and str undefined/empty, it will return the values in select fields).
 *
 * @param str - JSON to convert.
 * @param defaults - The defaults.
 * @returns The converted object for use in the UI.
 */
export const convertNtfyActionsFromString = (
	str?: string | NotifyNtfyAction[],
	defaults?: string | NotifyNtfyAction[],
): NotifyNtfyAction[] => {
	// already converted.
	if (typeof str === 'object') return str;
	if (!str && typeof defaults === 'object') return defaults;

	// undefined/empty.
	const s = (str || defaults || '') as string;
	if (s === '') return [];

	const usingStr = !!str;

	// convert from a JSON string.
	try {
		return JSON.parse(s).map((obj: NotifyNtfyAction, i: number) => {
			const id = usingStr ? { id: i } : {};

			// View
			if (obj.action === 'view')
				return {
					...id,
					action: obj.action,
					label: usingStr ? obj.label : '',
					url: usingStr ? obj.url : '',
				};

			// HTTP
			if (obj.action === 'http')
				return {
					...id,
					action: obj.action,
					label: usingStr ? obj.label : '',
					url: usingStr ? obj.url : '',
					method: usingStr ? obj.method : '',
					headers: convertStringMapToHeaderType(
						obj.headers as StringStringMap,
						!usingStr,
					),
					body: obj.body,
				};

			// Broadcast
			if (obj.action === 'broadcast')
				return {
					...id,
					action: obj.action,
					label: usingStr ? obj.label : '',
					intent: usingStr ? obj.intent : '',
					extras: convertStringMapToHeaderType(
						obj.extras as StringStringMap,
						!usingStr,
					),
				};

			// Unknown action.
			return {
				...id,
				...obj,
			};
		}) as NotifyNtfyAction[];
	} catch (error) {
		return [];
	}
};

/**
 * The converted notify.X.url_fields for the UI.
 *
 * @param name - The react-hook-form path to the notify object.
 * @param type - The type of Notify.
 * @param urlFields - The url_fields object to convert.
 * @param otherOptionsData - The other options data, containing globals/defaults/hardDefaults.
 * @returns The converted URL Fields for use in the UI.
 */
export const convertNotifyURLFields = (
	name: string,
	type: NotifyTypesKeys,
	urlFields?: StringStringMap,
	otherOptionsData?: ServiceEditOtherData,
) => {
	// Generic
	if (type === 'generic') {
		const main = otherOptionsData?.notify?.[name] as
			| NotifyTypes[typeof type]
			| undefined;
		return {
			...urlFields,
			custom_headers: convertHeadersFromString(
				urlFields?.custom_headers,
				firstNonDefault(
					main?.url_fields?.custom_headers,
					otherOptionsData?.defaults?.notify?.[type]?.url_fields
						?.custom_headers,
					otherOptionsData?.hard_defaults?.notify?.[type]?.url_fields
						?.custom_headers,
				),
			),
			json_payload_vars: convertHeadersFromString(
				urlFields?.json_payload_vars,
				firstNonDefault(
					main?.url_fields?.json_payload_vars,
					otherOptionsData?.defaults?.notify?.[type]?.url_fields
						?.json_payload_vars,
					otherOptionsData?.hard_defaults?.notify?.[type]?.url_fields
						?.json_payload_vars,
				),
			),
			query_vars: convertHeadersFromString(
				urlFields?.query_vars,
				firstNonDefault(
					main?.url_fields?.query_vars,
					otherOptionsData?.defaults?.notify?.[type]?.url_fields?.query_vars,
					otherOptionsData?.hard_defaults?.notify?.[type]?.url_fields
						?.query_vars,
				),
			),
		};
	}

	return urlFields;
};

/**
 * The headers in the format {key: KEY, value: VAL}[], for the UI.
 *
 * @param headers - The {KEY:VAL, ...} object to convert.
 * @param omitValues - If true, will omit the values from the object.
 * @returns Converted headers, {key: KEY, value: VAL}[] for use in the UI.
 */
const convertStringMapToHeaderType = (
	headers?: StringStringMap,
	omitValues?: boolean,
): HeaderType[] => {
	if (!headers) return [];

	if (omitValues)
		return Object.keys(headers).map(() => ({ key: '', value: '' }));

	return Object.keys(headers).map((key) => ({
		key: key,
		value: headers[key],
	}));
};

/**
 * The converted deployed_version data for the UI.
 *
 * @param deployed_version - The service.deployed_version data from the API.
 * @returns The converted deployed_version data for use in the UI.
 */
const convertAPIDeployedVersionDataEditToUI = (
	deployed_version?: DeployedVersionLookupEditType,
): DeployedVersionLookupEditType => {
	return {
		method: 'GET',
		...deployed_version,
		allow_invalid_certs: strToBool(deployed_version?.allow_invalid_certs),
		basic_auth: {
			username: deployed_version?.basic_auth?.username ?? '',
			password: deployed_version?.basic_auth?.password ?? '',
		},
		headers: convertAPIHeadersDataEditToUI(deployed_version?.headers),
		template_toggle: !isEmptyOrNull(deployed_version?.regex_template),
	};
};

/**
 * The converted options data for the UI.
 *
 * @param options - The service.options data from the API.
 * @returns The converted options data for use in the UI.
 */
const convertAPIOptionsDataEditToUI = (
	options?: ServiceEditType['options'],
) => {
	return {
		...options,
		active: options?.active !== false,
		semantic_versioning: strToBool(options?.semantic_versioning),
	};
};

/**
 * The converted command data for the UI.
 *
 * @param command - The service.command data from the API.
 * @returns The converted command data for use in the UI.
 */
const convertAPICommandDataEditToUI = (command?: string[][]) => {
	if (!command) return [];
	return command.map((args) => ({
		args: args.map((arg) => ({ arg })),
	}));
};

/**
 * The converted webhook data for the UI.
 *
 * @param webhook - The service.webhook data from the API.
 * @param otherOptionsData - The other options data, containing globals/defaults/hardDefaults.
 * @returns The converted webhook data for use in the UI.
 */
const convertAPIWebhookDataEditToUI = (
	webhook?: WebHookType[],
	otherOptionsData?: ServiceEditOtherData,
) => {
	if (!webhook) return [];

	return webhook.map((item) => {
		// Determine webhook name and type.
		const whName = item.name as string;
		const whType = (item.type ??
			otherOptionsData?.webhook?.[whName]?.type ??
			whName) as NonNullable<WebHookType['type']>;

		// Construct custom headers.
		const customHeaders = !isEmptyArray(item.custom_headers)
			? item.custom_headers?.map((header, index) => ({
					...header,
					oldIndex: index,
			  }))
			: firstNonEmpty(
					otherOptionsData?.webhook?.[whName]?.custom_headers,
					(
						otherOptionsData?.defaults?.webhook?.[whType] as
							| WebHookType
							| undefined
					)?.custom_headers,
					(
						otherOptionsData?.hard_defaults?.webhook?.[whType] as
							| WebHookType
							| undefined
					)?.custom_headers,
			  ).map(() => ({ key: '', item: '' }));

		return {
			...item,
			oldIndex: whName,
			type: whType,
			custom_headers: customHeaders,
		} as WebHookEditType;
	});
};

/**
 * The converted notify data for the UI.
 *
 * @param notify - The service.notify data from the API.
 * @param otherOptionsData - The other options data, containing globals/defaults/hardDefaults.
 * @returns The converted notify data for use in the UI.
 */
const convertAPINotifyDataEditToUI = (
	notify?: NotifyEditAPIType[],
	otherOptionsData?: ServiceEditOtherData,
) => {
	if (!notify) return [];

	return notify.map((item) => {
		// Determine Notify name and type.
		const notifyName = item.name as string;
		const notifyType = (item.type ||
			otherOptionsData?.notify?.[notifyName]?.type ||
			notifyName) as NotifyTypesKeys;

		return {
			...item,
			oldIndex: notifyName,
			type: notifyType,
			url_fields: convertNotifyURLFields(
				notifyName,
				notifyType,
				item.url_fields,
				otherOptionsData,
			),
			params: {
				avatar: '', // controlled param.
				color: '', // ^
				icon: '', // ^
				...convertNotifyParams(
					notifyName,
					notifyType,
					item.params,
					otherOptionsData,
				),
			},
		} as NotifyEditType;
	});
};

/**
 * The converted notify.X.params for the UI.
 *
 * @param name - The react-hook-form path to the notify object.
 * @param type - The type of Notify.
 * @param params - The params object to convert.
 * @param otherOptionsData - The other options data, containing globals/defaults/hardDefaults.
 * @returns The converted Params for use in the UI.
 */
export const convertNotifyParams = (
	name: string,
	type: NotifyTypesKeys,
	params?: StringStringMap,
	otherOptionsData?: ServiceEditOtherData,
) => {
	switch (type) {
		case 'bark':
		case 'join':
		case 'mattermost':
			return {
				icon: '', // controlled param.
				...params,
			};

		case 'discord':
			return {
				avatar: '', // controlled param.
				...params,
			};

		// NTFY
		case 'ntfy': {
			const main = otherOptionsData?.notify?.[name] as
				| NotifyTypes[typeof type]
				| undefined;
			return {
				icon: '', // controlled param.
				...params,
				actions: convertNtfyActionsFromString(
					params?.actions,
					firstNonDefault(
						main?.params?.actions,
						otherOptionsData?.defaults?.notify?.[type]?.params?.actions,
						otherOptionsData?.hard_defaults?.notify?.[type]?.params?.actions,
					),
				),
			};
		}
		// OpsGenie
		case 'opsgenie': {
			const main = otherOptionsData?.notify?.[name] as
				| NotifyTypes[typeof type]
				| undefined;
			return {
				...params,
				actions: convertStringToFieldArray(
					params?.actions,
					firstNonDefault(
						main?.params?.actions,
						otherOptionsData?.defaults?.notify?.[type]?.params?.actions,
						otherOptionsData?.hard_defaults?.notify?.[type]?.params?.actions,
					),
				),
				details: convertHeadersFromString(
					params?.details,
					firstNonDefault(
						main?.params?.details,
						otherOptionsData?.defaults?.notify?.[type]?.params?.details,
						otherOptionsData?.hard_defaults?.notify?.[type]?.params?.details,
					),
				),
				responders: convertOpsGenieTargetFromString(
					params?.responders,
					firstNonDefault(
						main?.params?.responders,
						otherOptionsData?.defaults?.notify?.[type]?.params?.responders,
						otherOptionsData?.hard_defaults?.notify?.[type]?.params?.responders,
					),
				),
				visibleto: convertOpsGenieTargetFromString(
					params?.visibleto,
					firstNonDefault(
						main?.params?.visibleto,
						otherOptionsData?.defaults?.notify?.[type]?.params?.visibleto,
						otherOptionsData?.hard_defaults?.notify?.[type]?.params?.visibleto,
					),
				),
			};
		}
		// Slack
		case 'slack': {
			return {
				...params,
				// Remove hashtag from hex.
				color: (params?.color ?? '').replace('%23', '#').replace('#', ''),
			};
		}
	}

	// Other
	return params;
};

/**
 * The converted dashboard data for the UI.
 *
 * @param dashboard - The service.dashboard data from the API.
 * @returns The converted dashboard data for use in the UI.
 */
const convertAPIDashboardDataEditToUI = (
	dashboard?: ServiceEditType['dashboard'],
) => {
	return {
		icon: '',
		...dashboard,
		auto_approve: strToBool(dashboard?.auto_approve),
	};
};

/**
 * The converted headers data for the UI with oldIndex tracking.
 *
 * @param headers - The headers array from the API.
 * @returns The converted headers array with oldIndex for use in the UI.
 */
const convertAPIHeadersDataEditToUI = (headers?: HeaderType[]) => {
	if (!headers) return [];
	return headers.map((header, key) => ({
		...header,
		oldIndex: key,
	}));
};
