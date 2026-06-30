import type { Headers } from '@/utils/api/types/config/shared';

export const WEBHOOK_TYPE = {
	GITHUB: { label: 'GitHub', value: 'github' },
	GITLAB: { label: 'GitLab', value: 'gitlab' },
} as const;
export type WebHookType =
	(typeof WEBHOOK_TYPE)[keyof typeof WEBHOOK_TYPE]['value'];
export const webhookTypeOptions = Object.values(WEBHOOK_TYPE);
export const isWebHookType = (value?: string | null): value is WebHookType =>
	value != null && webhookTypeOptions.some((v) => v.value === value);

export type WebHook = {
	name: string;

	type?: WebHookType | null;
	url?: string;
	allow_invalid_certs?: boolean | null;
	headers?: Headers;
	secret?: string;
	desired_status_code?: number;
	delay?: string;
	max_tries?: number;
	silent_fails?: boolean;
};

export type WebHookMap = Record<string, WebHook>;
