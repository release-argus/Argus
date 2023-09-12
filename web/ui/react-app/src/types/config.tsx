/* eslint-disable @typescript-eslint/no-explicit-any */
import {
  BarkSchemeOptions,
  BarkSoundOptions,
} from "components/modals/service-edit/notify-types/bark";
import {
  NtfyPriorityOptions,
  NtfySchemeOptions,
} from "components/modals/service-edit/notify-types/ntfy";
import {
  SMTPAuthOptions,
  SMTPEncryptionOptions,
} from "components/modals/service-edit/notify-types/smtp";

import { TelegramParseModeOptions } from "components/modals/service-edit/notify-types/telegram";

export interface Dict<T> {
  [id: string]: T;
}
export const DictToList = <T,>(dict: Dict<T>, giveIndexTo?: string[]): T[] => {
  return Object.entries(dict).map(([name, value]) => {
    let newValue = value as T & { name: string; oldIndex: string };
    newValue.name = name;
    newValue.oldIndex = name;
    giveIndexTo?.forEach((prop) => {
      if ((value as any)[prop] && Array.isArray((value as any)[prop]))
        newValue = {
          ...newValue,
          [prop]: (value as any)[prop].map((v: any, k: string) => ({
            ...v,
            oldIndex: k,
          })),
        };
    });
    return newValue;
  });
};

export interface ConfigState {
  data: ConfigType;
  waiting_on: string[];
}

export interface ConfigType {
  settings?: SettingsType;
  defaults?: DefaultsType;
  notify?: Dict<NotifyType>;
  webhook?: Dict<WebHookType>;
  service?: Dict<ServiceType>;
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
  service?: DefaultServiceType;
  notify?: Dict<NotifyType>;
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

export interface DefaultServiceType {
  [key: string]: any;
  options?: ServiceOptionsType;
  latest_version?: DefaultLatestVersionLookupType;
  deployed_version?: DeployedVersionLookupType;
  notify?: Dict<undefined>;
  command?: string[][];
  webhook?: Dict<undefined>;
  dashboard?: ServiceDashboardOptionsType;
}

export interface ServiceType {
  [key: string]: any;
  comment?: string;
  options?: ServiceOptionsType;
  latest_version?: LatestVersionLookupType;
  deployed_version?: DeployedVersionLookupType;
  command?: string[][];
  webhook?: Dict<WebHookType>;
  notify?: Dict<NotifyType>;
  dashboard?: ServiceDashboardOptionsType;
}

export interface ServiceOptionsType {
  [key: string]: string | boolean | undefined;
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
export interface DockerFilterType {
  [key: string]: string | undefined;
  type?: string;
  image?: string;
  tag?: string;
  username?: string;
  token?: string;
}

export interface BaseLatestVersionLookupType {
  [key: string]:
    | string
    | boolean
    | undefined
    | URLCommandType[]
    | LatestVersionFiltersType
    | DefaultLatestVersionFiltersType;
  type?: "github" | "url";
  url?: string;
  access_token?: string;
  allow_invalid_certs?: boolean;
  use_prerelease?: boolean;
  url_commands?: URLCommandType[];
}
export interface DefaultLatestVersionLookupType
  extends BaseLatestVersionLookupType {
  require?: DefaultLatestVersionFiltersType;
}

export interface LatestVersionLookupType extends BaseLatestVersionLookupType {
  require?: LatestVersionFiltersType;
}

export interface DefaultLatestVersionFiltersType {
  [key: string]: DefaultDockerFilterType | undefined;
  docker?: DefaultDockerFilterType;
}
export interface DefaultDockerFilterType {
  [key: string]: string | DefaultDockerFilterRegistryType | undefined;
  type?: string;
  ghcr?: DefaultDockerFilterRegistryType;
  hub?: DefaultDockerFilterRegistryType;
  quay?: DefaultDockerFilterRegistryType;
}
export interface DefaultDockerFilterRegistryType {
  [key: string]: string | undefined;
  token?: string;
  username?: string;
}

export interface LatestVersionFiltersType {
  [key: string]: string | CommandType | DockerFilterType | undefined;
  regex_content?: string;
  regex_version?: string;
  command?: CommandType;
  docker?: DockerFilterType;
}
export interface DeployedVersionLookupType {
  [key: string]: string | boolean | undefined | BasicAuthType | HeaderType[];
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

export type CommandType = string[];

export type URLCommandTypes = "regex" | "replace" | "split" | string;
export interface URLCommandType {
  [key: string]: string | URLCommandTypes | number | undefined;

