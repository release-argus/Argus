// flattenErrors produces a 2d key:val map of errors
//
// e.g. { first: { second: [ {item1: {message: "reason"}}, {item2: {message: "otherReason"}} ] } }
// becomes { first.second.1.item1: "reason", first.second.2.item2: "otherReason"}
// note that numerical indices start at 1

import { FieldErrors, FieldValues } from "react-hook-form";

export const flattenErrors = (errors: FieldErrors<FieldValues>) => {
  const flatErrors: { [key: string]: string } = {};
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const traverse = (prefix: string, obj: any) => {
    for (const key in obj) {
      const value = obj[key];
      if (value !== null) {
        const fullPath = `${prefix}${prefix ? `.${key}` : key}`;
        if (typeof value === "object" && !value.hasOwnProperty("type"))
          traverse(fullPath, value);
        else if (value?.hasOwnProperty("type"))
          flatErrors[fullPath] = value.message;
      }
    }
  };
  traverse("", errors);
  return flatErrors;
};

// extractErrors will extract errors matching the provided path
export const extractErrors = (
  errors: FieldErrors<FieldValues>,
  path: string
): { [key: string]: string } => {
  const flatErrors: { [key: string]: string } = {};
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const traverse = (prefix: string, obj: any) => {
    for (const key in obj) {
      const value = obj[key];
      if (value !== null) {
        const fullPath = `${prefix}${prefix ? `.${key}` : key}`;
        const atPath = fullPath.startsWith(path); // if the path is in the key
        if (atPath || path.includes(fullPath)) {
          if (typeof value === "object" && !value.hasOwnProperty("type"))
            traverse(fullPath, value);
          else if (atPath && value?.hasOwnProperty("type"))
            flatErrors[fullPath.substring(path.length + 1)] = value.message;
        }
      }
    }
  };
  traverse("", errors);
  return flatErrors;
};
