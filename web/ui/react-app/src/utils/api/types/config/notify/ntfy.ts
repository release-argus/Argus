export const NTFY_ACTION_TYPE = {
	BROADCAST: { label: 'Broadcast', value: 'broadcast' },
	HTTP: { label: 'HTTP', value: 'http' },
	VIEW: { label: 'View', value: 'view' },
} as const;
export type NtfyActionType =
	(typeof NTFY_ACTION_TYPE)[keyof typeof NTFY_ACTION_TYPE]['value'];
export const ntfyActionTypeOptions = Object.values(NTFY_ACTION_TYPE);

export const NTFY_SCHEME = {
	HTTP: { label: 'HTTP', value: 'http' },
	HTTPS: { label: 'HTTPS', value: 'https' },
} as const;
export type NTFYScheme =
	(typeof NTFY_SCHEME)[keyof typeof NTFY_SCHEME]['value'];
export const ntfySchemeOptions = Object.values(NTFY_SCHEME);

// biome-ignore assist/source/useSortedKeys: ascending order.
export const NTFY_PRIORITY = {
	MIN: { label: 'Min', value: 'min' },
	LOW: { label: 'Low', value: 'low' },
	DEFAULT: { label: 'Default', value: 'default' },
	HIGH: { label: 'High', value: 'high' },
	MAX: { label: 'Max', value: 'max' },
} as const;
export type NTFYPriority =
	(typeof NTFY_PRIORITY)[keyof typeof NTFY_PRIORITY]['value'];
export const ntfyPriorityOptions = Object.values(NTFY_PRIORITY);

export const NTFY_ACTION_HTTP_METHOD = {
	DELETE: { label: 'DELETE', value: 'delete' },
	GET: { label: 'GET', value: 'get' },
	PATCH: { label: 'PATCH', value: 'patch' },
	POST: { label: 'POST', value: 'post' },
	PUT: { label: 'PUT', value: 'put' },
};
export type NTFYActionHTTPMethod =
	(typeof NTFY_ACTION_HTTP_METHOD)[keyof typeof NTFY_ACTION_HTTP_METHOD]['value'];
export const ntfyActionHTTPMethodOptions = Object.values(
	NTFY_ACTION_HTTP_METHOD,
);
