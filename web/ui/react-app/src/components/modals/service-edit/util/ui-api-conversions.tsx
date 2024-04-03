import { ArgType, NotifyEditType, ServiceEditType } from "types/service-edit";
import { Dict, NotifyType, ServiceType, WebHookType } from "types/config";

import { convertValuesToString } from "./notify-string-string-map";
import removeEmptyValues from "utils/remove-empty-values";
import { urlCommandTrim } from "./url-command-trim";

/**
 * Returns the properly formatted service data for the API
 *
 * @param data - The service to convert
 * @returns The formatted service to send to the API
 */
export const convertUIServiceDataEditToAPI = (
  data: ServiceEditType
): ServiceType => {
  const payload: ServiceType = {
    name: data.name,
    comment: data.comment,
  };

  // Options
  payload.options = {
    active: data.options?.active,
    interval: data.options?.interval,
    semantic_versioning: data.options?.semantic_versioning,
  };

  // Latest version
  payload.latest_version = {
    type: data.latest_version?.type,
    url: data.latest_version?.url,
    access_token: data.latest_version?.access_token,
    allow_invalid_certs: data.latest_version?.allow_invalid_certs,
    use_prerelease: data.latest_version?.use_prerelease,
    url_commands: data.latest_version?.url_commands?.map((command) => ({
      ...urlCommandTrim(command, true),
      index: command.index ? Number(command.index) : undefined,
    })),
  };
  // Latest version - Require
  if (data.latest_version?.require)
    payload.latest_version.require = {
      regex_content: data.latest_version.require?.regex_content,
      regex_version: data.latest_version.require?.regex_version,
      command: (data.latest_version.require.command ?? []).map(
        (obj) => (obj as ArgType).arg
      ),
      docker: {
        type: data.latest_version.require?.docker?.type,
        image: data.latest_version.require?.docker?.image,
        tag: data.latest_version.require?.docker?.tag,
        username: data.latest_version.require?.docker?.username,
        token: data.latest_version.require?.docker?.token,
      },
    };

  // Deployed version - omit if no url is set
  payload.deployed_version = data.deployed_version?.url
    ? {
        url: data.deployed_version?.url,
        allow_invalid_certs: data.deployed_version?.allow_invalid_certs,
        headers: data.deployed_version?.headers,
        json: data.deployed_version?.json,
        regex: data.deployed_version?.regex,
        regex_template: data.deployed_version?.regex_template,
        basic_auth: {
          username: data.deployed_version?.basic_auth?.username ?? "",
          password: data.deployed_version?.basic_auth?.password ?? "",
        },
      }
    : {};

  // Command
  if (data.command && data.command.length > 0)
    payload.command = data.command.map((item) => item.args.map((a) => a.arg));

  // WebHook
  if (data.webhook)
    payload.webhook = data.webhook.reduce((acc, webhook) => {
      webhook = removeEmptyValues(webhook);
      // Defaults were being shown if key/value were empty
      const removeCustomHeaders = (webhook.custom_headers ?? []).find(
        (header) => header.key === "" || header.value === ""
      );
      acc[webhook.name as string] = {
        ...webhook,
        custom_headers: removeCustomHeaders
          ? undefined
          : webhook.custom_headers,
        desired_status_code:
          webhook?.desired_status_code !== undefined
            ? Number(webhook?.desired_status_code)
            : undefined,
        max_tries:
          webhook.max_tries !== undefined
            ? Number(webhook.max_tries)
            : undefined,
      };
      return acc;
    }, {} as Dict<WebHookType>);

  // Notify
  if (data.notify)
    payload.notify = data.notify.reduce((acc, notify) => {
      acc[notify.name as string] = convertNotifyToAPI(notify);
      return acc;
    }, {} as Dict<NotifyType>);

  // Dashboard
  payload.dashboard = {
    auto_approve: data.dashboard?.auto_approve,
    icon: data.dashboard?.icon,
    icon_link_to: data.dashboard?.icon_link_to,
    web_url: data.dashboard?.web_url,
  };

  return payload;
};

/**
 * Returns a properly formatted notify object for the API
 *
 * @param notify - The notify object to convert
 * @returns The notify object with string values and any empty values removed
 */
export const convertNotifyToAPI = (notify: NotifyEditType) => {
  notify = removeEmptyValues(notify);
  if (notify?.url_fields)
    notify.url_fields = convertValuesToString(notify.url_fields, notify.type);
  if (notify?.params)
    notify.params = convertValuesToString(notify.params, notify.type);

  return notify as NotifyType;
};
