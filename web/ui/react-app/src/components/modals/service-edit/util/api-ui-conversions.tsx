import {
  HeaderType,
  NonNullable,
  NotifyNtfyAction,
  NotifyOpsGenieTarget,
  NotifyTypes,
  NotifyTypesKeys,
  StringFieldArray,
  StringStringMap,
  WebHookType,
} from "types/config";
import {
  NotifyEditType,
  ServiceEditAPIType,
  ServiceEditOtherData,
  ServiceEditType,
  WebHookEditType,
} from "types/service-edit";
import {
  firstNonDefault,
  firstNonEmpty,
  isEmptyArray,
  isEmptyOrNull,
} from "utils";

import { urlCommandsTrimArray } from "./url-command-trim";

/**
 * Returns the converted service data for the UI
 *
 * @param name - The name of the service
 * @param serviceData - The service data from the API
 * @param otherOptionsData - The other options data, containing globals/defaults/hardDefaults
 * @returns The converted service data for use in the UI
 */
export const convertAPIServiceDataEditToUI = (
  name: string,
  serviceData?: ServiceEditAPIType,
  otherOptionsData?: ServiceEditOtherData
): ServiceEditType => {
  if (!serviceData || !name)
    // New service defaults
    return {
      name: "",
      options: { active: true },
      latest_version: {
        type: "github",
        require: { docker: { type: "" } },
      },
      deployed_version: {
        headers: [],
      },
      command: [],
      webhook: [],
      notify: [],
      dashboard: {
        icon: "",
        icon_link_to: "",
        web_url: "",
      },
    };

  // Edit service defaults
  return {
    ...serviceData,
    name: name,
    options: {
      ...serviceData?.options,
      active: serviceData?.options?.active !== false,
    },
    latest_version: {
      ...serviceData?.latest_version,
      url_commands:
        serviceData?.latest_version?.url_commands &&
        urlCommandsTrimArray(serviceData.latest_version.url_commands),
      require: {
        ...serviceData?.latest_version?.require,
        command: serviceData?.latest_version?.require?.command?.map((arg) => ({
          arg: arg as string,
        })),
        docker: {
          ...serviceData?.latest_version?.require?.docker,
          type: serviceData?.latest_version?.require?.docker?.type ?? "",
        },
      },
    },
    deployed_version: {
      method: "GET",
      ...serviceData?.deployed_version,
      basic_auth: {
        username: serviceData?.deployed_version?.basic_auth?.username ?? "",
        password: serviceData?.deployed_version?.basic_auth?.password ?? "",
      },
      headers:
        serviceData?.deployed_version?.headers?.map((header, key) => ({
          ...header,
          oldIndex: key,
        })) ?? [],
      template_toggle: !isEmptyOrNull(
        serviceData?.deployed_version?.regex_template
      ),
    },
    command: serviceData?.command?.map((args) => ({
      args: args.map((arg) => ({ arg })),
    })),
    webhook: serviceData.webhook
      ? serviceData.webhook.map((item) => {
          // Determine webhook name and type
          const whName = item.name as string;
          const whType = (item.type ??
            otherOptionsData?.webhook?.[whName]?.type ??
            whName) as NonNullable<WebHookType["type"]>;

          // Construct custom headers
          const customHeaders = !isEmptyArray(item.custom_headers)
            ? item.custom_headers?.map((header, index) => ({
                ...header,
                oldIndex: index,
              }))
            : firstNonEmpty(
                otherOptionsData?.webhook?.[whName]?.custom_headers,
                (
                  otherOptionsData?.defaults?.webhook?.[whType] as
                    | WebHookType
                    | undefined
                )?.custom_headers,
                (
                  otherOptionsData?.hard_defaults?.webhook?.[whType] as
                    | WebHookType
                    | undefined
                )?.custom_headers
              ).map(() => ({ key: "", item: "" }));

          return {
            ...item,
            oldIndex: whName,
            type: whType,
            custom_headers: customHeaders,
          } as WebHookEditType;
        })
      : [],
    notify: serviceData.notify
      ? serviceData.notify.map((item) => {
          // Determine notify name and type
          const notifyName = item.name as string;
          const notifyType = (item.type ||
            otherOptionsData?.notify?.[notifyName]?.type ||
            notifyName) as NotifyTypesKeys;

          return {
            ...item,
            oldIndex: notifyName,
            type: notifyType,
            url_fields: convertNotifyURLFields(
              notifyName,
              notifyType,
              item.url_fields,
              otherOptionsData
            ),
            params: {
              avatar: "", // controlled param
              color: "", // ^
              icon: "", // ^
              ...convertNotifyParams(
                notifyName,
                notifyType,
                item.params,
                otherOptionsData
              ),
            },
          } as NotifyEditType;
        })
      : [],
    dashboard: {
      icon: "",
      ...serviceData?.dashboard,
    },
  };
};

/**
 * Returns the converted field array for the UI
 *
 * (If defaults are provided and str is undefined/empty, it will only return only empty fields)
 *
 * @param str - JSON list or string to convert
 * @param defaults - The defaults
 * @param key - key to use for the object
 * @returns The converted object for use in the UI
 */
