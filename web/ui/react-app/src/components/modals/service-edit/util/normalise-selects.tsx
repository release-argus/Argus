/**
 * Returns the option from the provided options array that matches the value
 *
 * @param options - The options to search
 * @param value - The value to search for
 * @returns The option that matches the value, case-insensitive
 */
export const normaliseForSelect = (
  options: { value: string; label: string }[],
  value?: string
): { value: string; label: string } | undefined => {
  if (value === undefined) return undefined;

  const wantedValue = value.toLowerCase();
  return options.find((option) => option.value.toLowerCase() === wantedValue);
};
