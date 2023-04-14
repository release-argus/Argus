// removeEmptyValues will remove all empty strings/lists from an object recursively
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const removeEmptyValues = (obj: { [x: string]: any }) => {
  for (const key in obj) {
    // [] Empty array
    if (Array.isArray(obj[key])) {
      if (obj[key].length === 0) {
        delete obj[key];
      }
      // {} Object
    } else if (
      typeof obj[key] === "object" &&
      !["notify", "webhook"].includes(key)
    ) {
      // Check object
      removeEmptyValues(obj[key]);
      // Empty object
      if (Object.keys(obj[key]).length === 0) {
        delete obj[key];
        continue;
      }
      // "" Empty/undefined string
    } else if (obj[key] === "" || obj[key] === undefined) {
      delete obj[key];
    }
  }
  return obj;
};

export default removeEmptyValues;
