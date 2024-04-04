import isEmptyOrNull from "./is-empty-or-null";

/**
 * Returns the first non-empty from the list of arguments
 *
 * - All undefined/empty = ""
 *
 * @param args -  The list of arguments to check
 * @returns The first non-empty string
 */
const firstNonDefault: (...args: unknown[]) => string = (
  ...args: unknown[]
) => {
  // Iterate through all arguments and return the first non-empty one
  for (const arg of args) {
    if (!isEmptyOrNull(arg)) return `${arg}`;
  }
  // If no non-empty argument is found, return an empty string by default
  return "";
};

export default firstNonDefault;
