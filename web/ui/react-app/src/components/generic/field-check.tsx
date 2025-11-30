import type { FC } from 'react';
import { Controller, useFormContext } from 'react-hook-form';
import FieldLabelWithTooltip, {
	type FieldLabelSize,
} from '@/components/generic/field-label';
import {
	type ColSize,
	getColSpanClasses,
} from '@/components/generic/field-shared';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { Checkbox } from '@/components/ui/checkbox';
import { Field, FieldError, FieldGroup } from '@/components/ui/field';
import { cn } from '@/lib/utils';

type BaseProps = {
	/* The name of the field. */
	name: string;
	/* Classnames to apply to the field. */
	className?: string;
	/* Classnames to apply to the checkbox. */
	checkboxClassName?: string;

	/* The column width on different screen sizes. */
	colSize?: ColSize;

	/* The form label to display. */
	label?: string;
	/* The size of the form label. */
	labelSize?: FieldLabelSize;
};

type FieldCheckProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip?: TooltipWithAriaProps;
};

/**
 * A form checkbox
 *
 * @param name - The name of the field.
 * @param className - Additional classes for the form item.
 * @param checkboxClassName - Additional classes for the checkbox.
 *
 * @param colSize - The column width on different screen sizes.
 *
 * @param label - The form label to display.
 * @param labelSize - The size of the form label.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - The tooltip content type: either 'string' for plain text or 'element' for a React element.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 * @returns A form checkbox with a label and tooltip.
 */
const FieldCheck: FC<FieldCheckProps> = ({
	name,
	className,
	checkboxClassName,

	colSize,

	label,
	labelSize,
	tooltip,
}) => {
	const { control } = useFormContext();
	const responsiveColSpan = getColSpanClasses(colSize, { sm: 6, xs: 12 });

	return (
		<FieldGroup className={cn(responsiveColSpan, 'w-full py-1', className)}>
			<Controller
				control={control}
				defaultValue={false}
				name={name}
				render={({ field, fieldState }) => (
					<Field data-invalid={fieldState.invalid} orientation="vertical">
						{label && (
							<FieldLabelWithTooltip
								htmlFor={name}
								size={labelSize}
								text={label}
								tooltip={tooltip}
							/>
						)}
						<Checkbox
							aria-invalid={fieldState.invalid}
							checked={field.value}
							className={cn('min-h-9 w-full', checkboxClassName)}
							name={field.name}
							onCheckedChange={field.onChange}
						/>
						{fieldState.invalid && <FieldError errors={[fieldState.error]} />}
					</Field>
				)}
			/>
		</FieldGroup>
	);
};

export default FieldCheck;
