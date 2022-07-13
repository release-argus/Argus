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
  active?: boolean;
  comment?: string;
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
  icon_link_to?: string;
  command?: string[];
  webhook?: ServiceDict<WebHookType>;
  notify?: ServiceDict<NotifyType>;
  deployed_version?: DeployedVersionLookupType;
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
