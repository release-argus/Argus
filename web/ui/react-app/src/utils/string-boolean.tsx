import isEmptyOrNull from "./is-empty-or-null";

/**
 * Returns the boolean value of a string
 *
 * @param str - The string to convert to a boolean
 * @returns The boolean value of the string, or null if the string is empty
 */
export const strToBool = (str?: string | boolean | null): boolean | null => {
  if (typeof str === "boolean") return str;
  if (isEmptyOrNull(str)) return null;
  return ["true", "yes"].includes((str as string).toLowerCase());
};

export const boolToStr = (bool?: boolean) =>
  bool === undefined ? "" : bool.toString();
