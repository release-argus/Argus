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

import { firstNonDefault } from "components/modals/service-edit/notify-types/util";
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
          oldIndex: item.name,
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
 * @param omitValues - If true, will omit the values from the object
 * @param key - key to use for the object
 * @returns The converted list, or undefined if the input is empty
 */
export const convertStringToFieldArray = (
  str?: string,
  omitValues?: boolean,
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
  if (omitValues) return list.map(() => ({ [key]: "" }));
  return list.map((arg: string) => ({ [key]: arg }));
};

/**
 * convertHeadersFromString will convert a JSON string to a HeaderType[]
 * @param str - JSON to convert
 * @param omitValues - If true, will omit the values from the object
 * @returns
 */
export const convertHeadersFromString = (
  str?: string | NotifyHeaderType[],
  omitValues?: boolean
): NotifyHeaderType[] | undefined => {
  // already converted
  if (typeof str === "object") return str;
  // undefined/empty
  if (str === undefined || str === "") return undefined;

  // convert from a JSON string
  try {
    if (omitValues)
      return Object.entries(JSON.parse(str)).map(() => ({
        key: "",
        value: "",
      })) as NotifyHeaderType[];

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
 * (If defaults are provided and str is undefined/empty, it will only return the values in select fields)
 *
 * for opsgenie.responders and opsgenie.visibleto
 *
 * @param str - JSON to convert
 * @param defaults - The defaults
 * @returns The converted object, or undefined if the input is empty
 */
export const convertOpsGenieTargetFromString = (
  str?: string | NotifyOpsGenieTarget[],
  defaults?: string
): NotifyOpsGenieTarget[] | undefined => {
  // already converted
  if (typeof str === "object") return str;

  // undefined/empty
  if ((str === undefined || str === "") && (defaults ?? "") === "")
    return undefined;

  const usingStr = str ? true : false;

  // convert from a JSON string
  try {
    return JSON.parse(str || defaults || "").map(
      (
        obj: { id: string; type: string; name: string; username: string },
        i: number
      ) => {
        const id = usingStr ? i : undefined;
        // team/user - id
        if (obj.id) {
          return {
            id: id,
            type: obj.type,
            sub_type: "id",
            value: usingStr ? obj.id : "",
          };
        } else {
          // team/user - username/name
          return {
            id: id,
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
 * convertOpsGenieTargetToString will convert a JSON string to a NotifyNtfyAction[]
 *
 * (If defaults are provided and str is undefined/empty, it will only return the values in select fields)
 *
 * for ntify.actions
 *
 * @param str - JSON to convert
 * @param defaults - The defaults
 * @returns The converted object, or undefined if the input is empty
 */
export const convertNtfyActionsFromString = (
  str?: string | NotifyNtfyAction[],
  defaults?: string
): NotifyNtfyAction[] | undefined => {
  // already converted
  if (typeof str === "object") return str;

  // undefined/empty
  if ((str ?? "") === "" && (defaults ?? "") === "") return undefined;

  const usingStr = str ? true : false;

  // convert from a JSON string
  try {
    return JSON.parse(str || defaults || "").map(
      (obj: NotifyNtfyAction, i: number) => {
        const id = usingStr ? i : undefined;

        // View
        if (obj.action === "view")
          return {
            id: id,
            action: obj.action,
            label: usingStr ? obj.label : "",
            url: usingStr ? obj.url : "",
          };

        // HTTP
        if (obj.action === "http")
          return {
            id: id,
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
            id: id,
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
            firstNonDefault(
              otherOptionsData?.notify?.[name]?.urlFields?.custom_headers,
              otherOptionsData?.defaults?.notify?.[notifyType]?.urlFields
                ?.custom_headers,
              otherOptionsData?.hard_defaults?.notify?.[notifyType]?.urlFields
                ?.custom_headers
            ),
            true
          ),
      json_payload_vars: urlFields?.json_payload_vars
        ? convertHeadersFromString(urlFields.json_payload_vars)
        : convertHeadersFromString(
            firstNonDefault(
              otherOptionsData?.notify?.[name]?.urlFields?.json_payload_vars,
              otherOptionsData?.defaults?.notify?.[notifyType]?.urlFields
                ?.json_payload_vars,
              otherOptionsData?.hard_defaults?.notify?.[notifyType]?.urlFields
                ?.json_payload_vars
            ),
            true
          ),
      query_vars: urlFields?.query_vars
        ? convertHeadersFromString(urlFields.query_vars)
        : convertHeadersFromString(
            firstNonDefault(
              otherOptionsData?.notify?.[name]?.urlFields?.query_vars,
              otherOptionsData?.defaults?.notify?.[notifyType]?.urlFields
                ?.query_vars,
              otherOptionsData?.hard_defaults?.notify?.[notifyType]?.urlFields
                ?.query_vars
            ),
            true
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
      actions: convertNtfyActionsFromString(
        params?.actions,
        firstNonDefault(
          otherOptionsData?.notify?.[name]?.params?.actions,
          otherOptionsData?.defaults?.notify?.[notifyType]?.params?.actions,
          otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params?.actions
        )
      ),
    };

  // OpsGenie
  if (notifyType === "opsgenie")
    return {
      ...params,
      actions: params?.actions
        ? convertStringToFieldArray(params.actions)
        : convertStringToFieldArray(
            firstNonDefault(
              otherOptionsData?.notify?.[name]?.params?.actions,
              otherOptionsData?.defaults?.notify?.[notifyType]?.params?.actions,
              otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params
                ?.actions
            ),
            true
          ),
      details: params?.details
        ? convertHeadersFromString(params.details)
        : convertHeadersFromString(
            firstNonDefault(
              otherOptionsData?.notify?.[name]?.params?.details,
              otherOptionsData?.defaults?.notify?.[notifyType]?.params?.details,
              otherOptionsData?.hard_defaults?.notify?.[notifyType]?.params
                ?.details
            ),
            true
          ),
      responders: convertOpsGenieTargetFromString(
        params?.responders,
        firstNonDefault(
          otherOptionsData?.notify?.[name]?.params?.responders,
          otherOptionsData?.defaults?.notify?.[notifyType]?.params?.responders,
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
 * @param headers - The {KEY:VAL, ...} object to convert
 * @param omitValues - If true, will omit the values from the object
 * @returns Converted headers, {key: KEY, value: VAL}[] or undefined if the input is empty
 */
const convertStringMapToHeaderType = (
  headers?: StringStringMap,
  omitValues?: boolean
): HeaderType[] | undefined => {
  if (!headers) return undefined;
  if (omitValues)
    return Object.keys(headers).map((key) => ({ key, value: "" }));
  return Object.keys(headers).map((key) => ({
    key: key,
    value: headers[key],
  }));
};
