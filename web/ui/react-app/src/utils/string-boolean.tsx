// StrToBool will convert str to a boolean
export const strToBool = (str?: string): boolean | undefined => {
  return str === "undefined" || str === undefined ? undefined : str === "true";
};

// boolToStr will xonvert bool to a string
export const boolToStr = (bool?: boolean): string => {
  return `${bool === undefined ? "undefined" : bool}`;
};
