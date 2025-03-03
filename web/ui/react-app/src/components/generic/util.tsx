import { Position } from 'types/config';
import { ScreenBreakpoint } from 'types/util';

const getPadding = (
	position?: Position,
	col?: number,
	breakpoint?: ScreenBreakpoint,
	previous?: { pos?: Position; col?: number; applied?: boolean },
) => {
	// Skip padding if the column is full width,
	// or the position is the same as the previous position (and that padding has already been applied).
	if (col === 12 || (previous?.applied && position === previous?.pos))
		return null;

	const breakpointPrefix = breakpoint ? '-' + breakpoint : '';

	// Apply the padding based on the position.
	switch (position) {
		case 'right':
			return `ps${breakpointPrefix}-1 pe${breakpointPrefix}-0`;
		case 'middle':
			return `ps${breakpointPrefix}-1 pe${breakpointPrefix}-1`;
		case 'left':
			return `ps${breakpointPrefix}-0 pe${breakpointPrefix}-1`;
	}

	return null; // No positioning given.
};

type formPaddingProps = {
	col_xs?: number;
	col_sm?: number;
	col_md?: number;
	col_lg?: number;

	positionXS?: Position;
	positionSM?: Position;
	positionMD?: Position;
	positionLG?: Position;
};

/**
 * The padding classes for a form item.
 *
 * @param col_xs - The number of columns the item takes up on XS+ screens.
 * @param col_sm - The number of columns the item takes up on SM+ screens.
 * @param col_md - The number of columns the item takes up on MD+ screens.
 * @param col_lg - The number of columns the item takes up on LG+ screens.
 *
 * @param positionXS - The position of the item on XS+ screens.
 * @param positionSM - The position of the item on SM+ screens.
 * @param positionMD - The position of the item on MD+ screens.
 * @param positionLG - The position of the item on LG+ screens.
 * @returns The padding classes for a form item depending on the screen size.
 */
export const formPadding = ({
	col_xs,
	col_sm,
	col_md,
	col_lg,

	positionXS,
	positionSM,
	positionMD,
	positionLG,
}: formPaddingProps) => {
	const paddingClasses: string[] = [];

	// All widths are max, so no padding needed.
	if (col_lg === 12 && col_md === 12 && col_sm === 12 && col_xs === 12)
		return '';

	// Same padding if all widths are not max.
	const allNotMax =
		col_lg !== 12 && col_md !== 12 && col_sm !== 12 && col_xs !== 12;
	// Same positioning on all breakpoints.
	if (
		(positionLG || positionMD || positionSM || positionXS) && // Have a position
		positionXS === (positionSM ?? positionXS) && // XS === SM
		(positionSM ?? positionXS) === (positionMD ?? positionSM ?? positionXS) && // SM === MD
		(positionMD ?? positionSM ?? positionXS) ===
			(positionLG ?? positionMD ?? positionSM ?? positionXS) && // MD === LG
		allNotMax
	)
		return getPadding(positionLG ?? positionMD ?? positionSM ?? positionXS);

	const addPadding = (
		position?: Position,
		col?: number,
		breakpoint?: ScreenBreakpoint,
		previous?: { pos?: Position; col?: number; applied?: boolean },
	) => {
		// Skip full-width columns.
		if (col === 12)
			return {
				pos: position ?? previous?.pos,
				col: col ?? previous?.col,
				applied: previous?.applied,
			};

		const newPosition = position ?? previous?.pos;
		const padding = getPadding(newPosition, col, breakpoint, previous);

		// Only push the padding when not null.
		if (padding) paddingClasses.push(padding);

		// Return the updated previous position, column and applied status.
		return {
			pos: newPosition,
			col: col ?? previous?.col,
			applied:
				!!padding || (newPosition === previous?.pos && previous?.applied),
		};
	};

	// Padding for each screen size (if it's different to the previous breakpoint).
	// XS+
	let previous = addPadding(positionXS, col_xs);
	// SM+
	previous = addPadding(positionSM, col_sm, 'sm', previous);
	// MD+
	previous = addPadding(positionMD, col_md, 'md', previous);
	// LG+
	addPadding(positionLG, col_lg, 'lg', previous);

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
