import {
  HeaderType,
  NotifyHeaderType,
  NotifyNtfyAction,
  NotifyOpsGenieTarget,
  StringFieldArray,
  StringStringMap,
  WebHookType,
} from "types/config";
import {
  ServiceEditAPIType,
  ServiceEditOtherData,
  ServiceEditType,
} from "types/service-edit";

import { globalOrDefault } from "components/modals/service-edit/notify-types/util";
import { urlCommandsTrimArray } from "./url-command-trim";

/**
 * convertAPIServiceDataEditToUI will convert the API service data to the format needed for the web UI
 *
 * @param name - The name of the service
 * @param serviceData - The service data from the API
 * @param otherOptionsData - The other options data, containingglobals/defaults/hardDefaults
 * @returns The converted service data
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
        template_toggle:
          (serviceData?.deployed_version?.regex_template ?? "") !== "",
      },
      command: serviceData?.command?.map((args) => ({
        args: args.map((arg) => ({ arg })),
      })),
      webhook: serviceData?.webhook?.map((item) => {
        // Determine webhook name and type
        const whName = item.name ?? "";
        const whType = item.type ?? "";

        // Construct custom headers
        const customHeaders = item.custom_headers
          ? item.custom_headers.map((header, index) => ({
              ...header,
              oldIndex: index,
            }))
          : otherOptionsData?.webhook?.[whName]?.custom_headers ||
            (
              otherOptionsData?.defaults?.webhook?.[whType] as
                | WebHookType
                | undefined
            )?.custom_headers ||
            (
              otherOptionsData?.hard_defaults?.webhook?.[whType] as
                | WebHookType
                | undefined
            )?.custom_headers ||
            [];

        // Return modified item
        return {
          ...item,
          custom_headers: customHeaders,
          oldIndex: item.custom_headers ? whName : undefined,
        };
      }),
      notify: serviceData?.notify?.map((item) => ({
        ...item,
        oldIndex: item.name,
        url_fields: {
          ...convertNotifyURLFields(
            item.name ?? "",
            item.type,
            item.url_fields,
            otherOptionsData
          ),
        },
        params: {
          avatar: "", // controlled param
          color: "", // ^
          icon: "", // ^
          ...convertNotifyParams(
            item.name ?? "",
            item.type,
            item.params,
            otherOptionsData
          ),
        },
      })),
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
    dashboard: {
      auto_approve: undefined,
      icon: "",
      icon_link_to: "",
      web_url: "",
    },
  };
};

/**
 * convertStringToFieldArray will convert a JSON string to a {[key]: string}[]
 *
 * @param str - JSON list or string to convert
 * @param key - key to use for the object
 * @returns The converted list, or undefined if the input is empty
 */
export const convertStringToFieldArray = (
  str?: string,
  key = "arg"
): StringFieldArray | undefined => {
  if (str === undefined || str === "") return undefined;

  let list: string[];
  try {
    list = JSON.parse(str);
    list = Array.isArray(list) ? list : [str];
  } catch (error) {
    list = [str];
  }

  // map the []string to {arg: string} for the form
  return list.map((arg: string) => ({ [key]: arg }));
};

export const convertHeadersFromString = (
  str?: string | NotifyHeaderType[]
): NotifyHeaderType[] | undefined => {
  // already converted
  if (typeof str === "object") return str;
  // undefined/empty
  if (str === undefined || str === "") return undefined;

  // convert from a JSON string
  try {
    return Object.entries(JSON.parse(str)).map(([key, value], i) => ({
      id: i,
      key: key,
      value: value,
    })) as NotifyHeaderType[];
  } catch (error) {
    return [];
  }
};

/**
 * convertOpsGenieTargetFromString will convert a JSON string to a NotifyOpsGenieTarget[]
 *
 * (If global/defaults/hardDefaults are provided, it will only return the values in select fields)
 *
 * for opsgenie.responders and opsgenie.visibleto
 *
 * @param str - JSON to convert
 * @param global - The global defaults
 * @param defaults - The defaults
 * @param hardDefaults - The hard defaults
 * @returns The converted object, or undefined if the input is empty
 */
