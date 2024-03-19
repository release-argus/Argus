export const globalOrDefault = (
  globalValue?: string | number,
  defaultValue: string | number = "",
  hardDefaultValue: string | number = ""
) => `${globalValue || defaultValue || hardDefaultValue}`;
