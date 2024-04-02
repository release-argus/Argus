import { FieldError, useFormState } from "react-hook-form";

import { getNestedError } from "utils";

/**
 * Returns the error of a field in a form
 *
 * @param name - The name of the field in the form
 * @param wanted - Whether the error is wanted
 * @returns The error of the field
 */
export const useError = (
  name: string,
  wanted?: boolean
): FieldError | undefined => {
  const { errors } = useFormState({ name: name, exact: true });
  if (!wanted) return undefined;
  return getNestedError(errors, name);
};
