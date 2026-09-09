import type { Headers } from '@/utils/api/types/config/shared';

export const WEBHOOK_TYPE = {
	GITHUB: { label: 'GitHub', value: 'github' },
	GITLAB: { label: 'GitLab', value: 'gitlab' },
} as const;
export type WebhookType =
	(typeof WEBHOOK_TYPE)[keyof typeof WEBHOOK_TYPE]['value'];
export const webhookTypeOptions = Object.values(WEBHOOK_TYPE);
export const isWebhookType = (value?: string | null): value is WebhookType =>
	value != null && webhookTypeOptions.some((v) => v.value === value);

export type Webhook = {
	name: string;

	type?: WebhookType | null;
	url?: string;
	allow_invalid_certs?: boolean | null;
	headers?: Headers;
	secret?: string;
	desired_status_code?: number;
	delay?: string;
	max_tries?: number;
	silent_fails?: boolean;
};

export type WebhookMap = Record<string, Webhook>;
