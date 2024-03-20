export function diffObjects<T>(a?: T, b?: T): boolean {
  // no defaults
  if (b === undefined && a !== undefined) return (a || "") == (b || "");
  // identical or a is completely unchanged
  if (a === b || a == null) return true;
  // if the default is null, return false
  if (b == null) return false;
  // if a is an array, check it's the same length as b
  if (
    Array.isArray(a) &&
    (!Array.isArray(b) ||
      (a.length !== b.length &&
        ("id" in b ? !("id" in a) && a.length + 1 === b.length : true)))
  ) {
    return false;
  }

  if (typeof b === "object") {
    const keys: Array<keyof T> = Object.keys(a) as Array<keyof T>;
    // check each key in object
    for (const key of keys) {
      if (!diffObjects(a[key], b[key])) return false;
    }
    return true;
  } else if (typeof b === "string")
    // a is undefined/empty or the same as b
    return (a || "") === "" || a === (b || "");
  // different
  else return a === b;
}
