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
    | StringFieldArray
    | NotifyNtfyAction[]
    | NotifyOpsGenieTarget[]
    | HeaderType[];
}
interface StringStringMap {
  [key: string]: string;
}

/**
 * convertValuesToString will convert the values of an object to a properly formatted strings
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
          !(value as NotifyOpsGenieTarget[]).find(
            (item) => (item.value ?? "") !== ""
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
          !(value as StringFieldArray).find(
            (item) => (item.label || item.arg || "") !== ""
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
          !(value as NotifyOpsGenieTarget[]).find(
            (item) => (item.value ?? "") !== ""
          )
        ) {
          return result;
        }
        result[key] = JSON.stringify(flattenHeaderArray(value as HeaderType[]));
      }
    } else {
      result[key] = String(value);
    }
    return result;
  }, {} as StringStringMap);

/**
 * FlattenStringFieldArray will extract the values into a JSON string
 *
 * @param obj - The StringFieldArray to flatten { arg: "value1" }[]
 * @returns A JSON string of the values ["value1", "value2", ...]
 */
const FlattenStringFieldArray = (obj: StringFieldArray): string =>
  JSON.stringify(obj.map((item) => Object.values(item)[0]));

/**
 * flattenHeaderArray will convert {key:KEY, val:VAL}[] to {KEY:VAL, ...}
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
 * convertNtfyActionsToString will convert the NotifyNtfyAction[] to a JSON string
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
 * convertOpsGenieTargetToString will convert the NotifyOpsGenieTarget[] to a JSON string
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
