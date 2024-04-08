import {
  FieldValues,
  UseFormClearErrors,
  UseFormGetValues,
  UseFormSetError,
} from "react-hook-form";

/**
 * Returns an error message if the value is not a number
 *
 * @param value - The value to test
 * @param use - Whether to use this test
 * @returns - An error message if the value is not a number
 */
export const numberTest = (value: string, use?: boolean) => {
  if (!value || !use) return;

  if (isNaN(Number(value))) return "Must be a number";
};

/**
 * Returns an error message if the value is not a valid RegEx
 *
 * @param value - The value to test
 * @param use - Whether to use this test
 * @returns - An error message if the value is not a valid RegEx
 */
export const regexTest = (value: string, use?: boolean) => {
  if (!value || !use) return;

  try {
    new RegExp(value);
  } catch (error) {
    return "Invalid RegEx";
  }
};

/**
 * Returns an error message if the value is required
 *
 * @param value - The value to test
 * @param name - The name of the field
 * @param setError - The function to set an error
 * @param clearErrors - The function to clear errors
 * @param use - Whether to use this test
 * @returns - An error message if the value is required and clears any errors if the value is non-empty
 */
export const requiredTest = (
  value: string,
  name: string,
  setError: UseFormSetError<FieldValues>,
  clearErrors: UseFormClearErrors<FieldValues>,
  use?: boolean | string
) => {
  if (!use) return;

  if (/.+/.test(value)) {
    clearErrors(name);
    return;
  }
  setError(name, {
    type: "required",
    message: use === true ? "Required" : use,
  });
  return use === true ? "Required" : use;
};

/**
 * Returns an error message if the value is not a unique child of the parent
 *
 * @param value - The value to test
 * @param name - The name of the field
 * @param getValues - The function to get the values of the form
 * @param use - Whether to use this test
 * @returns - An error message if the value is not a unique child of the parent
 */
export const uniqueTest = (
  value: string,
  name: string,
  getValues: UseFormGetValues<FieldValues>,
  use?: boolean
) => {
  if (!value || !use) return;

  const parts = name.split(".");
  const parent = parts.slice(0, parts.length - 2).join(".");
  const values = getValues(parent);
  const uniqueName = parts[parts.length - 1];
  const unique =
    values &&
    values
      .map((item: { [x: string]: string }) => item[uniqueName])
      // <=1 in case of default value
      .filter((item: string) => item === value).length <= 1;
  return unique || "Must be unique";
};

/**
 * Returns an error message if the value is not a valid URL
 *
 * @param value - The value to test
 * @param use - Whether to use this test
 * @returns An error message if the value is not a valid URL
 */
export const urlTest = (value: string, use?: boolean) => {
  if (!value || !use) return;

  try {
    const parsedURL = new URL(value);
    if (!["http:", "https:"].includes(parsedURL.protocol))
      throw new Error("Invalid protocol");
  } catch (error) {
    if (/^https?:\/\//.test(value)) {
      return "Invalid URL";
    }
    return "Invalid URL - http(s):// prefix required";
  }
};
