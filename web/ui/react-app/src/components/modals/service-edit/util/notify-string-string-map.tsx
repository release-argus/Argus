import { NotifyOpsGenieDetailType } from "types/service-edit";
import { NotifyOpsGenieTarget } from "types/config";

interface StringAnyMap {
  [key: string]:
    | string
    | number
    | boolean
    | undefined
    | NotifyOpsGenieTarget[]
    | NotifyOpsGenieDetailType[]
    | { [key: string]: string }[];
}
interface StringStringMap {
  [key: string]: string;
}

export const convertValuesToString = (obj: StringAnyMap): StringStringMap =>
  Object.entries(obj).reduce((result, [key, value]) => {
    if (typeof value === "object") {
      if (["responders", "visibleto"].includes(key)) {
        result[key] = JSON.stringify(
          (value as NotifyOpsGenieTarget[]).map(
            ({ type, sub_type, value }) => ({ type: type, [sub_type]: value })
          )
        );
        // details
      } else {
        result[key] = JSON.stringify(
          (value as { [key: string]: string }[]).reduce(
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
