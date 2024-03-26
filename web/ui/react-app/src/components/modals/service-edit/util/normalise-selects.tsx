/**
 * normaliseForSelect will take an any-case value and check whether it's in the provided options array and
 * return the value if it is
 *
 * @param options - The options to check against
 * @param value - The value to normalise
 * @returns The normalised value
 */
export const normaliseForSelect = (
  options: { value: string; label: string }[],
  value?: string
): { value: string; label: string } | undefined => {
  if (value === undefined) return undefined;

  const option = options.find(
    (option) => option.value.toLowerCase() === value.toLowerCase()
  );

  return option;
};
