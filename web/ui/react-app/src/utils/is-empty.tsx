/**
 * Returns whether the array is empty, null, or undefined
 *
 * @param arg - The array to check
 * @returns Whether the array is empty, null, or undefined
 */
const isEmptyArray = <T extends unknown[] | unknown>(arg: T): boolean =>
  !arg || (arg as unknown[]).length === 0;

export default isEmptyArray;
