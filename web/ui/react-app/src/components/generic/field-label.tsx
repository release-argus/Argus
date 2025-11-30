import { cva, type VariantProps } from 'class-variance-authority';
import type { FC } from 'react';
import { HelpTooltip } from '@/components/generic';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { FieldLabel } from '@/components/ui/field';
import { cn } from '@/lib/utils';

const fieldLabelVariants = cva('gap-0 col-span-full', {
	defaultVariants: {
		size: 'default',
	},
	variants: {
		size: {
			default: 'text-base',
			sm: 'text-sm',
			xs: 'text-xs',
		},
	},
});

type BaseProps = {
	/* The unique identifier for the form label. */
	id?: string;
	/* The ID of the associated input element for accessibility. */
	htmlFor?: string;
	/* The text content of the label. */
	text: string;
	/* Indicates if the label represents a required field. */
	required?: boolean;
};

export type FieldLabelSize = VariantProps<typeof fieldLabelVariants>['size'];

type FieldLabelProps = BaseProps & {
	/* The size variant for the form label. */
	size?: FieldLabelSize;
	/* The tooltip on the field label. */
	tooltip?: TooltipWithAriaProps;
};

/**
 * A form label with an optional tooltip.
 *
 * @param id - The unique identifier for the form label.
 * @param htmlFor - The ID of the associated input element for accessibility.
 * @param text - The text content of the label.
 * @param size - The size variant for the form label.
 * @param required - Indicates if the label represents a required field.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - The tooltip content type: either 'string' for plain text or 'element' for a React element.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 * @param tooltip.tooltip - An optional configuration object specifying tooltip content and properties.
 */
const FieldLabelWithTooltip: FC<FieldLabelProps> = ({
	id,
	htmlFor,
	text,
	size,
	required,
	tooltip,
}) => {
	return (
		<FieldLabel
			className={cn(fieldLabelVariants({ size }))}
			htmlFor={htmlFor}
			id={id}
		>
			{text}
			{required && <span className="text-destructive">*</span>}
			{tooltip && (
				<HelpTooltip id={htmlFor && `${htmlFor}-tooltip`} {...tooltip} />
			)}
		</FieldLabel>
	);
};

export default FieldLabelWithTooltip;
