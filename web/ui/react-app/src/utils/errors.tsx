import { FieldError, FieldErrors, FieldValues } from "react-hook-form";

import { StringStringMap } from "types/config";
import { isEmptyObject } from "./is-empty";

/**
 * getNestedError gets the error for a potentially nested key in a react-hook-form errors object
 *
 * @param errors - The errors object from react-hook-form
 * @param key - The key to get the error for
 * @returns The error for the provided key
 */
export const getNestedError = (
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  errors: any,
  key: string
): FieldError | undefined =>
  key.split(".").reduce((acc, key) => acc?.[key], errors);

/**
 * Extracts and flattens errors from a react-hook-form errors object
 *
 * e.g. { first: { second: [ {item1: {message: "reason"}}, {item2: {message: "otherReason"}} ] } }
 * becomes { first.second.0.item1: "reason", first.second.1.item2: "otherReason"}
 *
 * @param errors - The errors object from react-hook-form
 * @param name - The path to filter the errors by
 * @returns The flattened errors object for the provided path
 */
export const extractErrors = (
  errors: FieldErrors<FieldValues>,
  path = ""
): StringStringMap | undefined => {
  const flatErrors: StringStringMap = {};
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const traverse = (prefix: string, obj: any) => {
    for (const key in obj) {
      const value = obj[key];
      if (value !== null) {
        const fullPath = prefix ? `${prefix}.${key}` : key;
        const atPath = fullPath.startsWith(path); // if the path is in the key
        if (atPath || path.includes(fullPath)) {
          if (typeof value === "object" && !value.hasOwnProperty("ref"))
            traverse(fullPath, value);
          else if (atPath && value?.hasOwnProperty("ref")) {
            const trimmedPath = path
              ? fullPath.substring(path.length + 1)
              : fullPath;
            flatErrors[trimmedPath] = value.message;
          }
        }
      }
    }
  };
  traverse("", errors);
  return isEmptyObject(flatErrors) ? undefined : flatErrors;
};
