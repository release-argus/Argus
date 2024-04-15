import {
  BasicAuthType,
  DefaultsType,
  Dict,
  DockerFilterType,
  HeaderType,
  NotifyOptionsType,
  NotifyType,
  NotifyTypes,
  ServiceDashboardOptionsType,
  ServiceOptionsType,
  StringStringMap,
  URLCommandType,
  WebHookType,
} from "./config";

export interface ServiceEditModalData {
  service?: ServiceEditType;
}

export interface ServiceEditOtherData {
  webhook?: Dict<WebHookType>;
  notify?: Dict<NotifyType>;
  defaults?: DefaultsType;
  hard_defaults?: DefaultsType;
}

export interface ServiceEditType {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any;
  comment?: string;
  options?: ServiceOptionsType;
  latest_version?: LatestVersionLookupEditType;
  deployed_version?: DeployedVersionLookupEditType;
  command?: EditCommandType[];
  webhook?: WebHookEditType[];
  notify?: NotifyEditType[];
  dashboard?: ServiceDashboardOptionsType;
}

export interface EditCommandType {
  args: ArgType[];
}

export interface ArgType {
  arg: string;
}

export interface ServiceRefreshType {
  version?: string;
  error?: string;
  timestamp: string;

  message?: string; // Route disabled message
}

export interface LatestVersionLookupEditType {
  [key: string]:
    | string
    | boolean
    | undefined
    | URLCommandType[]
    | LatestVersionFiltersEditType;
  type?: "github" | "url";
  url?: string;
  access_token?: string;
  allow_invalid_certs?: boolean;
  use_prerelease?: boolean;
  url_commands?: URLCommandType[];
  require?: LatestVersionFiltersEditType;
}
export interface LatestVersionFiltersEditType {
  [key: string]: string | string[] | ArgType[] | DockerFilterType | undefined;
  command?: ArgType[] | string[];
  docker?: DockerFilterType;
  regex_content?: string;
  regex_version?: string;
}

export interface DeployedVersionLookupEditType {
  [key: string]:
    | string
    | boolean
    | undefined
    | BasicAuthType
    | HeaderEditType[];
  url?: string;
  allow_invalid_certs?: boolean;
  basic_auth?: BasicAuthType;
  headers?: HeaderEditType[];
  json?: string;
  regex?: string;
}

export interface NotifyEditType extends NotifyType {
  oldIndex?: string;
}

export interface HeaderEditType extends HeaderType {
  oldIndex?: number; // Index of existing secret
}

export interface WebHookEditType extends WebHookType {
  oldIndex?: string; // Index of existing secret
}

/////////////////////////////////
//             API             //
/////////////////////////////////
export interface ServiceEditAPIType {
  comment?: string;
  options?: ServiceOptionsType;
  latest_version?: LatestVersionLookupEditType;
  deployed_version?: DeployedVersionLookupEditType;
  command?: string[][];
  webhook?: WebHookEditType[];
  notify?: NotifyEditAPIType[];
  dashboard?: ServiceDashboardOptionsType;
}

export interface NotifyEditAPIType {
  name?: string;

  type?: NotifyTypes;
  options?: NotifyOptionsType;
  url_fields?: StringStringMap;
  params?: StringStringMap;
}
