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
  gotify?: ServiceDict<GotifyType>;
  slack?: ServiceDict<SlackType>;
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
  gotify?: GotifyType;
  slack?: SlackType;
  webhook?: WebHookType;
}
export interface WebSettingsType {
  listen_address: string;
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
  gotify?: ServiceDict<GotifyType>;
  slack?: ServiceDict<SlackType>;
  webhook?: ServiceDict<WebHookType>;
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
  current_version?: string;
  current_version_timestamp?: string;
  latest_version?: string;
  latest_version_timestamp?: string;
  last_queried?: number;
  regex_misses_content?: number;
  regex_misses_version?: number;
  service_misses?: string;
  fails: StatusFailsType;
}

export interface StatusFailsType {
  gotify?: boolean[];
  slack?: boolean[];
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

export interface GotifyType {
  url: string;
  token?: string;
  title?: string;
  message?: string;
  extras?: ExtrasType;
  priority?: string;
  delay?: string;
  maxTries?: number;
}

export interface ExtrasType {
  android_action?: string;
  client_display?: string;
  client_notification?: string;
}

export interface SlackType {
  url: string;
  icon_emoji?: string;
  icon_url?: string;
  username?: string;
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
