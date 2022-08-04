export interface ServiceDict<T> {
  [id: string]: T;
}

export interface ConfigState {
  data: ConfigType;
  waiting_on: string[];
}

export interface ConfigType {
  settings?: SettingsType;
  defaults?: DefaultsType;
  notify?: ServiceDict<NotifyType>;
  webhook?: ServiceDict<WebHookType>;
  service?: ServiceDict<ServiceType>;
  order?: string[];
}

export interface SettingsType {
  log?: LogSettingsType;
  web?: WebSettingsType;
}
export interface LogSettingsType {
  timestamps?: boolean;
  level?: string;
}

export interface DefaultsType {
  service?: ServiceType;
  notify?: ServiceDict<NotifyType>;
  webhook?: WebHookType;
}
export interface WebSettingsType {
  listen_host: string;
  listen_port: string;
  cert_file: string;
  pkey_file: string;
}

export interface ServiceListType {
  [id: string]: ServiceType;
}

export interface ServiceType {
  comment?: string;
  options?: ServiceOptionsType;
  latest_version?: LatestVersionLookupType;
  deployed_version?: DeployedVersionLookupType;
  command?: string[];
  webhook?: ServiceDict<WebHookType>;
  notify?: ServiceDict<NotifyType>;
  dashboard?: ServiceDashboardOptionsType;
}

export interface ServiceOptionsType {
  active?: boolean;
  interval?: string;
  semantic_versioning?: boolean;
}

export interface ServiceDashboardOptionsType {
  auto_approve?: boolean;
  icon?: string;
  icon_link_to?: string;
  web_url?: string;
}

export interface LatestVersionFilters {
  regex_content?: string;
  regex_version?: string;
}

export interface LatestVersionLookupType {
  type: string;
  url?: string;
  access_token?: string;
  allow_invalid_certs?: boolean;
  use_prerelease?: string;
  url_commands?: URLCommandsType[];
  require?: LatestVersionFilters;
}
export interface DeployedVersionLookupType {
  url?: string;
  allow_invalid_certs?: boolean;
  basic_auth?: BasicAuthType;
  headers?: HeaderType[];
  json?: string;
  regex?: string;
}

export interface BasicAuthType {
  username: string;
  password: string;
}

export interface HeaderType {
  key: string;
  value: string;
}

export interface URLCommandsType {
  type: string;
  regex?: string;
  index?: number;
  text?: string;
  old?: string;
  new?: string;
}

export interface NotifyType {
  type?: string;
  options?: OptionsType;
  url_fields?: Map<string, string>;
  params?: Map<string, string>;
}

export interface OptionsType {
  message?: string;
  delay?: string;
  max_tries?: number;
}

export interface WebHookType {
  type?: string;
  url?: string;
  allow_invalid_certs?: boolean;
  custom_headers?: Map<string, string>;
  secret?: string;
  desired_status_code?: number;
  delay?: string;
  max_tries?: number;
  silent_fails?: boolean;
}
