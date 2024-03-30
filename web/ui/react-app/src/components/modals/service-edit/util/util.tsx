/**
 * Returns the first non-empty value global > default > hard default
 * @param globalValue - The global value
 * @param defaultValue - The default value
 * @param hardDefaultValue - The hard default value
 * @returns The first non-empty value as a string
 */
export const globalOrDefault = (
  globalValue?: string | number,
  defaultValue: string | number = "",
  hardDefaultValue: string | number = ""
) => `${globalValue ?? defaultValue ?? hardDefaultValue}`;
