import { formatDistanceToNow } from 'date-fns';
import { enGB } from 'date-fns/locale';

/**
 * Returns how long ago date was, e.g. '20 days ago'.
 *
 * @param date - The date to format.
 */
const relativeDate = (date: Date) =>
	formatDistanceToNow(date, { addSuffix: true, locale: enGB });
export default relativeDate;
