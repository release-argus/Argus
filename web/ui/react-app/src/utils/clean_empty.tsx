export const cleanEmpty = function (
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  obj: any,
  defaults = [undefined, null, NaN, ""]
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
): any {
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
