import { ArgType } from "types/service-edit";
import { urlCommandsTrim } from "../components/modals/service-edit/util/url-command-trim";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type DiffObject = { [key: string]: any };

// deepDiff oldObj with newObj and return what's changed with newObj
const deepDiff = (oldObj: DiffObject, newObj: DiffObject): DiffObject => {
  const diff: DiffObject = {};

  // if oldObj is undefined, return newObj
  // e.g. DeployedVersion has no defaults
  if (oldObj === undefined) return newObj;

  // get all keys from both objects
  const keys = Object.keys(oldObj).concat(Object.keys(newObj));

  keys.forEach((key) => {
    // skip same values
    if ((newObj[key] || "") === (oldObj[key] || "")) return;

    // diff arrays
    if (Array.isArray(oldObj[key]) && Array.isArray(newObj[key])) {
      // if array lengths differ, include all elements in diff
      if (oldObj[key].length !== newObj[key].length) {
        diff[key] = newObj[key];
        // else, recurse on each element
      } else {
        const subDiff = oldObj[key].map(
          (elem: DiffObject, i: string | number) =>
            deepDiff(elem, newObj[key][i])
        );
        if (
          subDiff.some(
            (diffElem: DiffObject) => Object.keys(diffElem).length > 0
          )
        ) {
          diff[key] = newObj[key];
        }
      }
      // diff objects
    } else if (
      typeof oldObj[key] === "object" &&
      typeof newObj[key] === "object"
    ) {
      // recurse on nested objects
      const subDiff = deepDiff(oldObj[key], newObj[key]);
      if (Object.keys(subDiff).length > 0) diff[key] = subDiff;
      // diff scalars
    } else if (oldObj[key] !== newObj[key]) diff[key] = newObj[key];
  });

  return diff;
};

// stringifyQueryParam will return a query param string for a given key/value pair
// if value is undefined/null, it will return an empty string if omitUndefined is true
export const stringifyQueryParam = (
  key: string,
  value?: string | number | boolean,
  omitUndefined?: boolean
) =>
  omitUndefined && value == null
    ? ""
    : `${key}=${encodeURIComponent(value || "")}`;

export const convertToQueryParams = ({
  params,
  defaults,
  prefix = "",
}: {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  params: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  defaults?: any;
  prefix?: string;
}): string => {
  const queryParams: string[] = [];
  const changedParams = deepDiff(defaults, params);

  for (const key in changedParams) {
    const paramKey = prefix ? `${prefix}[${key}]` : key;
    if (key === "headers") {
      // Push all headers if any of them have changed
      queryParams.push(
        `${paramKey}=${encodeURIComponent(JSON.stringify(params[key]))}`
      );
      continue;
    } else if (typeof changedParams[key] === "object") {
      let modifiedObj;
      if (key === "url_commands")
        modifiedObj = toJSON(urlCommandsTrim(params[key]));
      else if (key === "require") {
        // Convert array of objects to array of strings
        modifiedObj =
          changedParams[key]?.command &&
          Object.keys(changedParams[key]?.command).length > 0
            ? toJSON({
                ...changedParams[key],
                command: Object.values<ArgType>(
                  changedParams.require.command
                ).map((value) => value.arg),
              })
            : toJSON(changedParams[key]);
      } else modifiedObj = toJSON(changedParams[key]);

      // Skip empty objects
      if (modifiedObj === "{}") continue;
      // Push all other objects
      queryParams.push(stringifyQueryParam(paramKey, modifiedObj));
      continue;
    }
    // Push all other scalars
    queryParams.push(stringifyQueryParam(paramKey, params[key]));
  }

  return queryParams.join("&");
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const removeEmpty = (obj: any) => {
  if (Array.isArray(obj)) return obj;

  const copy = { ...obj };
  Object.keys(copy).forEach((key) => {
    if (
      copy[key] == null ||
      copy[key] === "" ||
      (Array.isArray(copy[key]) && copy[key].length === 0) ||
      (typeof copy[key] === "object" &&
        Object.keys(removeEmpty(copy[key])).length === 0)
    ) {
      delete copy[key];
    }
  });
  return copy;
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const toJSON = (obj: any) => {
  return JSON.stringify(removeEmpty(obj), (key, value) => {
    return value;
  });
};
