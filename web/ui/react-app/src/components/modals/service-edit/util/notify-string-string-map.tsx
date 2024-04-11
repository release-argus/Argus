import {
  HeaderType,
  NotifyNtfyAction,
  NotifyOpsGenieTarget,
  StringFieldArray,
} from "types/config";

import { isEmptyOrNull } from "utils";

interface StringAnyMap {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any;
}
interface StringStringMap {
  [key: string]: string;
}

/**
 * Returns a properly formatted string of the notify.(params|url_fields) for the API
 *
 * @param obj - The object to convert
 * @param notifyType - The type of Notify to convert
 * @returns The object with string values
 */
export const convertValuesToString = (
  obj: StringAnyMap,
  notifyType?: string
): StringStringMap =>
  Object.entries(obj).reduce((result, [key, value]) => {
    if (typeof value === "object") {
      // opsGenie.responders || opsGenie.visibleto
      if ("responders" === key || "visibleto" === key) {
        // `value` empty means defaults were used. Skip.
        if (
          (value as NotifyOpsGenieTarget[]).find((item) =>
            isEmptyOrNull(item.value)
          )
        ) {
          return result;
        }
        // convert to string
        result[key] = convertOpsGenieTargetToString(
          value as NotifyOpsGenieTarget[]
        );
        // (ntfy|opsgenie).actions
      } else if (key === "actions") {
        // Ntfy - `label` empty means defaults were used. Skip.
        // OpsGenie - `arg` empty means defaults were used. Skip.
        if (
          (value as StringFieldArray).find((item) =>
            isEmptyOrNull(item.label || item.arg)
          )
        ) {
          return result;
        }
        // convert to string
        result[key] =
          notifyType === "ntfy"
            ? convertNtfyActionsToString(value as NotifyNtfyAction[])
            : FlattenStringFieldArray(value as StringFieldArray);
        // opsGenie.details
      } else {
        // `value` empty means defaults were used. Skip.
        if (
          (value as NotifyOpsGenieTarget[]).find((item) =>
            isEmptyOrNull(item.value)
          )
        ) {
          return result;
        }
        result[key] = JSON.stringify(flattenHeaderArray(value as HeaderType[]));
      }
    } else {
      // Give # to slack hex colours
      if (notifyType === "slack" && key === "color") {
        result[key] = (
          /^[\da-f]{6}$/i.test(value as string) ? `#${value}` : value
        ) as string;

        // Convert to string
      } else result[key] = String(value);
    }
    return result;
  }, {} as StringStringMap);

/**
 * Returns a flattened JSON string of the object for the API
 *
 * @param obj - The StringFieldArray to flatten { arg: "value1" }[]
 * @returns A JSON string of the values ["value1", "value2", ...]
 */
const FlattenStringFieldArray = (obj: StringFieldArray): string =>
  JSON.stringify(obj.map((item) => Object.values(item)[0]));

/**
 * Returns a flattened object of headers for the API
 *
 * @param headers - The HeaderType[] to flatten { key: "KEY", value: "VAL" }[]
 * @returns The flattened object { KEY: VAL, ... }
 */
const flattenHeaderArray = (headers?: HeaderType[]) => {
  if (!headers) return undefined;
  return headers.reduce((obj, header) => {
    obj[header.key] = header.value;
    return obj;
  }, {} as StringStringMap);
};

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
          headers: flattenHeaderArray(item.headers as HeaderType[] | undefined),
          body: item.body,
        };
      // broadcast - extras as {KEY:VAL}, not {key:KEY, val:VAL}
      else if (item.action === "broadcast")
        return {
          action: item.action,
          label: item.label,
          intent: item.intent,
          extras: flattenHeaderArray(item.extras as HeaderType[] | undefined),
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
