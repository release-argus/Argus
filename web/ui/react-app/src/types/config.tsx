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
  type: string;
  url?: string;
  allow_invalid_certs?: boolean;
  access_token?: string;
  semantic_versioning?: boolean;
  interval?: string;
  url_commands?: URLCommandsType[];
  regex_content?: string;
  regex_version?: string;
  use_prerelease?: string;
  web_url?: string;
  auto_approve?: boolean;
  ignore_misses?: string;
  icon?: string;
  notify?: ServiceDict<NotifyType>;
  webhook?: ServiceDict<WebHookType>;
  deployed_version?: DeployedVersionLookupType;
  status?: StatusType;
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

export interface StatusType {
  auto_approve?: string;
  deployed_version?: string;
  deployed_version_timestamp?: string;
  latest_version?: string;
  latest_version_timestamp?: string;
  last_queried?: number;
  regex_misses_content?: number;
  regex_misses_version?: number;
  service_misses?: string;
  fails: StatusFailsType;
}

export interface StatusFailsType {
  notify?: boolean[];
  webhook?: boolean[];
}

export interface URLCommandsType {
  type: string;
  regex?: string;
  index?: number;
  text?: string;
  old?: string;
  new?: string;
  ignore_misses?: string;
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
  type: string;
  url: string;
  secret?: string;
  desired_status_code?: number;
  delay?: string;
  max_tries?: number;
  silent_fails?: boolean;
}
