import type { CustomHeaders } from '@/utils/api/types/config/shared';

export const WEBHOOK_TYPE = {
	GITHUB: { label: 'GitHub', value: 'github' },
	GITLAB: { label: 'GitLab', value: 'gitlab' },
} as const;
export type WebHookType =
	(typeof WEBHOOK_TYPE)[keyof typeof WEBHOOK_TYPE]['value'];
export const webhookTypeOptions = Object.values(WEBHOOK_TYPE);
const webhookTypeValues: WebHookType[] = webhookTypeOptions.map(
	(option) => option.value,
);
export const isWebHookType = (key: string): key is WebHookType =>
	(webhookTypeValues as string[]).includes(key.toLowerCase());

export type WebHook = {
	name: string;

	type?: WebHookType;
	url?: string;
	allow_invalid_certs?: boolean | null;
	custom_headers?: CustomHeaders;
	secret?: string;
	desired_status_code?: number;
	delay?: string;
	max_tries?: number;
	silent_fails?: boolean;
};

export type WebHookMap = Record<string, WebHook>;