export const convertStringToFieldArray = (
  str?: string,
  defaults?: string,
  key = "arg"
): StringFieldArray | undefined => {
  // already converted
  if (typeof str === "object") return str;
  if (!str && typeof defaults === "object") return defaults;

  // undefined/empty
  const s = str || defaults || "";
  if (s === "") return [];

  let list: string[];
  try {
    list = JSON.parse(s);
    list = Array.isArray(list) ? list : [s];
  } catch (error) {
    list = [s];
  }

  // map the []string to {arg: string} for the form
  if (!str) return list.map(() => ({ [key]: "" }));
  return list.map((arg: string) => ({ [key]: arg }));
};

/**
 * Returns the converted notify.X.headers for the UI
 *
 * (If defaults are provided and str is undefined/empty, it will only return only empty fields)
 *
 * @param str - JSON to convert
 * @param defaults - The defaults
 * @returns The converted object for use in the UI
 */
export const convertHeadersFromString = (
  str?: string | HeaderType[],
  defaults?: string | HeaderType[]
): HeaderType[] => {
  // already converted
  if (typeof str === "object") return str;
  if (!str && typeof defaults === "object") return defaults;

  // undefined/empty
  const s = (str || defaults || "") as string;
  if (s === "") return [];

  const usingStr = !!str;

  // convert from a JSON string
  try {
    return Object.entries(JSON.parse(s)).map(([key, value], i) => {
      const id = usingStr ? { id: i } : {};
      return {
        ...id,
        key: usingStr ? key : "",
        value: usingStr ? value : "",
      };
    }) as HeaderType[];
  } catch (error) {
    return [];
  }
};

/**
 * Returns the converted notify.X.params.(responders|visibleto) for the UI
 *
 * (If defaults are provided and str is undefined/empty, it will only return the values in select fields)
 *
 * @param str - JSON to convert
 * @param defaults - The defaults
 * @returns The converted object for use in the UI
 */
export const convertOpsGenieTargetFromString = (
  str?: string | NotifyOpsGenieTarget[],
  defaults?: string | NotifyOpsGenieTarget[]
): NotifyOpsGenieTarget[] => {
  // already converted
  if (typeof str === "object") return str;
  if (!str && typeof defaults === "object") return defaults;

  // undefined/empty
  const s = (str || defaults || "") as string;
  if (s === "") return [];

  const usingStr = !!str;

  // convert from a JSON string
  try {
    return JSON.parse(s).map(
      (obj: { id?: string; type: string; name: string; username: string }) => {
        // team/user - id
        if (obj.id) {
          return {
            type: obj.type,
            sub_type: "id",
            value: usingStr ? obj.id : "",
          };
        } else {
          // team/user - username/name
          return {
            type: obj.type,
            sub_type: obj.type === "user" ? "username" : "name",
            value: usingStr ? obj.name || obj.username : "",
          };
        }
      }
    ) as NotifyOpsGenieTarget[];
  } catch (error) {
    return [];
  }
};

/**
 * Returns the converted notify.X.actions for the UI
 *
 * (If defaults are provided and str is undefined/empty, it will only return the values in select fields)
 *
 * @param str - JSON to convert
 * @param defaults - The defaults
 * @returns The converted object for use in the UI
 */
export const convertNtfyActionsFromString = (
  str?: string | NotifyNtfyAction[],
  defaults?: string | NotifyNtfyAction[]
): NotifyNtfyAction[] => {
  // already converted
  if (typeof str === "object") return str;
  if (!str && typeof defaults === "object") return defaults;

  // undefined/empty
  const s = (str || defaults || "") as string;
  if (s === "") return [];

  const usingStr = !!str;

  // convert from a JSON string
  try {
    return JSON.parse(s).map((obj: NotifyNtfyAction, i: number) => {
      const id = usingStr ? { id: i } : {};

      // View
      if (obj.action === "view")
        return {
          ...id,
          action: obj.action,
          label: usingStr ? obj.label : "",
          url: usingStr ? obj.url : "",
        };

      // HTTP
      if (obj.action === "http")
        return {
          ...id,
          action: obj.action,
          label: usingStr ? obj.label : "",
          url: usingStr ? obj.url : "",
          method: usingStr ? obj.method : "",
          headers: convertStringMapToHeaderType(
            obj.headers as StringStringMap,
            !usingStr
          ),
          body: obj.body,
        };

      // Broadcast
      if (obj.action === "broadcast")
        return {
          ...id,
          action: obj.action,
          label: usingStr ? obj.label : "",
          intent: usingStr ? obj.intent : "",
          extras: convertStringMapToHeaderType(
            obj.extras as StringStringMap,
            !usingStr
          ),
        };

      // Unknown action
      return {
        ...id,
        ...obj,
      };
    }) as NotifyNtfyAction[];
  } catch (error) {
    return [];
  }
};

/**
 * Returns the converted notify.X.url_fields for the UI
 *
 * @param name - The react-hook-form path to the notify object
 * @param type - The type of notify
 * @param urlFields - The url_fields object to convert
 * @param otherOptionsData - The other options data, containing globals/defaults/hardDefaults
 * @returns The converted URL Fields for use in the UI
 */
