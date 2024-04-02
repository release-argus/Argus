import { ArgType } from "types/service-edit";
import { urlCommandsTrim } from "components/modals/service-edit/util";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type DiffObject = { [key: string]: any };

/**
 * Returns the differences between the two objects
 *
 * @param newObj - The new object to compare
 * @param oldObj - The old object to compare
 * @returns The differences between the two objects
 */
export const deepDiff = (
  newObj: DiffObject,
  oldObj?: DiffObject
): DiffObject => {
  const diff: DiffObject = {};

  // if oldObj is undefined, return newObj
  // e.g. DeployedVersion has no defaults
  if (oldObj === undefined) return newObj;

  // get all keys from both objects
  const keys = Object.keys(oldObj).concat(Object.keys(newObj));

  keys.forEach((key) => {
    // skip same values
    if ((newObj[key] ?? "") === (oldObj[key] ?? "")) return;

    // diff arrays
    if (Array.isArray(oldObj[key]) || Array.isArray(newObj[key])) {
      // if array lengths differ, include all elements in diff
      if ((oldObj[key] || []).length !== newObj[key].length) {
        diff[key] = newObj[key];
        // else, recurse on each element
      } else {
        const subDiff = (oldObj[key] || []).map(
          (elem: DiffObject, i: string | number) =>
            deepDiff(newObj[key][i], elem)
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
      const subDiff = deepDiff(newObj[key], oldObj[key]);
      if (Object.keys(subDiff).length > 0) diff[key] = subDiff;
      // diff scalars
    } else if (oldObj[key] !== newObj[key]) diff[key] = newObj[key];
  });

  return diff;
};

/**
 * Returns an encoded query param string for a given key/value pair
 *
 * @param key - The key of the query param
 * @param value - The value of the query param
 * @param omitUndefined - Whether to omit undefined values
 * @returns An encoded query param string
 */
export const stringifyQueryParam = (
  key: string,
  value?: string | number | boolean,
  omitUndefined?: boolean
) =>
  omitUndefined && value == null
    ? ""
    : `${key}=${encodeURIComponent(value ?? "")}`;

/**
 * Returns the query params for the given object changes
 *
 * @param params - The new object
 * @param defaults - The old object
 * @param prefix - The prefix of the query params
 * @returns The query params of any changed values between the two objects
 */
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
  const changedParams = deepDiff(params, defaults);

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
        modifiedObj = JSON.stringify(urlCommandsTrim(params[key]));
      else if (key === "require") {
        // Convert array of objects to array of strings
        modifiedObj = JSON.stringify(
          changedParams[key]?.command &&
            Object.keys(changedParams[key]?.command).length > 0
            ? {
                ...changedParams[key],
                command: Object.values<ArgType>(
                  changedParams.require.command
                ).map((value) => value.arg),
              }
            : changedParams[key]
        );
      }
      // Include new undefined's in the JSONification
      else
        modifiedObj = JSON.stringify(changedParams[key], (_key, value) => {
          if (value === undefined) {
            return null;
          }
          return value;
        });

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
