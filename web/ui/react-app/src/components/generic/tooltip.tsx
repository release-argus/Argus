import { FC, JSX, memo } from 'react';
import { OverlayTrigger, Tooltip } from 'react-bootstrap';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faQuestionCircle } from '@fortawesome/free-solid-svg-icons';

export type TooltipWithAriaProps =
	| { tooltip: string; tooltipAriaLabel?: string }
	| { tooltip: JSX.Element; tooltipAriaLabel: string }
	| { tooltip?: never; tooltipAriaLabel?: never };

type Props = {
	id?: string;
	placement?: 'top' | 'right' | 'bottom' | 'left';
};

type HelpTooltipProps = TooltipWithAriaProps & Props;

/**
 * A tooltip inside a question mark icon.
 *
 * @param id - The id of the tooltip.
 * @param tooltip - The text to display in the tooltip.
 * @param tooltipAriaLabel - The aria label for the tooltip (Defaults to the text).
 * @param placement - The placement of the tooltip.
 * @returns A hoverable tooltip inside a question mark icon.
 */
const HelpTooltip: FC<HelpTooltipProps> = ({
	id,
	tooltip,
	tooltipAriaLabel,
	placement = 'top',
}) => (
	<OverlayTrigger
		placement={placement}
		delay={{ show: 500, hide: 500 }}
		overlay={<Tooltip id="tooltip-help">{tooltip}</Tooltip>}
	>
		<FontAwesomeIcon
			id={id}
			aria-label={tooltipAriaLabel ?? tooltip}
			icon={faQuestionCircle}
			style={{
				paddingLeft: '0.25em',
				height: '0.75em',
			}}
		/>
	</OverlayTrigger>
);

export default memo(HelpTooltip);
