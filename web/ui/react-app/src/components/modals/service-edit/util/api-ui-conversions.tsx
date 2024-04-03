import {
  HeaderType,
  NotifyNtfyAction,
  NotifyOpsGenieTarget,
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
 * @param otherOptionsData - The other options data, containingglobals/defaults/hardDefaults
 * @returns The converted service data for use in the UI
 */
export const convertAPIServiceDataEditToUI = (
  name: string,
  serviceData?: ServiceEditAPIType,
  otherOptionsData?: ServiceEditOtherData
): ServiceEditType => {
  if (serviceData && name)
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
          command: serviceData?.latest_version?.require?.command?.map(
            (arg) => ({
              arg: arg as string,
            })
          ),
          docker: {
            ...serviceData?.latest_version?.require?.docker,
            type: serviceData?.latest_version?.require?.docker?.type ?? "",
          },
        },
      },
      deployed_version: {
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
        ? Object.entries(serviceData.webhook).reduce(
            (acc: WebHookEditType[], [_key, value]) => {
              // Determine webhook name and type
              const whName = value.name ?? "";
              const whType = value.type ?? "";

              // Construct custom headers
              const customHeaders = !isEmptyArray(value.custom_headers)
                ? value.custom_headers?.map((header, index) => ({
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
                  ).map(() => ({ key: "", value: "" }));

              const transformedWebhook = {
                ...value,
                custom_headers: customHeaders,
                oldIndex: whName,
              };
              return [...acc, transformedWebhook];
            },
            []
          )
        : [],
      notify: serviceData.notify
        ? Object.entries(serviceData.notify).reduce(
            (acc: NotifyEditType[], [_key, value]) => {
              // Determine notify name and type
              const notifyName = value.name ?? "";
              const notifyType = value.type ?? "";

              const transformedNotify = {
                ...value,
                id: notifyName,
                url_fields: convertNotifyURLFields(
                  notifyName,
                  notifyType,
                  value.url_fields,
                  otherOptionsData
                ),
                params: {
                  avatar: "", // controlled param
                  color: "", // ^
                  icon: "", // ^
                  ...convertNotifyParams(
                    notifyName,
                    notifyType,
                    value.params,
                    otherOptionsData
                  ),
                },
              };
              return [...acc, transformedNotify];
            },
            []
          )
        : [],
      dashboard: {
        auto_approve: undefined,
        icon: "",
        ...serviceData?.dashboard,
      },
    };

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
      auto_approve: undefined,
      icon: "",
      icon_link_to: "",
      web_url: "",
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
    list = JSON.parse(s as string);
    list = Array.isArray(list) ? list : [s as string];
  } catch (error) {
    list = [s as string];
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

  const usingStr = str ? true : false;

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

  const usingStr = str ? true : false;

  // convert from a JSON string
  try {
    return JSON.parse(s).map(
      (
        obj: { id: string; type: string; name: string; username: string },
        i: number
      ) => {
        const id = usingStr ? { id: i } : {};
        // team/user - id
        if (obj.id) {
          return {
            ...id,
            type: obj.type,
            sub_type: "id",
            value: usingStr ? obj.id : "",
          };
        } else {
          // team/user - username/name
          return {
            ...id,
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

  const usingStr = str ? true : false;

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
  type?: string,
  urlFields?: StringStringMap,
  otherOptionsData?: ServiceEditOtherData
) => {
  const notifyType = type || otherOptionsData?.notify?.[name]?.type || name;

  // Generic
  if (notifyType === "generic")
    return {
      ...urlFields,
      custom_headers: convertHeadersFromString(
        urlFields?.custom_headers,
        firstNonDefault(
          otherOptionsData?.notify?.[name]?.urlFields?.custom_headers,
          otherOptionsData?.defaults?.notify?.[notifyType]?.urlFields
            ?.custom_headers,
          otherOptionsData?.hard_defaults?.notify?.[notifyType]?.urlFields
            ?.custom_headers
        )
      ),
      json_payload_vars: convertHeadersFromString(
        urlFields?.json_payload_vars,
        firstNonDefault(
          otherOptionsData?.notify?.[name]?.urlFields?.json_payload_vars,
          otherOptionsData?.defaults?.notify?.[notifyType]?.urlFields
            ?.json_payload_vars,
          otherOptionsData?.hard_defaults?.notify?.[notifyType]?.urlFields
            ?.json_payload_vars
        )
      ),
      query_vars: convertHeadersFromString(
        // urlFields.query_vars,
        firstNonDefault(
          otherOptionsData?.notify?.[name]?.urlFields?.query_vars,
          otherOptionsData?.defaults?.notify?.[notifyType]?.urlFields
            ?.query_vars,
          otherOptionsData?.hard_defaults?.notify?.[notifyType]?.urlFields
            ?.query_vars
        )
      ),
    };

  return urlFields;
};

/**
 * Returns the converted notify.X.params for the UI
 *
 * @param name - The react-hook-form path to the notify object
 * @param type - The type of notify
 * @param urlFields - The params object to convert
 * @param otherOptionsData - The other options data, containing globals/defaults/hardDefaults
 * @returns The converted Params for use in the UI
 */
export const convertNotifyParams = (
  name: string,
  type?: string,
  params?: StringStringMap,
  otherOptionsData?: ServiceEditOtherData
) => {
  const notifyType = type || otherOptionsData?.notify?.[name]?.type || name;

  switch (notifyType) {
    // NTFY
    case "ntfy":
      return {
        ...params,
        actions: convertNtfyActionsFromString(
          params?.actions,
          firstNonDefault(
            otherOptionsData?.notify?.[name]?.params?.actions,
            otherOptionsData?.defaults?.notify?.[notifyType]?.params?.actions,
            otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params
              ?.actions
          )
        ),
      };

    // OpsGenie
    case "opsgenie":
      return {
        ...params,
        actions: convertStringToFieldArray(
          params?.actions,
          firstNonDefault(
            otherOptionsData?.notify?.[name]?.params?.actions,
            otherOptionsData?.defaults?.notify?.[notifyType]?.params?.actions,
            otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params
              ?.actions
          )
        ),
        details: convertHeadersFromString(
          params?.details,
          firstNonDefault(
            otherOptionsData?.notify?.[name]?.params?.details,
            otherOptionsData?.defaults?.notify?.[notifyType]?.params?.details,
            otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params
              ?.details
          )
        ),
        responders: convertOpsGenieTargetFromString(
          params?.responders,
          firstNonDefault(
            otherOptionsData?.notify?.[name]?.params?.responders,
            otherOptionsData?.defaults?.notify?.[notifyType]?.params
              ?.responders,
            otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params
              ?.responders
          )
        ),
        visibleto: convertOpsGenieTargetFromString(
          params?.visibleto,
          firstNonDefault(
            otherOptionsData?.notify?.[name]?.params?.visibleto,
            otherOptionsData?.defaults?.notify?.[notifyType]?.params?.visibleto,
            otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params
              ?.visibleto
          )
        ),
      };

    // Slack
    case "slack":
      return {
        ...params,
        // Remove hashtag from hex
        color: (params?.color ?? "").replace("%23", "#").replace("#", ""),
      };

    // Other
    default:
      return params;
  }
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
