import type { FC, FocusEventHandler } from 'react';
import { Controller, useFormContext } from 'react-hook-form';
import FieldLabelWithTooltip, {
	type FieldLabelSize,
} from '@/components/generic/field-label';
import {
	type ColSize,
	getColSpanClasses,
} from '@/components/generic/field-shared';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { Field, FieldError, FieldGroup } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';

type BaseProps = {
	/* The name of the field. */
	name: string;
	/* Classnames to apply to the field. */
	className?: string;
	/* Whether the field is required. */
	required?: boolean | string;

	/* The column width on different screen sizes. */
	colSize?: ColSize;

	/* The form label to display. */
	label?: string;
	/* The size of the form label. */
	labelSize?: FieldLabelSize;

	/* The 'type' of the input. */
	type?: 'text' | 'url' | 'number';
	/* The default value of the input. */
	defaultVal?: string | null;
	/* The placeholder of the input. */
	placeholder?: string;
};

type FieldTextProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip?: TooltipWithAriaProps;
};

/**
 * A form text input item.
 *
 * @param name - The name of the form item.
 * @param className - Additional classes for the form item.
 * @param required - Marks the form item as required; may provide the error text when missing.
 *
 * @param colSize - The column width on different screen sizes.
 *
 * @param label - The label of the form item.
 * @param labelSize - The size of the form label.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - The tooltip content type: either 'string' for plain text or 'element' for a React element.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 *
 * @param type - The type of the form item.
 * @param defaultVal - The default value of the form item.
 * @param placeholder - The placeholder of the form item.
 * @returns A form text input item at `name` with a label and tooltip.
 */
const FieldText: FC<FieldTextProps> = ({
	name,
	className,
	required = false,

	colSize,

	label,
	labelSize,
	tooltip,

	type = 'text',
	defaultVal,
	placeholder,
}) => {
	const { control } = useFormContext();
	const responsiveColSpan = getColSpanClasses(colSize, { sm: 6, xs: 12 });

	return (
		<FieldGroup className={cn(responsiveColSpan, 'py-1', className)}>
			<Controller
				control={control}
				name={name}
				render={({ field, fieldState }) => {
					// Defer onBlur to allow the browser to complete focus traversal (e.g. Tab)
					// before validation triggers re-renders. This prevents focus from jumping to
					// the start of the container when an invalid field becomes valid on blur.
					const { onBlur, ...fieldRest } = field;
					const handleBlur: FocusEventHandler<HTMLInputElement> = () => {
						// Call RHF onBlur after focus has moved.
						setTimeout(() => onBlur(), 0);
					};

					return (
						<Field data-invalid={fieldState.invalid}>
							{label && (
								<FieldLabelWithTooltip
									htmlFor={name}
									required={!!required}
									size={labelSize}
									text={label}
									tooltip={tooltip}
								/>
							)}
							<Input
								{...fieldRest}
								aria-invalid={fieldState.invalid}
								aria-label={`Value field for ${label ?? name}`}
								id={name}
								onBlur={handleBlur}
								placeholder={defaultVal ?? placeholder}
								type={type}
								value={fieldRest.value ?? ''}
							/>

							{fieldState.invalid && <FieldError errors={[fieldState.error]} />}
						</Field>
					);
				}}
			/>
		</FieldGroup>
	);
};

export default FieldText;