export const convertNotifyURLFields = (
  name: string,
  type: NotifyTypesKeys,
  urlFields?: StringStringMap,
  otherOptionsData?: ServiceEditOtherData
) => {
  // Generic
  if (type === "generic") {
    const main = otherOptionsData?.notify?.[name] as
      | NotifyTypes[typeof type]
      | undefined;
    return {
      ...urlFields,
      custom_headers: convertHeadersFromString(
        urlFields?.custom_headers,
        firstNonDefault(
          main?.url_fields?.custom_headers,
          otherOptionsData?.defaults?.notify?.[type]?.url_fields
            ?.custom_headers,
          otherOptionsData?.hard_defaults?.notify?.[type]?.url_fields
            ?.custom_headers
        )
      ),
      json_payload_vars: convertHeadersFromString(
        urlFields?.json_payload_vars,
        firstNonDefault(
          main?.url_fields?.json_payload_vars,
          otherOptionsData?.defaults?.notify?.[type]?.url_fields
            ?.json_payload_vars,
          otherOptionsData?.hard_defaults?.notify?.[type]?.url_fields
            ?.json_payload_vars
        )
      ),
      query_vars: convertHeadersFromString(
        urlFields?.query_vars,
        firstNonDefault(
          main?.url_fields?.query_vars,
          otherOptionsData?.defaults?.notify?.[type]?.url_fields?.query_vars,
          otherOptionsData?.hard_defaults?.notify?.[type]?.url_fields
            ?.query_vars
        )
      ),
    };
  }

  return urlFields;
};

/**
 * Returns the converted notify.X.params for the UI
 *
 * @param name - The react-hook-form path to the notify object
 * @param type - The type of notify
 * @param url_fields - The params object to convert
 * @param otherOptionsData - The other options data, containing globals/defaults/hardDefaults
 * @returns The converted Params for use in the UI
 */
export const convertNotifyParams = (
  name: string,
  type: NotifyTypesKeys,
  params?: StringStringMap,
  otherOptionsData?: ServiceEditOtherData
) => {
  switch (type) {
    case "bark":
    case "join":
    case "mattermost":
      return {
        icon: "", // controlled param
        ...params,
      };

    case "discord":
      return {
        avatar: "", // controlled param
        ...params,
      };

    // NTFY
    case "ntfy": {
      const main = otherOptionsData?.notify?.[name] as
        | NotifyTypes[typeof type]
        | undefined;
      return {
        icon: "", // controlled param
        ...params,
        actions: convertNtfyActionsFromString(
          params?.actions,
          firstNonDefault(
            main?.params?.actions,
            otherOptionsData?.defaults?.notify?.[type]?.params?.actions,
            otherOptionsData?.hard_defaults?.notify?.[type]?.params?.actions
          )
        ),
      };
    }
    // OpsGenie
    case "opsgenie": {
      const main = otherOptionsData?.notify?.[name] as
        | NotifyTypes[typeof type]
        | undefined;
      return {
        ...params,
        actions: convertStringToFieldArray(
          params?.actions,
          firstNonDefault(
            main?.params?.actions,
            otherOptionsData?.defaults?.notify?.[type]?.params?.actions,
            otherOptionsData?.hard_defaults?.notify?.[type]?.params?.actions
          )
        ),
        details: convertHeadersFromString(
          params?.details,
          firstNonDefault(
            main?.params?.details,
            otherOptionsData?.defaults?.notify?.[type]?.params?.details,
            otherOptionsData?.hard_defaults?.notify?.[type]?.params?.details
          )
        ),
        responders: convertOpsGenieTargetFromString(
          params?.responders,
          firstNonDefault(
            main?.params?.responders,
            otherOptionsData?.defaults?.notify?.[type]?.params?.responders,
            otherOptionsData?.hard_defaults?.notify?.[type]?.params?.responders
          )
        ),
        visibleto: convertOpsGenieTargetFromString(
          params?.visibleto,
          firstNonDefault(
            main?.params?.visibleto,
            otherOptionsData?.defaults?.notify?.[type]?.params?.visibleto,
            otherOptionsData?.hard_defaults?.notify?.[type]?.params?.visibleto
          )
        ),
      };
    }
    // Slack
    case "slack": {
      return {
        ...params,
        // Remove hashtag from hex
        color: (params?.color ?? "").replace("%23", "#").replace("#", ""),
      };
    }
  }

  // Other
  return params;
};

/**
 * Returns the headers in the format {key: KEY, value: VAL}[] for the UI
 *
 * @param headers - The {KEY:VAL, ...} object to convert
 * @param omitValues - If true, will omit the values from the object
 * @returns Converted headers, {key: KEY, value: VAL}[] for use in the UI
 */
const convertStringMapToHeaderType = (
  headers?: StringStringMap,
  omitValues?: boolean
): HeaderType[] => {
  if (!headers) return [];

  if (omitValues)
    return Object.keys(headers).map(() => ({ key: "", value: "" }));

  return Object.keys(headers).map((key) => ({
    key: key,
    value: headers[key],
  }));
};
