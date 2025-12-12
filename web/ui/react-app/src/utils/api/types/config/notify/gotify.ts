export const GOTIFY_EXTRA_NAMESPACE = {
	ANDROID_ACTION: { label: 'android::action', value: 'android::action' },
	CLIENT_DISPLAY: { label: 'client::display', value: 'client::display' },
	CLIENT_NOTIFICATION: {
		label: 'client::notification',
		value: 'client::notification',
	},
	OTHER: { label: 'other', value: 'other' },
} as const;
export type GotifyExtraNamespace =
	(typeof GOTIFY_EXTRA_NAMESPACE)[keyof typeof GOTIFY_EXTRA_NAMESPACE]['value'];
export const gotifyExtraNamespaceOptions = Object.values(
	GOTIFY_EXTRA_NAMESPACE,
);

export const GOTIFY_EXTRA__CLIENT_DISPLAY__CONTENT_TYPE = {
	TEXT_MARKDOWN: { label: 'text/markdown', value: 'text/markdown' },
	TEXT_PLAIN: { label: 'text/plain', value: 'text/plain' },
} as const;
export type GotifyExtraClientDisplayContentType =
	(typeof GOTIFY_EXTRA__CLIENT_DISPLAY__CONTENT_TYPE)[keyof typeof GOTIFY_EXTRA__CLIENT_DISPLAY__CONTENT_TYPE]['value'];
export const gotifyExtraClientDisplayContentTypeOptions = Object.values(
	GOTIFY_EXTRA__CLIENT_DISPLAY__CONTENT_TYPE,
);
