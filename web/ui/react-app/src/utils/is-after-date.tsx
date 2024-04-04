/**
 * Returns whether the given date is in the past
 *
 * @param dateStr - The date string to compare to now
 * @returns Whether the date is after now
 */
const dateIsAfterNow = (dateStr: string) => {
  const then = new Date(dateStr);
  const now = new Date();
  return then > now;
};

export default dateIsAfterNow;