export const convertOpsGenieTargetFromString = (
  str?: string | NotifyOpsGenieTarget[],
  global?: string,
  defaults?: string,
  hardDefaults?: string
): NotifyOpsGenieTarget[] | undefined => {
  // already converted
  if (typeof str === "object") return str;

  const firstDefault = globalOrDefault(global, defaults, hardDefaults);

  // undefined/empty
  if ((str === undefined || str === "") && firstDefault === "")
    return undefined;

  // convert from a JSON string
  try {
    return JSON.parse(str || firstDefault).map(
      (
        obj: { id: string; type: string; name: string; username: string },
        i: number
      ) => {
        const id = firstDefault ? undefined : i;
        // team/user - id
        if (obj.id) {
          return {
            id: id,
            type: obj.type,
            sub_type: "id",
            value: firstDefault ? "" : obj.id,
          };
        } else {
          // team/user - username/name
          return {
            id: id,
            type: obj.type,
            sub_type: obj.type === "user" ? "username" : "name",
            value: firstDefault ? "" : obj.name || obj.username,
          };
        }
      }
    ) as NotifyOpsGenieTarget[];
  } catch (error) {
    return [];
  }
};

/**
 * convertOpsGenieTargetToString will convert a JSON string to a NotifyNtfyAction[]
 *
 * (If global/defaults/hardDefaults are provided, it will only return the values in select fields)
 *
 * for ntify.actions
 *
 * @param str - JSON to convert
 * @param global - The global defaults
 * @param defaults - The defaults
 * @param hardDefaults - The hard defaults
 * @returns The converted object, or undefined if the input is empty
 */
export const convertNtfyActionsFromString = (
  str?: string | NotifyNtfyAction[],
  global?: string,
  defaults?: string,
  hardDefaults?: string
): NotifyNtfyAction[] | undefined => {
  // already converted
  if (typeof str === "object") return str;

  const firstDefault = globalOrDefault(global, defaults, hardDefaults);

  // undefined/empty
  if ((str === undefined || str === "") && firstDefault === "")
    return undefined;

  // convert from a JSON string
  try {
    return JSON.parse(str || firstDefault).map(
      (obj: NotifyNtfyAction, i: number) => {
        const id = firstDefault ? undefined : i;

        // View
        if (obj.action === "view")
          return {
            id: id,
            action: obj.action,
            label: firstDefault ? "" : obj.label,
            url: firstDefault ? "" : obj.url,
          };

        // HTTP
        if (obj.action === "http")
          return {
            id: id,
            action: obj.action,
            label: firstDefault ? "" : obj.label,
            url: firstDefault ? "" : obj.url,
            method: firstDefault ? "" : obj.method,
            headers: firstDefault
              ? convertStringMapToHeaderType(
                  obj.headers as StringStringMap
                )?.map((item) => ({ ...item, key: "", value: "" }))
              : convertStringMapToHeaderType(obj.headers as StringStringMap),
            body: obj.body,
          };

        // Broadcast
        if (obj.action === "broadcast")
          return {
            id: id,
            action: obj.action,
            label: firstDefault ? "" : obj.label,
            intent: firstDefault ? "" : obj.intent,
            extras: firstDefault
              ? convertStringMapToHeaderType(
                  obj.extras as StringStringMap
                )?.map((item) => ({ ...item, key: "", value: "" }))
              : convertStringMapToHeaderType(obj.extras as StringStringMap),
          };

        // Unknown action
        return {
          id: id,
          ...obj,
        };
      }
    ) as NotifyNtfyAction[];
  } catch (error) {
    return [];
  }
};

