import isEmptyOrNull from "./is-empty-or-null";

/**
 * Returns whether a is different from b after allowing for only values at
 * allowedDefined to be the value in b
 *
 * @param a - The fieldValues to compare
 * @param b - The defaults to compare against
 * @param allowedDefined - The keys that are allowed to be defined and match the defaults
 * @param key - The key path of the current fieldValues
 * @returns Whether the fieldValues are different from the defaults
 */
export function diffObjects<T>(
  a?: T,
  b?: T,
  allowedDefined?: string[],
  key?: string
): boolean {
  // no defaults
  if (b === undefined && a !== undefined) return (a ?? "") === (b ?? "");
  // a is completely unchanged
  if (a == null) return true;
  // if the default is null, return false
  if (b == null) return false;
  // if a is an array, check it's the same length as b
  if (Array.isArray(a) && (!Array.isArray(b) || a.length !== b.length)) {
    // Treat empty fieldValues as the same
    return a.length === 0;
  }

  if (typeof b === "object") {
    const keys: Array<keyof T> = Object.keys(a) as Array<keyof T>;
    // check each key in object
    for (const k of keys) {
      if (
        !diffObjects(
          a[k],
          b[k],
          allowedDefined,
          key ? `${key}.${String(k)}` : String(k)
        )
      )
        return false;
    }
    return true;
  } else if (typeof b === "string") {
    // a is undefined/empty
    if (isEmptyOrNull(a)) return true;
    // a is defined, and on a key that's allowed and is the same as b
    if (
      containsEndsWith(
        key || "-",
        allowedDefined,
        allowedDefined ? true : false
      )
    )
      return a === b;
    // else, we've got a difference
    return false;
  }
  // different - a is undefined, or ix a key that is allowed
  else
    return (
      a === undefined ||
      containsEndsWith(
        key || "-",
        allowedDefined,
        allowedDefined ? true : false
      )
    );
}
/**
 * Returns whether `list` contains a string that `substring` ends with.
 *
 *
 * @param substring - The string to check if it ends with any of the items in the list
 * @param list - The list of strings to check against
 * @param undefinedListDefault - The default value to return if the list is undefined (default=false)
 * @default undefinedListDefault=false
 * @returns Whether the substring ends with any of the items in the list
 */
const containsEndsWith = (
  substring: string,
  list?: string[],
  undefinedListDefault?: boolean
): boolean => {
  return list
    ? list.some((item) => substring.endsWith(item))
    : undefinedListDefault ?? false;
};
