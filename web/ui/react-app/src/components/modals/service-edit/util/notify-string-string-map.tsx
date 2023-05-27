import {
  HeaderType,
  NotifyNtfyAction,
  NotifyOpsGenieTarget,
} from "types/config";

import { NotifyOpsGenieDetailType } from "types/service-edit";

interface StringAnyMap {
  [key: string]:
    | string
    | number
    | boolean
    | undefined
    | NotifyNtfyAction[]
    | NotifyOpsGenieTarget[]
    | NotifyOpsGenieDetailType[];
}
interface StringStringMap {
  [key: string]: string;
}

export const convertValuesToString = (obj: StringAnyMap): StringStringMap =>
  Object.entries(obj).reduce((result, [key, value]) => {
    if (typeof value === "object") {
      // opsGenie.responders || opsGenie.visibleto
      if (["responders", "visibleto"].includes(key)) {
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
          (value as NotifyNtfyAction[]).find(
            (item) => (item.label || "") === ""
          )
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
            (item) => (item.value || "") === ""
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

// convertNtfyActionsToString will convert the NotifyNtfyAction[] to a JSON string
const convertNtfyActionsToString = (obj: NotifyNtfyAction[]): string =>
  JSON.stringify(
    (obj as NotifyNtfyAction[]).map((item) => {
      if (item.action === "view")
        return {
          action: item.action,
          label: item.label,
          url: item.url,
        };
      else if (item.action === "http")
        return {
          action: item.action,
          label: item.label,
          url: item.url,
          method: item.method,
          headers: item.headers,
          body: item.body,
        };
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

// convertOpsGenieTargetToString will convert the NotifyOpsGenieTarget[] to a JSON string
const convertOpsGenieTargetToString = (obj: NotifyOpsGenieTarget[]): string =>
  JSON.stringify(
    (obj as NotifyOpsGenieTarget[]).map(({ type, sub_type, value }) => ({
      type: type,
      [sub_type]: value,
    }))
  );
