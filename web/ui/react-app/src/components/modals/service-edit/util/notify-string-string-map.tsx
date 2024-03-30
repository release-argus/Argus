import {
  HeaderType,
  NotifyNtfyAction,
  NotifyOpsGenieTarget,
  StringFieldArray,
} from "types/config";

interface StringAnyMap {
  [key: string]:
    | string
    | number
    | boolean
    | undefined
    | HeaderType[]
    | NotifyNtfyAction[]
    | NotifyOpsGenieTarget[]
    | StringFieldArray;
}
interface StringStringMap {
  [key: string]: string;
}

/**
 * Returns a properly formatted string of the notify.(params|url_fields) for the API
 *
 * @param obj - The object to convert
 * @returns The object with string values
 */
export const convertValuesToString = (obj: StringAnyMap): StringStringMap =>
  Object.entries(obj).reduce((result, [key, value]) => {
    if (typeof value === "object") {
      // opsGenie.responders || opsGenie.visibleto
      if ("responders" === key || "visibleto" === key) {
        // `value` empty means defaults were used. Skip.
        if (
          (value as NotifyOpsGenieTarget[]).find(
            (item) => (item.value || "") === ""
          )
        ) {
          return result;
        }
        // convert to string
        result[key] = convertOpsGenieTargetToString(
          value as NotifyOpsGenieTarget[]
        );
        // ntfy.actions
      } else if (key === "actions") {
        // `label` empty means defaults were used. Skip.
        if (
          (value as NotifyNtfyAction[]).find((item) => (item.label || "") == "")
        ) {
          return result;
        }
        // convert to string
        result[key] = convertNtfyActionsToString(value as NotifyNtfyAction[]);
        // opsGenie.details
      } else {
        // `value` empty means defaults were used. Skip.
        if (
          (value as NotifyOpsGenieTarget[]).find(
            (item) => (item.value ?? "") === ""
          )
        ) {
          return result;
        }
        result[key] = JSON.stringify(
          (value as HeaderType[]).reduce(
            (acc, { key, value }) => ({ ...acc, [key]: value }),
            {}
          )
        );
      }
    } else {
      result[key] = String(value);
    }
    return result;
  }, {} as StringStringMap);

/**
 * Returns a JSON string of the Ntfy actions for the API
 *
 * @param obj - The NotifyNtfyAction[] to convert
 * @returns A JSON string of the actions
 */
const convertNtfyActionsToString = (obj: NotifyNtfyAction[]): string =>
  JSON.stringify(
    obj.map((item) => {
      if (item.action === "view")
        return {
          action: item.action,
          label: item.label,
          url: item.url,
        };
      // http - headers as {KEY:VAL}, not {key:KEY, val:VAL}
      else if (item.action === "http")
        return {
          action: item.action,
          label: item.label,
          url: item.url,
          method: item.method,
          headers: item.headers,
          body: item.body,
        };
      // broadcast - extras as {KEY:VAL}, not {key:KEY, val:VAL}
      else if (item.action === "broadcast")
        return {
          action: item.action,
          label: item.label,
          intent: item.intent,
          extras: item.extras,
        };
      else return item;
    })
  );

/**
 * Returns a JSON string of the OpsGenie targets for the API
 *
 * @param obj - The NotifyOpsGenieTarget[] to convert
 * @returns A JSON string of the targets
 */
const convertOpsGenieTargetToString = (obj: NotifyOpsGenieTarget[]): string =>
  JSON.stringify(
    obj.map(({ type, sub_type, value }) => ({
      type: type,
      [sub_type]: value,
    }))
  );
