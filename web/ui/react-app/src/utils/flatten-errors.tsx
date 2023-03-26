// flattenErrors produces a 2d key:val map of errors
//
// e.g. { first: { second: [ {item1: {message: "reason"}}, {item2: {message: "otherReason"}} ] } }
// becomes { first.second.1.item1: "reason", first.second.2.item2: "otherReason"}
// note that numerical indices start at 1
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const flattenErrors = (errors: any) => {
  const flatErrors: { [key: string]: string } = {};
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const traverse = (prefix: string, obj: any) => {
    for (const key in obj) {
      if (obj[key] !== null) {
        if (typeof obj[key] === "object" && !obj[key].hasOwnProperty("type")) {
          // Add 1 to indices
          const humanisedKey = isNaN(parseFloat(key))
            ? key
            : parseFloat(key) + 1;
          traverse(`${prefix}${prefix ? `.${humanisedKey}` : key}`, obj[key]);
        } else if (obj[key]?.hasOwnProperty("type")) {
          flatErrors[`${prefix}${prefix ? `.${key}` : key}`] = obj[key].message;
        }
      }
    }
  };
  traverse("", errors);
  return flatErrors;
};

export default flattenErrors;
