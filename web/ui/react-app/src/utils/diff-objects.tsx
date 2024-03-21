// diffObjects will return true if the fieldValues (a) are undefined or the allowedDefined keys are the same as the defaults (b)
export function diffObjects<T>(
  a?: T,
  b?: T,
  allowedDefined?: string[],
  key?: string
): boolean {
  // no defaults
  if (b === undefined && a !== undefined) return (a || "") == (b || "");
  // a is completely unchanged
  if (a == null) return true;
  // if the default is null, return false
  if (b == null) return false;
  // if a is an array, check it's the same length as b
  if (
    Array.isArray(a) &&
    (!Array.isArray(b) ||
      (a.length !== b.length &&
        ("id" in b ? !("id" in a) && a.length + 1 === b.length : true)))
  ) {
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
    if ((a || "") === "") return true;
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
// containsEndsWith will return true if the list contains a string that this substring ends with
// if list is undefined, it will return undefinedListDefault (default=false)
const containsEndsWith = (
  substring: string,
  list?: string[],
  undefinedListDefault?: boolean
): boolean => {
  return list
    ? list.some((item) => substring.endsWith(item))
    : undefinedListDefault ?? false;
};
