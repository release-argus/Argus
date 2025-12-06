/**
 * Whether the date is in the future.
 *
 * @param dateStr - The date string to compare to now.
 * @returns Whether the date is in the future.
 */
const dateIsAfterNow = (dateStr: string) => {
	const then = new Date(dateStr);
	const now = new Date();
	return then > now;
};

export default dateIsAfterNow;
