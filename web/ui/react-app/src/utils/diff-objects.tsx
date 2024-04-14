import { isEmptyArray } from "./is-empty";
import isEmptyOrNull from "./is-empty-or-null";

/**
 * Returns whether a is the same as b after allowing for only values at
 * allowDefined to be the value in b
 *
 * @param a - The fieldValues to compare
 * @param b - The defaults to compare against
 * @param allowDefined - The keys that are allowed to be defined and match the defaults
 * @param key - The key path of the current fieldValues
 * @returns Whether the fieldValues are different from the defaults
 */
export function diffObjects<T>(
  a?: T,
  b?: T,
  allowDefined?: string[],
  key?: string
): boolean {
  // no defaults
  if (b === undefined && a !== undefined) return (a ?? "") !== (b ?? "");
  // if a is empty/undefined, treat as unchanged
  if (isEmptyOrNull(a)) return false;
  // if there are no defaults, treat as changed
  if (isEmptyOrNull(b)) return true;
  // if a is an array, check it's the same length as b
  if (
    Array.isArray(a) &&
    (!Array.isArray(b) ||
      a.length !== b.length ||
      // if only one has an id, ensure it's not a length difference of 1
      (a.hasOwnProperty("id") != b.hasOwnProperty("id") &&
        Math.abs(a.length - b.length) !== 1))
  )
    // Non-empty means it's different as the lengths are different
    return !isEmptyArray(a);

  if (typeof b === "object") {
    const keys: Array<keyof T> = Object.keys(a as object) as Array<keyof T>;
    // check each key in object
    for (const k of keys) {
      if (
        diffObjects(
          a?.[k],
          b?.[k],
          allowDefined,
          key ? `${key}.${String(k)}` : String(k)
        )
      )
        // difference!
        return true;
    }
    // No differences found
    return false;
  } else if (typeof b === "string") {
    // a is defined, and on a key that's allowed. is a the same as b?
    if (containsEndsWith(key || "-", allowDefined)) return a !== b;
    // else, we've got a difference
    return true;
  }
  // different - a is defined at a key that is not allowed
  else return containsEndsWith(key || "-", allowDefined) ? a !== b : true;
}

/**
 * Returns whether `list` contains a string that `substring` starts with.
 *
 * @param substring - The string to check if it starts with any of the items in the list
 * @param list - The list of strings to check against
 * @param undefinedListDefault - The default value to return if the list is undefined
 * @default undefinedListDefault=false
 * @returns Whether the substring starts with any of the items in the list
 */
export const containsStartsWith = (
  substring: string,
  list?: string[],
  undefinedListDefault = false
): boolean => {
  return list
    ? list.some((item) => substring.startsWith(item))
    : undefinedListDefault;
};

/**
 * Returns whether `list` contains a string that `substring` ends with.
 *
 *
 * @param substring - The string to check if it ends with any of the items in the list
 * @param list - The list of strings to check against
 * @param undefinedListDefault - The default value to return if the list is undefined
 * @default undefinedListDefault=false
 * @returns Whether the substring ends with any of the items in the list
 */
export const containsEndsWith = (
  substring: string,
  list?: string[],
  undefinedListDefault = false
): boolean => {
  return list
    ? list.some((item) => substring.endsWith(item))
    : undefinedListDefault;
};
