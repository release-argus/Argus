export const useGlobalOrDefault = (
  globalValue?: string | number,
  defaultValue?: string | number,
  hardDefaultValue?: string | number
): string => {
  return `${globalValue ?? defaultValue ?? hardDefaultValue ?? ""}`;
};