/**
 * convertNotifyURLFields will convert a notify url_fields object to the correct types for the UI
 *
 * @param name - The react-hook-form path to the notify object
 * @param type - The type of notify
 * @param urlFields - The url_fields object to convert
 * @param otherOptionsData - The other options data, containing globals/defaults/hardDefaults
 * @returns The converted URL Fields
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
      custom_headers: urlFields?.custom_headers
        ? convertHeadersFromString(urlFields.custom_headers)
        : convertHeadersFromString(
            globalOrDefault(
              otherOptionsData?.notify?.[name]?.urlFields?.custom_headers,
              otherOptionsData?.defaults?.notify?.[notifyType]?.urlFields
                ?.custom_headers,
              otherOptionsData?.hard_defaults?.notify?.[notifyType]?.urlFields
                ?.custom_headers
            )
          ),
      json_payload_vars: urlFields?.json_payload_vars
        ? convertHeadersFromString(urlFields.json_payload_vars)
        : convertHeadersFromString(
            globalOrDefault(
              otherOptionsData?.notify?.[name]?.urlFields?.json_payload_vars,
              otherOptionsData?.defaults?.notify?.[notifyType]?.urlFields
                ?.json_payload_vars,
              otherOptionsData?.hard_defaults?.notify?.[notifyType]?.urlFields
                ?.json_payload_vars
            )
          ),
      query_vars: urlFields?.query_vars
        ? convertHeadersFromString(urlFields.query_vars)
        : convertHeadersFromString(
            globalOrDefault(
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
 * convertNotifyURLFields will convert a notify params object to the correct types for the UI
 *
 * @param name - The react-hook-form path to the notify object
 * @param type - The type of notify
 * @param urlFields - The params object to convert
 * @param otherOptionsData - The other options data, containing globals/defaults/hardDefaults
 * @returns The converted Params
 */
export const convertNotifyParams = (
  name: string,
  type?: string,
  params?: StringStringMap,
  otherOptionsData?: ServiceEditOtherData
) => {
  const notifyType = type || otherOptionsData?.notify?.[name]?.type || name;

  // NTFY
  if (notifyType === "ntfy")
    return {
      ...params,
      actions: params?.actions
        ? convertNtfyActionsFromString(params.actions)
        : convertNtfyActionsFromString(
            params?.actions,
            otherOptionsData?.notify?.[name]?.params?.actions as
              | string
              | undefined,
            otherOptionsData?.defaults?.notify?.[notifyType]?.params
              ?.actions as string | undefined,
            otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params
              ?.actions as string | undefined
          ),
    };

  // OpsGenie
  if (notifyType === "opsgenie")
    return {
      ...params,
      actions: params?.actions
        ? convertStringToFieldArray(params.actions)
        : convertStringToFieldArray(
            globalOrDefault(
              otherOptionsData?.notify?.[name]?.params?.actions,
              otherOptionsData?.defaults?.notify?.[notifyType]?.params?.actions,
              otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params
                ?.actions
            )
          ),
      details: params?.details
        ? convertHeadersFromString(params.details)
        : convertHeadersFromString(
            globalOrDefault(
              otherOptionsData?.notify?.[name]?.params?.details,
              otherOptionsData?.defaults?.notify?.[notifyType]?.params?.details,
              otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params
                ?.details
            )
          ),
      responders: params?.responders
        ? convertOpsGenieTargetFromString(params.responders)
        : convertOpsGenieTargetFromString(
            globalOrDefault(
              otherOptionsData?.notify?.[name]?.params?.responders,
              otherOptionsData?.defaults?.notify?.[notifyType]?.params
                ?.responders,
              otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params
                ?.responders
            )
          ),
      visibleto: params?.visibileto
        ? convertOpsGenieTargetFromString(params.visibleto)
        : convertOpsGenieTargetFromString(
            globalOrDefault(
              otherOptionsData?.notify?.[name]?.params?.visibleto,
              otherOptionsData?.defaults?.notify?.[notifyType]?.params
                ?.visibleto,
              otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params
                ?.visibleto
            )
          ),
    };

  if (notifyType === "slack")
    return {
      ...params,
      // Add # to the color if it's a hex code
      color: /^[\da-f]{6}$/i.test(params?.color ?? "")
        ? `#${params?.color}`
        : params?.color,
    };

  // Other
  return params;
};

/**
 * convertStringMapToHeaderType will convert a {[key]: string, ...} to a HeaderType[]
 *
 * @param headers The {KEY:VAL, ...} object to convert
 * @returns Converted headers, {key: KEY, value: VAL}[] or undefined if the input is empty
 */
const convertStringMapToHeaderType = (headers?: {
  [key: string]: string;
}): HeaderType[] | undefined => {
  if (!headers) return undefined;
  return Object.keys(headers).map((key) => ({
    key: key,
    value: headers[key],
  }));
};
