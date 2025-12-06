import type { FC } from 'react';
import { Controller, useFormContext } from 'react-hook-form';
import FieldLabel from '@/components/generic/field-label';
import {
	type ColSize,
	getColSpanClasses,
} from '@/components/generic/field-shared';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { Field, FieldError, FieldGroup } from '@/components/ui/field';
import { Textarea } from '@/components/ui/textarea';
import { cn } from '@/lib/utils';

type BaseProps = {
	/* The name of the field. */
	name: string;
	/* Whether the field is required. */
	required?: boolean;

	/* The column width on different screen sizes. */
	colSize?: ColSize;

	/* The form label to display. */
	label?: string;

	/* The default value of the input. */
	defaultVal?: string | null;
	/* The placeholder of the input. */
	placeholder?: string;

	/* The number of rows for the textarea. */
	rows?: number;
};

type FieldTextAreaProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip?: TooltipWithAriaProps;
};

/**
 * A form textarea
 *
 * @param name - The name of the form item.
 * @param required - Marks the form item as required; may provide the error text when missing.
 *
 * @param colSize - The column width on different screen sizes.
 *
 * @param label - The label of the form item.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - The tooltip content type: either 'string' for plain text or 'element' for a React element.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 *
 * @param defaultVal - The default value of the form item.
 * @param placeholder - The placeholder of the form item.
 *
 * @param rows - The number of rows for the textarea.
 * @returns A form textarea with a label and tooltip.
 */
const FieldTextArea: FC<FieldTextAreaProps> = ({
	name,
	required,

	colSize,

	label,
	tooltip,

	defaultVal,
	placeholder,

	rows,
}) => {
	const { control } = useFormContext();
	const responsiveColSpan = getColSpanClasses(colSize, { sm: 6, xs: 12 });

	return (
		<FieldGroup className={cn(responsiveColSpan, 'py-1')}>
			<Controller
				control={control}
				name={name}
				render={({ field, fieldState }) => (
					<Field data-invalid={fieldState.invalid}>
						{label && (
							<FieldLabel
								htmlFor={name}
								required={required}
								text={label}
								tooltip={tooltip}
							/>
						)}
						<Textarea
							{...field}
							aria-describedby={cn(tooltip && `${name}-tooltip`)}
							aria-invalid={fieldState.invalid}
							id={name}
							placeholder={defaultVal ?? placeholder}
							rows={rows}
						/>
						{fieldState.invalid && <FieldError errors={[fieldState.error]} />}
					</Field>
				)}
			/>
		</FieldGroup>
	);
};

export default FieldTextArea;
