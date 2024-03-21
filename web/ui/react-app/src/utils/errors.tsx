import { FieldErrors, FieldValues } from "react-hook-form";

// extractErrors will extract and flatten errors with the option to filter only
// errors starting with the provided path
//
// e.g. { first: { second: [ {item1: {message: "reason"}}, {item2: {message: "otherReason"}} ] } }
// becomes { first.second.0.item1: "reason", first.second.1.item2: "otherReason"}
export const extractErrors = (
  errors: FieldErrors<FieldValues>,
  path = ""
): { [key: string]: string } => {
  const flatErrors: { [key: string]: string } = {};
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const traverse = (prefix: string, obj: any) => {
    for (const key in obj) {
      const value = obj[key];
      if (value !== null) {
        const fullPath = `${prefix}${prefix ? `.${key}` : key}`;
        const atPath = fullPath.startsWith(path); // if the path is in the key
        if (atPath || path.includes(fullPath)) {
          if (typeof value === "object" && !value.hasOwnProperty("type"))
            traverse(fullPath, value);
          else if (atPath && value?.hasOwnProperty("type")) {
            const trimmedPath = path
              ? fullPath.substring(path.length + 1)
              : fullPath;
            flatErrors[trimmedPath] = value.message;
          }
        }
      }
    }
  };
  traverse("", errors);
  return flatErrors;
};