  type: URLCommandTypes;
  regex?: string; // regex
  text?: string; // split
  index?: number; // split
  old?: string; // replace
  new?: string; // replace
}
export const NotifyTypesConst = [
  "bark",
  "discord",
  "smtp",
  "googlechat",
  "gotify",
  "ifttt",
  "join",
  "mattermost",
  "matrix",
  "ntfy",
  "opsgenie",
  "pushbullet",
  "pushover",
  "rocketchat",
  "slack",
  "teams",
  "telegram",
  "zulip",
];
export type NotifyTypes = (typeof NotifyTypesConst)[number];
export interface NotifyDefaults {
  barl: NotifyBarkType;
  discord: NotifyDiscordType;
  smtp: NotifySMTPType;
  googlechat: NotifyGoogleChatType;
  gotify: NotifyGotifyType;
  ifttt: NotifyIFTTTType;
  join: NotifyJoinType;
  mattermost: NotifyMatterMostType;
  matrix: NotifyMatrixType;
  ntfy: NotifyNtfyType;
  opsgenie: NotifyOpsGenieType;
  pushbullet: NotifyPushbulletType;
  pushover: NotifyPushoverType;
  rocketchat: NotifyRocketChatType;
  slack: NotifySlackType;
  teams: NotifyTeamsType;
  telegram: NotifyTelegramType;
  zulip: NotifyZulipType;
}
export interface NotifyType {
  [key: string]: any;
  id?: number;
  name?: string;

  type?: NotifyTypes;
  options?: NotifyOptionsType;
  url_fields?: { [key: string]: undefined | string | number | boolean };
  params?: {
    [key: string]:
      | undefined
      | string
      | number
      | boolean
      | NotifyNtfyAction[]
      | NotifyOpsGenieTarget[]
      | { [key: string]: string };
  };
}

export interface NotifyBarkType extends NotifyType {
  type: "bark";
  url_fields: {
    devicekey?: string;
    host?: string;
    port?: string;
    path?: string;
  };
  params: {
    badge?: number;
    copy?: string;
    group?: string;
    icon?: string;
    scheme?: (typeof BarkSchemeOptions)[number]["value"];
    sound?: (typeof BarkSoundOptions)[number]["value"];
    title?: string;
    url?: string;
  };
}

export interface NotifyDiscordType extends NotifyType {
  type: "discord";
  url_fields: {
    token?: string;
    webhookid?: string;
  };
  params: {
    avatar?: string;
    title?: string;
    username?: string;
    splitlines?: string;
  };
}
export interface NotifySMTPType extends NotifyType {
  type: "smtp";
  url_fields: {
    host?: string;
    password?: string;
    port?: number;
    username?: string;
  };
  params: {
    auth?: (typeof SMTPAuthOptions)[number]["value"];
    clienthost?: string;
    encryption?: (typeof SMTPEncryptionOptions)[number]["value"];
    fromaddress?: string;
    fromname?: string;
    subject?: string;
    toaddresses?: string;
    usehtml?: boolean | string;
    usestarttls?: boolean | string;
  };
}
export interface NotifyGoogleChatType extends NotifyType {
  type: "googlechat";
  url_fields: {
    raw?: string;
  };
}
export interface NotifyGotifyType extends NotifyType {
  type: "gotify";
  url_fields: {
    host?: string;
    port?: number;
    path?: string;
    token?: string;
  };
  params: {
    disabletls?: boolean | string;
    priority?: number;
    title?: string;
  };
}
export interface NotifyIFTTTType extends NotifyType {
  type: "ifttt";
  url_fields: {
    usemessageasvalue: string | number | undefined;
    webhookid?: string;
  };
  params: {
    events?: string;
    title?: string;
    usemessageasvalue?: number;
    usetitleasvalue?: number;
    value1?: string;
    value2?: string;
    value3?: string;
  };
}
export interface NotifyJoinType extends NotifyType {
  type: "join";
  url_fields: {
    apikey?: string;
  };
  params: {
    devices?: string;
    icon?: string;
    title?: string;
  };
}
export interface NotifyMatterMostType extends NotifyType {
  type: "mattermost";
  url_fields: {
    host?: string;
    password?: string;
    path?: string;
    port?: number;
    username?: string;
    token?: string;
    channel?: string;
  };
  params: {
    icon?: string;
  };
}
export interface NotifyMatrixType extends NotifyType {
  type: "matrix";
  url_fields: {
    host?: string;
    password?: string;
    port?: number;
    username?: string;
  };
  params: {
    disabletls?: boolean | string;
    rooms?: string;
    title?: string;
  };
}
export interface NotifyNtfyType extends NotifyType {
  type: "ntfy";
  url_fields: {
    host?: string;
    password?: string;
    port?: number;
    topic?: string;
    username?: string;
  };
  params: {
    actions?: string | NotifyNtfyAction[];
    attach?: string;
    cache?: boolean | string;
    click?: string;
    delay?: string;
    email?: string;
    filename?: string;
    firebase?: boolean | string;
    icon?: string;
    priority?: (typeof NtfyPriorityOptions)[number]["value"];
    scheme?: (typeof NtfySchemeOptions)[number]["value"];
    tags?: string;
    title?: string;
  };
}
export type NotifyNtfyActionTypes = "view" | "http" | "broadcast";
export interface NotifyNtfyAction {
  action: string;
  label: string;

