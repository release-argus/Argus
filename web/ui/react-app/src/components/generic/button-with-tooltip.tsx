import { Button, OverlayTrigger, Tooltip } from 'react-bootstrap';
import { FC, ReactElement } from 'react';

type ButtonWithTooltipProps = {
	hoverTooltip?: boolean;
	tooltip: string;
	onClick: () => void;
	icon: ReactElement;
};

/**
 * A button that displays a tooltip on hover.
 *
 * @param hoverTooltip - Whether the tooltip should be displayed on hover.
 * @param tooltip - The text to display in the tooltip.
 * @param onClick - The function to call when the button is clicked.
 * @param icon - The icon to display on the button.
 */
const ButtonWithTooltip: FC<ButtonWithTooltipProps> = ({
	hoverTooltip,
	tooltip,
	onClick,
	icon,
}) => {
	const button = (
		<Button
			variant="secondary"
			className="border-0"
			onClick={onClick}
			aria-label={tooltip}
		>
			{icon}
		</Button>
	);

	if (hoverTooltip) {
		return (
			<OverlayTrigger
				delay={{ show: 500, hide: 500 }}
				overlay={<Tooltip id="tooltip-help">{tooltip}</Tooltip>}
			>
				{button}
			</OverlayTrigger>
		);
	}
	return button;
};

export default ButtonWithTooltip;
