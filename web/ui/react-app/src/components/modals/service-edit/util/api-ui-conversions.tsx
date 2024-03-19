import {
  HeaderType,
  NotifyNtfyAction,
  NotifyOpsGenieTarget,
} from "types/config";
import {
  NotifyHeaderType,
  ServiceEditAPIType,
  ServiceEditOtherData,
  ServiceEditType,
  StringFieldArray,
} from "types/service-edit";

import { urlCommandsTrimArray } from "./url-command-trim";

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
            item.name as string,
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

// convertStringToFieldArray will convert a JSON string to a {[key]: string}[]
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

// convertOpsGenieTargetFromString will convert a JSON string to a NotifyOpsGenieTarget[]
// for opsgenie.responders and opsgenie.visibleto
export const convertOpsGenieTargetFromString = (
  str?: string | NotifyOpsGenieTarget[]
): NotifyOpsGenieTarget[] | undefined => {
  // already converted
  if (typeof str === "object") return str;
  // undefined/empty
  if (str === undefined || str === "") return undefined;

  // convert from a JSON string
  try {
    return JSON.parse(str).map(
      (
        obj: { id: string; type: string; name: string; username: string },
        i: number
      ) => {
        // id
        if (obj.id) {
          return {
            id: i,
            type: obj.type,
            sub_type: "id",
            value: obj.id,
          };
        } else {
          // username/name
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

// convertNtfyActionsFromString will convert a JSON string to a NotifyNtfyAction[]
// for ntfy.actions
export const convertNtfyActionsFromString = (
  str?: string | NotifyNtfyAction[]
): NotifyNtfyAction[] | undefined => {
  // already converted
  if (typeof str === "object") return str;
  if (str === undefined || str === "") return undefined;

  // convert from a JSON string
  try {
    return JSON.parse(str).map((obj: NotifyNtfyAction, i: number) => {
      if (obj.action === "view") {
        return {
          id: i,
          action: obj.action,
          label: obj.label,
          url: obj.url,
        };
      } else if (obj.action === "http") {
        return {
          id: i,
          action: obj.action,
          label: obj.label,
          url: obj.url,
          method: obj.method,
          headers: obj.headers
            ? convertStringMapToHeaderType(
                obj.headers as { [key: string]: string }
              )
            : undefined,
          body: obj.body,
        };
      } else if (obj.action === "broadcast") {
        return {
          id: i,
          action: obj.action,
          label: obj.label,
          intent: obj.intent,
          extras: obj.extras
            ? convertStringMapToHeaderType(
                obj.extras as { [key: string]: string }
              )
            : undefined,
        };
      }
      // unknown action
      return {
        id: i,
        ...obj,
      };
    }) as NotifyNtfyAction[];
  } catch (error) {
    return [];
  }
};

// convertNotifyParams will convert a notify param to the correct types for the UI
export const convertNotifyParams = (
  name: string,
  type?: string,
  params?: { [key: string]: string },
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
        actions: convertStringToFieldArray(params?.actions),
        details: convertHeadersFromString(params?.details),
        responders: convertOpsGenieTargetFromString(params?.responders),
        visibleto: convertOpsGenieTargetFromString(params?.visibleto),
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

// convertStringMapToHeaderType will convert a {[key]: string, ...} to a HeaderType[]
const convertStringMapToHeaderType = (headers?: {
  [key: string]: string;
}): HeaderType[] | undefined => {
  if (!headers) return undefined;
  return Object.keys(headers).map((key) => ({
    key: key,
    value: headers[key],
  }));
};
