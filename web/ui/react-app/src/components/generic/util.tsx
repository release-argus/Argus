import { Position } from 'types/config';

type formPaddingProps = {
	col_xs: number;
	col_sm: number;
	position?: Position;
	positionXS?: Position;
};

/**
 * The padding classes for a form item.
 *
 * @param col_xs - The number of columns the item takes up on XS+ screens.
 * @param col_sm - The number of columns the item takes up on SM+ screens.
 * @param position - The position of the item on SM+.
 * @param positionXS - The position of the item on XS.
 * @returns The padding classes for a form item depending on the screen size.
 */
export const formPadding = ({
	col_xs,
	col_sm,
	position,
	positionXS = position,
}: formPaddingProps) => {
	const paddingClasses = [];

	// Padding for SM+.
	if (col_sm !== 12) {
		// Padding for being on the right.
		if (position === 'right') {
			paddingClasses.push('ps-sm-2');
		}
		// Padding for being in the middle.
		else if (position === 'middle') {
			paddingClasses.push('ps-sm-1');
			paddingClasses.push('pe-sm-1');
		}
		// Padding for being on the left.
		else {
			paddingClasses.push('pe-sm-2');
		}
	}

	// If the padding is the same on XS+ and SM+, convert the SM+ padding to XS+.
	if (position === positionXS && col_sm !== 12 && col_xs !== 12) {
		for (let i = 0; i < paddingClasses.length; i++) {
			paddingClasses[i] = paddingClasses[i].replace('-sm', '');
		}
	}

	// Padding for XS.
	else if (col_xs !== 12) {
		// XS Padding for being on the right.
		if (positionXS === 'right') {
			paddingClasses.push('ps-1');

			// Remove padding for SM+
			// if it's full width,
			// or we're on the left for SM+.
			if (col_sm === 12 || position === 'left') {
				paddingClasses.push('ps-sm-0');
			}
		}

		// XS Padding for being in the middle.
		else if (positionXS === 'middle') {
			paddingClasses.push('ps-1');
			paddingClasses.push('pe-1');

			// Remove padding for SM+ if it's full width.
			if (col_sm === 12) {
				paddingClasses.push('ps-sm-0');
				paddingClasses.push('pe-sm-0');
			}
			// If we're on the right, remove the pe on SM+.
			else if (position === 'right') {
				paddingClasses.push('pe-sm-0');
			}
			// If we're on the left, remove the ps on SM+.
			else if (position === 'left') {
				paddingClasses.push('ps-sm-0');
			}
		}

		// XS Padding for being on the left.
		else if (positionXS === 'left') {
			paddingClasses.push('pe-1');

			// Remove padding for SM+
			// if it's full width,
			// or we're on the right for SM+.
			if (col_sm === 12 || position === 'right') {
				paddingClasses.push('pe-sm-0');
			}
		}
	}

	return paddingClasses.join(' ');
};

/**
 * Pluralise a string.
 *
 * @param str - The string to pluralise.
 * @param count - The count of the string.
 * @returns The pluralised string (if count is not 1).
 */
export const pluralise = (str: string, count?: number): string => {
	if (count !== 1) return str + 's';
	return str;
};
