export function diffObjects<T>(a?: T, b?: T): boolean {
  // no defaults
  if (b === undefined && a !== undefined) {
    return false;
  }
  // identical or a is completely unchanged
  if (a === b || a == null) {
    return true;
  }
  // if the default is null, return false
  if (b == null) {
    return false;
  }
  // if a is an array, check it's the same length as b
  if (Array.isArray(a) && (!Array.isArray(b) || a.length !== b.length)) {
    return false;
  }

  if (typeof b === "object") {
    const keys: Array<keyof T> = Object.keys(a) as Array<keyof T>;
    // check each key in object
    for (const key of keys) {
      if (!diffObjects(a[key], b[key])) {
        return false;
      }
    }
    return true;
  } else if (typeof b === "string") {
    // a is undefined/empty, or a is the same as b
    return (a || "") === "" || a === (b || "");
  } else {
    // different
    return a === b;
  }
}
