import {
	ArgType,
	DeployedVersionLookupEditType,
	LatestVersionLookupEditType,
	NotifyEditType,
	ServiceEditType,
} from 'types/service-edit';
import {
	DeployedVersionLookupType,
	Dict,
	LatestVersionLookupType,
	NotifyTypesValues,
	ServiceType,
	WebHookType,
} from 'types/config';

import { convertValuesToString } from './notify-string-string-map';
import { isEmptyArray } from 'utils';
import removeEmptyValues from 'utils/remove-empty-values';
import { urlCommandTrim } from './url-command-trim';

/**
 * Converts the service data for the API.
 *
 * @param data - The service to convert.
 * @returns The formatted service to send to the API.
 */
export const convertUIServiceDataEditToAPI = (
	data: ServiceEditType,
): ServiceType => {
	const payload: ServiceType = {
		id: data.id,
		name: data.name,
		comment: data.comment,
	};

	// Options
	payload.options = {
		active: data.options?.active,
		interval: data.options?.interval,
		semantic_versioning: data.options?.semantic_versioning,
	};

	// Latest version
	payload.latest_version = convertUILatestVersionDataEditToAPI(
		data.latest_version,
	);

	// Deployed version - omit if no url set.
	payload.deployed_version = convertUIDeployedVersionDataEditToAPI(
		data.deployed_version,
	);

	// Command
	if (!isEmptyArray(data?.command))
		payload.command = data.command?.map((item) => item.args.map((a) => a.arg));

	// WebHook
	if (data.webhook)
		payload.webhook = data.webhook.reduce((acc, webhook) => {
			webhook = removeEmptyValues(webhook);
			// Defaults used if key/value empty.
			const removeCustomHeaders = (webhook.custom_headers ?? []).find(
				(header) => header.key === '' || header.value === '',
			);
			acc[webhook.name as string] = {
				...webhook,
				custom_headers: removeCustomHeaders
					? undefined
					: webhook.custom_headers,
				desired_status_code:
					webhook?.desired_status_code !== undefined
						? Number(webhook?.desired_status_code)
						: undefined,
				max_tries:
					webhook.max_tries !== undefined
						? Number(webhook.max_tries)
						: undefined,
			};
			return acc;
		}, {} as Dict<WebHookType>);

	// Notify
	if (data.notify)
		payload.notify = data.notify.reduce((acc, notify) => {
			acc[notify.name as string] = convertNotifyToAPI(notify);
			return acc;
		}, {} as Dict<NotifyTypesValues>);

	// Dashboard
	payload.dashboard = {
		auto_approve: data.dashboard?.auto_approve,
		icon: data.dashboard?.icon,
		icon_link_to: data.dashboard?.icon_link_to,
		web_url: data.dashboard?.web_url,
		tags: data.dashboard?.tags,
	};

	return payload;
};

/**
 * Converts the notify object for the API.
 *
 * @param notify - The notify object to convert.
 * @returns The notify object with string values and any empty values removed.
 */
export const convertNotifyToAPI = (notify: NotifyEditType) => {
	notify = removeEmptyValues(notify) as NotifyEditType;
	if (notify?.url_fields)
		notify.url_fields = convertValuesToString(notify.url_fields, notify.type);
	if (notify?.params)
		notify.params = convertValuesToString(notify.params, notify.type);

	return notify as NotifyTypesValues;
};

/**
 * Converts the latest_version for the API.
 *
 * @param data - The latest_version to convert.
 * @returns The latest_version in API format.
 */
export const convertUILatestVersionDataEditToAPI = (
	data?: LatestVersionLookupEditType,
): LatestVersionLookupType | null => {
	let converted: LatestVersionLookupType = {
		type: data?.type,
		url: data?.url,
		url_commands: data?.url_commands?.map((command) => ({
			...urlCommandTrim(command, true),
			index: command.index ? Number(command.index) : null,
		})),
	};

	// Type specific fields.
	switch (data?.type) {
		case 'github':
			converted.access_token = data.access_token ?? '';
			converted.use_prerelease = data.use_prerelease;
			break;
		case 'url':
			converted.allow_invalid_certs = data.allow_invalid_certs;
			break;
	}

	// Latest version - Require
	converted.require = {
		regex_content: data?.require?.regex_content ?? '',
		regex_version: data?.require?.regex_version ?? '',
		command: (data?.require?.command ?? []).map((obj) => (obj as ArgType).arg),
		docker: {
			type: data?.require?.docker?.type ?? '',
			image: data?.require?.docker?.image,
			tag: data?.require?.docker?.tag ?? '',
			username: data?.require?.docker?.username ?? '',
			token: data?.require?.docker?.token ?? '',
		},
	};

	return converted;
};

/**
 * Converts the deployed_version for the API.
 *
 * @param data - The deployed_version to convert.
 * @returns The deployed_version in API format.
 */
export const convertUIDeployedVersionDataEditToAPI = (
	data?: DeployedVersionLookupEditType,
): DeployedVersionLookupType | null => {
	let converted: DeployedVersionLookupType = {
		method: data?.method,
		url: data?.url,
		allow_invalid_certs: data?.allow_invalid_certs,
		headers: data?.headers ?? [],
		json: data?.json ?? '',
		regex: data?.regex ?? '',
		regex_template: data?.regex_template ?? '',
	};

	// Method - POST
	if (data?.method === 'POST') {
		converted.body = data?.body ?? '';
	}

	// Basic Auth
	converted.basic_auth = {
		username: data?.basic_auth?.username ?? '',
		password: data?.basic_auth?.password ?? '',
	};

	return converted;
};
