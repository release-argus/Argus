export const valueOrDefault = function <T>(
  value: T,
  defaultValue: string
): T | string {
  // Return `value` if it's not null, otherwise `defaultValue`
  return value ? defaultValue : value;
};

// Render `val` if it's not null, otherwise `defaultValue`
export const RenderValueOrDefault = <T,>(
  val: T,
  defaultValue: string
): React.ReactNode => <>{valueOrDefault(val, defaultValue)}</>;
