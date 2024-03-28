/**
 * firstNonDefault returns the first non-empty string from a list of arguments
 * @param args -  The list of arguments to check
 * @returns The first non-empty string
 */
export const firstNonDefault: (...args: unknown[]) => string = (
  ...args: unknown[]
) => {
  // Iterate through all arguments and return the first non-empty one
  for (const arg of args) {
    if ((arg ?? "") !== "") return `${arg}`;
  }
  // If no non-empty argument is found, return an empty string by default
  return "";
};
