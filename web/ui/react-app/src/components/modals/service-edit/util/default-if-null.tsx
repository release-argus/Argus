/**
 * Returns the default value if the provided variable is `null`.
 * Otherwise, returns the variable itself.
 *
 * @template T - The type of the variable.
 * @param value - The variable to check.
 * @param defaultValue - The default value to return if `value` is `null`.
 * @returns The `defaultValue` if `value` is `null`, `undefined` if `value` is `undefined`,
 *          or `value` otherwise.
 */
export const defaultIfNull = <T,>(
  value: T | null | undefined,
  defaultValue: T,
): T | undefined => {
  return value === null ? defaultValue : value;
};