  // view/http
  url?: string;

  // http
  method: string;
  headers?: HeaderType[] | { [key: string]: string };
  body?: string;

  // broadcast
  intent?: string;
  extras?: HeaderType[] | { [key: string]: string };
}

export interface NotifyOpsGenieType extends NotifyType {
  type: "opsgenie";
  url_fields: {
    apikey?: string;
    host?: string;
    port?: number;
  };
  params: {
    actions?: string;
    alias?: string;
    description?: string;
    details?: string | { [key: string]: string };
    entity?: string;
    note?: string;
    priority?: string;
    responders?: string | NotifyOpsGenieTarget[];
    source?: string;
    tags?: string;
    title?: string;
    user?: string;
    visibleto?: string | NotifyOpsGenieTarget[];
  };
}
// Format received from Argus
export interface NotifyOpsGenieTargetIncoming {
  [key: string]: undefined | string;
  type: string;
  id?: string;
  name?: string;
  username?: string;
}
export interface NotifyOpsGenieTarget {
  type: string;
  sub_type: string;
  value: string;
}
export interface NotifyPushbulletType extends NotifyType {
  type: "pushbullet";
  url_fields: {
    targets?: string;
    token?: string;
  };
  params: {
    title?: string;
  };
}
export interface NotifyPushoverType extends NotifyType {
  type: "pushover";
  url_fields: {
    token?: string;
    user?: string;
  };
  params: {
    devices?: string;
    priority?: number;
    title?: string;
  };
}
export interface NotifyRocketChatType extends NotifyType {
  type: "rocketchat";
  url_fields: {
    channel?: string;
    host?: string;
    path?: string;
    port?: number;
    tokena?: string;
    tokenb?: string;
    username?: string;
  };
}
export interface NotifySlackType extends NotifyType {
  type: "slack";
  url_fields: {
    channel?: string;
    token?: string;
  };
  params: {
    botname?: string;
    color?: string;
    icon?: string;
    title?: string;
  };
}
export interface NotifyTeamsType extends NotifyType {
  type: "teams";
  url_fields: {
    altid?: string;
    group?: string;
    groupowner: string;
    tenant?: string;
  };
  params: {
    color?: string;
    host?: string;
    title?: string;
  };
}
export interface NotifyTelegramType extends NotifyType {
  type: "telegram";
  url_fields: {
    token?: string;
  };
  params: {
    chats?: string;
    notification?: boolean | string;
    parsemode?: (typeof TelegramParseModeOptions)[number]["value"];
    preview?: boolean | string;
    title?: string;
  };
}
export interface NotifyZulipType extends NotifyType {
  type: "zulip";
  url_fields: {
    botmail?: string;
    botkey?: string;
    host?: string;
  };
  params: {
    stream?: string;
    topic?: string;
  };
}

export type NotifyGenericRequestMethods =
  | "OPTIONS"
  | "GET"
  | "HEAD"
  | "POST"
  | "PUT"
  | "DELETE"
  | "TRACE"
  | "CONNECT";
export interface NotifyGenericType extends NotifyType {
  type: "generic";
  url_fields: {
    host?: string;
    port?: number;
    path?: string;
  };
  params: {
    contenttype?: string;
    disabletls?: boolean;
    messagekey?: string;
    requestmethod?: string;
    template?: string;
    title?: string;
    titlekey?: string;
  };
}

export interface NotifyOptionsType {
  [key: string]: string | number | undefined;
  message?: string;
  delay?: string;
  max_tries?: number;
}

export interface WebHookType {
  // For edit
  [key: string]: string | boolean | number | undefined | HeaderType[];
  name?: string;

  type?: string;
  url?: string;
  allow_invalid_certs?: boolean;
  custom_headers?: HeaderType[];
  secret?: string;
  desired_status_code?: number;
  delay?: string;
  max_tries?: number;
  silent_fails?: boolean;
}
