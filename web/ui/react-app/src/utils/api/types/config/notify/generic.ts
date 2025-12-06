export const GENERIC_REQUEST_METHODS = {
	CONNECT: { label: 'CONNECT', value: 'CONNECT' },
	DELETE: { label: 'DELETE', value: 'DELETE' },
	GET: { label: 'GET', value: 'GET' },
	HEAD: { label: 'HEAD', value: 'HEAD' },
	OPTIONS: { label: 'OPTIONS', value: 'OPTIONS' },
	POST: { label: 'POST', value: 'POST' },
	PUT: { label: 'PUT', value: 'PUT' },
	TRACE: { label: 'TRACE', value: 'TRACE' },
} as const;
export type GenericRequestMethod =
	(typeof GENERIC_REQUEST_METHODS)[keyof typeof GENERIC_REQUEST_METHODS]['value'];
export const genericRequestMethodOptions = Object.values(
	GENERIC_REQUEST_METHODS,
);
