export const strToBool = (str?: string | boolean): boolean | null => {
  if (typeof str === "boolean") return str;
  if (str == null || str === "") return null;
  return ["true", "yes"].includes(str.toLowerCase());
};

export const boolToStr = (bool?: boolean) =>
  bool === undefined ? "" : bool.toString();
