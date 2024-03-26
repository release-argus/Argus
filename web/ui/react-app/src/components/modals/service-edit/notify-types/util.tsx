export const globalOrDefault: (...args: unknown[]) => string = (
  ...args: unknown[]
) => {
  // Iterate through all arguments and return the first non-empty one
  for (const arg of args) {
    if (arg !== undefined && arg !== null && arg !== "") {
      return `${arg}`;
    }
  }
  // If no non-empty argument is found, return an empty string by default
  return "";
};
