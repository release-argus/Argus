import { formatRelative } from 'date-fns';
import { enGB } from 'date-fns/locale';

/**
 * Returns a relative date string.
 *
 * @param date - The date to format.
 */
const relativeDate = (date: Date) => {
	const now = new Date();
	return formatRelative(date, now, { locale: enGB });
};
export default relativeDate;
