import {
  HeaderType,
  NotifyNtfyAction,
  NotifyOpsGenieTarget,
  StringStringMap,
} from "types/config";
import {
  ServiceEditAPIType,
  ServiceEditOtherData,
  ServiceEditType,
} from "types/service-edit";

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
      name: name,
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
      webhook: serviceData?.webhook?.map((item) => ({
        ...item,
        custom_headers: item.custom_headers?.map((header, index) => ({
          ...header,
          oldIndex: index,
        })),
        oldIndex: item.name,
      })),
      notify: serviceData?.notify?.map((item) => ({
        ...item,
        oldIndex: item.name,
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
 * Returns the converted headers for the UI
 *
 * @param str - JSON to convert
 * @param omitValues - If true, will omit the values from the object
 * @returns The converted object for use in the UI
 */
export const convertHeadersFromString = (
  str?: string | HeaderType[]
): HeaderType[] | undefined => {
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
    })) as HeaderType[];
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
  str?: string | NotifyOpsGenieTarget[]
): NotifyOpsGenieTarget[] | undefined => {
  // already converted
  if (typeof str === "object") return str;

  // undefined/empty
  if ((str ?? "") === "") return undefined;

  // convert from a JSON string
  try {
    return JSON.parse(str || "").map(
      (
        obj: { id: string; type: string; name: string; username: string },
        i: number
      ) => {
        // team/user - id
        if (obj.id) {
          return {
            id: i,
            type: obj.type,
            sub_type: "id",
            value: obj.id,
          };
        } else {
          // team/user - username/name
          return {
            id: i,
            type: obj.type,
            sub_type: obj.type === "user" ? "username" : "name",
            value: obj.name || obj.username,
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
  str?: string | NotifyNtfyAction[]
): NotifyNtfyAction[] | undefined => {
  // already converted
  if (typeof str === "object") return str;

  // undefined/empty
  if ((str ?? "") === "") return undefined;

  // convert from a JSON string
  try {
    return JSON.parse(str || "").map((obj: NotifyNtfyAction, i: number) => ({
      id: i,
      ...obj,
      headers: obj.headers
        ? convertStringMapToHeaderType(obj.headers as { [key: string]: string })
        : undefined,
      extras: obj.extras
        ? convertStringMapToHeaderType(obj.extras as { [key: string]: string })
        : undefined,
    })) as NotifyNtfyAction[];
  } catch (error) {
    return [];
  }
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
    case "ntfy":
      return {
        ...params,
        actions: convertNtfyActionsFromString(params?.actions),
      };
    case "opsgenie":
      return {
        ...params,
        details: convertHeadersFromString(params?.details),
        targets: convertOpsGenieTargetFromString(params?.targets),
      };
    case "generic":
      return {
        ...params,
        custom_headers: convertHeadersFromString(params?.custom_headers),
        json_payload_vars: convertHeadersFromString(params?.json_payload_vars),
        query_vars: convertHeadersFromString(params?.query_vars),
      };
    default:
      return params;
  }
};

const convertStringMapToHeaderType = (headers?: {
  [key: string]: string;
}): HeaderType[] | undefined => {
  if (!headers) return undefined;
  return Object.keys(headers).map((key) => ({
    key: key,
    value: headers[key],
  }));
};
