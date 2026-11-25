import type { Headers } from '@/utils/api/types/config/shared';

export const DEPLOYED_VERSION_LOOKUP_TYPE = {
	MANUAL: { label: 'Manual', value: 'manual' },
	URL: { label: 'URL', value: 'url' },
	FILE: { label: 'File', value: 'file' },
} as const;
export type DeployedVersionLookupType =
	(typeof DEPLOYED_VERSION_LOOKUP_TYPE)[keyof typeof DEPLOYED_VERSION_LOOKUP_TYPE]['value'];
export const deployedVersionLookupTypeOptions = Object.values(
	DEPLOYED_VERSION_LOOKUP_TYPE,
);

/**
 * Whether the deployed version lookup has the field its type reads the version from.
 *
 * @param dv - The deployed version lookup.
 * @returns `path` for 'file', `version` for 'manual', and `url` for 'url' is non-empty.
 */
export const hasDeployedVersionSource = (
	dv:
		| {
				type?: string | null;
				path?: string | null;
				url?: string | null;
				version?: string | null;
		  }
		| null
		| undefined,
): boolean => {
	switch (dv?.type) {
		case DEPLOYED_VERSION_LOOKUP_TYPE.FILE.value:
			return !!dv.path;
		case DEPLOYED_VERSION_LOOKUP_TYPE.MANUAL.value:
			return !!dv.version;
		case DEPLOYED_VERSION_LOOKUP_TYPE.URL.value:
			return !!dv.url;
		default:
			return false;
	}
};

export type DeployedVersionLookup =
	| DeployedVersionLookupFile
	| DeployedVersionLookupManual
	| DeployedVersionLookupURL;

export type DeployedVersionLookupFile = {
	type: typeof DEPLOYED_VERSION_LOOKUP_TYPE.FILE.value | null;
	path?: string;
	regex?: string;
	regex_template?: string;
};

/* Type: manual */
export type DeployedVersionLookupManual = {
	type: typeof DEPLOYED_VERSION_LOOKUP_TYPE.MANUAL.value | null;
	version?: string;
};

/* Type: url */
export type BasicAuthType = {
	username: string;
	password: string;
};

export const DEPLOYED_VERSION_LOOKUP__URL_METHOD = {
	GET: { label: 'GET', value: 'GET' },
	POST: { label: 'POST', value: 'POST' },
} as const;
export type DeployedVersionLookupURLMethod =
	(typeof DEPLOYED_VERSION_LOOKUP__URL_METHOD)[keyof typeof DEPLOYED_VERSION_LOOKUP__URL_METHOD]['value'];
export const deployedVersionLookupURLMethodOptions = Object.values(
	DEPLOYED_VERSION_LOOKUP__URL_METHOD,
);
export type DeployedVersionLookupURL = {
	type: typeof DEPLOYED_VERSION_LOOKUP_TYPE.URL.value | null;
	method?: DeployedVersionLookupURLMethod;
	url?: string;
	allow_invalid_certs?: boolean | null;
	basic_auth?: BasicAuthType;
	headers?: Headers;
	body?: string;
	target_header?: string;
	json?: string;
	regex?: string;
	regex_template?: string;
};
