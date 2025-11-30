import { QuestionMarkCircledIcon } from '@radix-ui/react-icons';
import { type ComponentProps, type FC, type JSX, memo } from 'react';
import Tip from '@/components/ui/tip';
import type { TooltipContent } from '@/components/ui/tooltip';
import type { ScreenBreakpoint } from '@/types/util';

type TooltipString = {
	type: 'string';
	content: string;
	ariaLabel?: string;
};
type TooltipElement = {
	type: 'element';
	content: JSX.Element;
	ariaLabel: string;
};

export type TooltipWithAriaProps = TooltipString | TooltipElement;

type BaseProps = {
	id?: string;
	side?: ComponentProps<typeof TooltipContent>['side'];
	size?: ScreenBreakpoint;
	delayDuration?: number;
};

// biome-ignore assist/source/useSortedKeys: descending order.
const sizeMap: Record<ScreenBreakpoint, number> = {
	xs: 12,
	sm: 16,
	md: 20,
	lg: 24,
	xl: 32,
	xxl: 40,
};

type HelpTooltipProps = TooltipWithAriaProps & BaseProps;

/**
 * A 'help' tooltip icon that shows extra information on `hover` or `focus`.
 *
 * Supports plain text or JSX elements as tooltip content.
 * Screen reader accessibility through appropriate ARIA labels.
 *
 * @param props - Configuration options for the help content.
 * @param props.id - A unique identifier for the tooltip trigger element.
 * @param props.side - Specifies the position of the tooltip relative-to the trigger.
 * @param props.size - Sets the icon size for the question mark symbol.
 * @param props.delayDuration - The delay in milliseconds before becoming visible after a `hover` or `focus` event.
 * @param props.type - The tooltip content type: either 'string' for plain text or 'element' for a React element.
 * @param props.content - The content to display inside the tooltip. Either plain text or JSX.
 * @param props.ariaLabel - An accessible label used by screen readers to describe the tooltip content.
 * Falls back to the text content of `tooltip`.
 */
const HelpTooltip: FC<HelpTooltipProps> = ({
	id,
	side = 'top',
	size = 'xs',
	delayDuration = 500,
	...props
}) => {
	let content: string | JSX.Element;
	let ariaLabel: string;

	if (props.type === 'string') {
		content = props.content;
		ariaLabel = props.ariaLabel ?? props.content;
	} else {
		content = props.content;
		ariaLabel = props.ariaLabel;
	}

	return (
		<Tip
			className="mb-auto ml-1 p-0.5 align-middle"
			content={content}
			contentProps={{ side }}
			delayDuration={delayDuration}
		>
			<QuestionMarkCircledIcon
				aria-label={ariaLabel}
				className="size-auto!"
				height={sizeMap[size]}
				id={id}
				width={sizeMap[size]}
			/>
		</Tip>
	);
};

export default memo(HelpTooltip);
