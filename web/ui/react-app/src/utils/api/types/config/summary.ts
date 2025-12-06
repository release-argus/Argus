import type { LatestVersionLookupType } from '@/utils/api/types/config/service/latest-version';

export type OrderAPIResponse = {
	order: string[];
};

export type MonitorSummaryType = {
	order: string[];
	names: Set<string>;
	tags?: Set<string>;
	tagsLoaded: boolean;
	service: Record<string, ServiceSummary>;
};

export type ServiceSummary = {
	active?: boolean;
	id: string;
	name?: string;
	loading?: boolean;
	type?: LatestVersionLookupType;
	url?: string;
	icon?: string;
	icon_link_to?: string;
	has_deployed_version?: boolean;
	notify?: boolean;
	webhook?: number;
	command?: number;
	status?: StatusSummaryType;
	tags?: string[];
};

export type ServiceModal = {
	actionType: ModalType;
	service: ServiceSummary;
};

export type ModalType =
	| 'EDIT'
	| 'RESEND'
	| 'RETRY'
	| 'SEND'
	| 'SKIP'
	| 'SKIP_NO_WH'
	| '';

export type ActionModalData = {
	service_id: string;
	sentC: string[];
	sentWH: string[];
	webhooks: WebHookSummaryListType;
	commands: CommandSummaryListType;
};

export type StatusSummaryType = {
	approved_version?: string;
	deployed_version?: string;
	deployed_version_timestamp?: string;
	latest_version?: string;
	latest_version_timestamp?: string;
	last_queried?: string;
};

export type WebHookSummaryType = {
	// undefined = unsent/sending.
	failed?: boolean;
	next_runnable?: string;

	// Waiting for webhook status.
	loading?: boolean;
};

export type WebHookSummaryListType = Record<string, WebHookSummaryType>;

export type CommandSummaryType = {
	// undefined = unsent/sending.
	failed?: boolean;
	next_runnable?: string;

	// Waiting for command status.
	loading?: boolean;
};

export type CommandSummaryListType = Record<string, CommandSummaryType>;

export type ActionAPIType = {
	command: CommandSummaryListType;
	webhook: WebHookSummaryListType;
};
