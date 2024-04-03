/**
 * Returns the first non-zero length argument
 *
 * @param args - The arguments to check for the first non-zero length
 * @returns The first non-zero length argument
 */
const firstNonEmpty = <T extends unknown[] | unknown>(
  ...args: T[]
): NonNullable<T> => {
  for (const arg of args) {
    if (arg && (arg as unknown[]).length) return arg;
  }
  return [] as NonNullable<T>;
};

export default firstNonEmpty;
