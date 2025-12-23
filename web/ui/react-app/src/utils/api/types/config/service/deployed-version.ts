import type { Headers } from '@/utils/api/types/config/shared';

export const DEPLOYED_VERSION_LOOKUP_TYPE = {
	MANUAL: { label: 'Manual', value: 'manual' },
	URL: { label: 'URL', value: 'url' },
} as const;
export type DeployedVersionLookupType =
	(typeof DEPLOYED_VERSION_LOOKUP_TYPE)[keyof typeof DEPLOYED_VERSION_LOOKUP_TYPE]['value'];
export const deployedVersionLookupTypeOptions = Object.values(
	DEPLOYED_VERSION_LOOKUP_TYPE,
);

export type DeployedVersionLookup =
	| DeployedVersionLookupManual
	| DeployedVersionLookupURL;

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
