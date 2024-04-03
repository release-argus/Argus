/**
 * Recursively trims the object, removing empty objects/values.
 *
 * @param obj - The object to remove empty values from
 * @returns The object with all empty values removed
 */

import isEmptyOrNull from "./is-empty-or-null";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const removeEmptyValues = (obj: { [x: string]: any }) => {
  for (const key in obj) {
    // [] Empty array
    if (Array.isArray(obj[key])) {
      if (obj[key].length === 0) delete obj[key];
      // {} Object
    } else if (
      typeof obj[key] === "object" &&
      !["notify", "webhook"].includes(key) // not notify/webhook as they may be empty to reference globals
    ) {
      // Check object
      removeEmptyValues(obj[key]);
      // Empty object
      if (Object.keys(obj[key] ?? []).length === 0) {
        delete obj[key];
        continue;
      }
      // "" Empty/undefined string
    } else if (isEmptyOrNull(obj[key])) delete obj[key];
  }
  return obj;
};

export default removeEmptyValues;
