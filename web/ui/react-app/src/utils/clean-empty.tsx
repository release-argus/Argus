/**
 * Recursively removes empty keys from an object and any key:value's that are the same as defaults
 *
 * @param obj - The object to clean
 * @param defaults - The default values to compare against
 * @returns A new object with empty keys removed and any key:value's removed that are the same as defaults
 */
const cleanEmpty = (
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  obj: any,
  defaults = [undefined, null, NaN, ""]
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
): any => {
  if (!defaults.length) return obj;
  if (defaults.includes(obj)) return;

  if (Array.isArray(obj))
    return obj
      .map((v) => (v && typeof v === "object" ? cleanEmpty(v, defaults) : v))
      .filter((v) => !defaults.includes(v));

  return Object.entries(obj).length
    ? Object.entries(obj)
        .map(([k, v]) => [
          k,
          v && typeof v === "object" ? cleanEmpty(v, defaults) : v,
        ])
        .reduce(
          (a, [k, v]) => (defaults.includes(v) ? a : { ...a, [k]: v }),
          {}
        )
    : obj;
};

export default cleanEmpty;
