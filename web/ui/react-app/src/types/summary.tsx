export interface MonitorSummaryType {
  service: ServiceSummaryListType;
  order: string[];
}
export interface ServiceSummaryListType {
  [id: string]: ServiceSummaryType;
}

export interface ServiceSummaryType {
  active?: boolean;
  id: string;
  loading: boolean;
  type?: string;
  url?: string;
  icon?: string;
  icon_link_to?: string;
  has_deployed_version?: boolean;
  notify?: boolean;
  webhook?: number;
  command?: number;
  status?: StatusSummaryType;
}

export interface ServiceModal {
  actionType: ModalType;
  service: ServiceSummaryType;
}

export type ModalType =
  | "EDIT"
  | "RESEND"
  | "RETRY"
  | "SEND"
  | "SKIP"
  | "SKIP_NO_WH"
  | "";

export interface ActionModalData {
  service_id: string;
  sentC: string[];
  sentWH: string[];
  webhooks: WebHookSummaryListType;
  commands: CommandSummaryListType;
}

export interface StatusSummaryType {
  approved_version?: string;
  deployed_version?: string;
  deployed_version_timestamp?: string;
  latest_version?: string;
  latest_version_timestamp?: string;
  last_queried?: string;
}

export interface StatusFailsSummaryType {
  notify?: boolean;
  webhook?: boolean;
}

export interface WebHookSummaryType {
  // undefined = unsent/sending
  failed?: boolean;
  next_runnable?: string;
}

export interface WebHookSummaryListType {
  [id: string]: WebHookSummaryType;
}

export interface CommandSummaryType {
  // undefined = unsent/sending
  failed?: boolean;
  next_runnable?: string;
}

export interface CommandSummaryListType {
  [id: string]: CommandSummaryType;
}
