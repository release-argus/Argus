import type { FC, FocusEventHandler } from 'react';
import { Controller, useFormContext, useWatch } from 'react-hook-form';
import FieldLabelWithTooltip from '@/components/generic/field-label';
import {
	type ColSize,
	getColSpanClasses,
} from '@/components/generic/field-shared';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { Field, FieldError, FieldGroup } from '@/components/ui/field';
import {
	InputGroup,
	InputGroupAddon,
	InputGroupInput,
} from '@/components/ui/input-group';
import { cn } from '@/lib/utils';

type BaseProps = {
	/* The name of the field. */
	name: string;

	/* The column width on different screen sizes. */
	colSize?: ColSize;

	/* The form label to display. */
	label: string;

	/* The default value of the field. */
	defaultVal?: string;
};

type FieldColourProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip?: TooltipWithAriaProps;
};

/**
 * A labelled form item for a hex colour (with colour picker).
 *
 * @param name - The name of the field.
 *
 * @param colSize - The column width on different screen sizes.
 *
 * @param label - The form label to display.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - The tooltip content type: either 'string' for plain text or 'element' for a React element.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 *
 * @param defaultVal - The default value of the field.
 */
const FieldColour: FC<FieldColourProps> = ({
	name,

	colSize,

	label,
	tooltip,

	defaultVal,
}) => {
	const { control } = useFormContext();
	const hexColour = useWatch({ name: name }) as string | undefined;
	const trimmedHex = hexColour?.replace('#', '');
	const hex = (trimmedHex || defaultVal?.replace('#', '')) ?? '';
	const responsiveColSpan = getColSpanClasses(colSize, { sm: 6, xs: 12 });

	return (
		<FieldGroup className={cn(responsiveColSpan, 'py-1')}>
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
									text={label}
									tooltip={tooltip}
								/>
							)}

							<InputGroup className="overflow-clip">
								<InputGroupAddon>#</InputGroupAddon>
								<InputGroupInput
									{...fieldRest}
									aria-describedby={cn(tooltip && `${name}-tooltip`)}
									aria-invalid={fieldState.invalid}
									autoFocus={false}
									id={name}
									maxLength={7}
									onBlur={handleBlur}
									onChange={(event) => {
										let colour = event.target.value.replaceAll('#', '');
										colour = colour.slice(0, 6);
										field.onChange(colour);
									}}
									placeholder={defaultVal}
									type="text"
								/>
								<InputGroupInput
									aria-label="Select a colour"
									autoFocus={false}
									className="h-full w-[30%] p-0"
									onChange={(event) => {
										const colour = event.target.value.replaceAll('#', '');
										field.onChange(colour);
									}}
									title="Choose your colour"
									type="color"
									value={`#${hex}`}
								/>
							</InputGroup>
							{fieldState.invalid && <FieldError errors={[fieldState.error]} />}
						</Field>
					);
				}}
			/>
		</FieldGroup>
	);
};

export default FieldColour;
