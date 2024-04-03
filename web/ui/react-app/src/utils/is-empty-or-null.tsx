/**
 * Returns whether the value is empty, null, or undefined
 *
 * @param value - The value to check
 * @returns Whether the value is empty, null, or undefined
 */
const isEmptyOrNull = (value: unknown | undefined): boolean => {
  return (value ?? "") === "";
};

export default isEmptyOrNull;
