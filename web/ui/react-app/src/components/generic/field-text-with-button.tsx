import type { LucideIcon } from 'lucide-react';
import type { FC, FocusEventHandler } from 'react';
import { Controller, useFormContext, useWatch } from 'react-hook-form';
import FieldLabel, {
	type FieldLabelSize,
} from '@/components/generic/field-label';
import {
	type ColSize,
	getColSpanClasses,
} from '@/components/generic/field-shared';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import { Field, FieldError, FieldGroup } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';

type BaseProps = {
	/* The name of the field. */
	name: string;
	/* Whether the field is required. */
	required?: boolean | string;

	/* The column width on different screen sizes. */
	colSize?: ColSize;

	/* The form label to display. */
	label?: string;
	/* The size of the form label. */
	labelSize?: FieldLabelSize;

	/* The 'type' of the input. */
	type?: 'text' | 'url';
	/* The default value of the input. */
	defaultVal?: string;
	/* The placeholder of the input. */
	placeholder?: string;
};

/* Defines the action for a form field button. */
export type FieldButtonAction =
	| {
			/* Indicates the button triggers a click callback. */
			kind: 'click';
			/**
			 * Function to run when the button is clicked.
			 * Receives the current value of the associated form field.
			 */
			onClick: (value: string) => void;
	  }
	| {
			/* Indicates the button navigates to a link. */
			kind: 'link';
			/**
			 * Function that returns the hyperlink URL to navigate to.
			 * Receives the current value of the associated form field.
			 */
			href: (value: string) => string;
			/* Optional target attribute for the anchor element. */
			target?: string;
			/* Optional rel attribute for the anchor element. */
			rel?: string;
	  };

/* Configuration for a button associated with a form field. */
export type FieldButton = {
	/* The icon to display inside the button. */
	Icon: LucideIcon;
	/* Aria-label for accessibility describing the button's action. */
	ariaLabel: string;
} & FieldButtonAction;

export type FieldTextWithButtonProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip?: TooltipWithAriaProps;
	/* Field button configuration. */
	button: FieldButton;
};

/**
 * A form text input item.
 *
 * @param name - The name of the form item.
 * @param required - Marks the form item as required.
 *
 * @param colSize - The column width on different screen sizes.
 *
 * @param label - The label of the form item.
 * @param labelSize - The size of the label.
 * @param tooltip - The tooltip on the field label.
 * @param tooltip.type - The tooltip content type: either 'string' for plain text or 'element' for a React element.
 * @param tooltip.side - The wide to render the tooltip content.
 * @param tooltip.size - The size of the tooltip.
 * @param tooltip.delayDuration - Time before rendering the tooltip.
 *
 * @param type - The type of the form item.
 * @param defaultVal - The default value of the form item.
 * @param placeholder - The placeholder of the form item.
 *
 * @param button - Configuration for the form field button. Specify `kind: 'click' | 'link'`.
 * @param button.Icon - Icon to display inside the button.
 * @param button.ariaLabel - Aria-label describing the button for accessibility.
 * @param button.onClick - Callback triggered when the button is clicked (only for `kind: 'click'`).
 * @param button.href - Function returning the URL to navigate to (only for `kind: 'link'`).
 * @param button.target - Optional target attribute for link buttons (`kind: 'link'`) (default: `_blank`).
 * @param button.rel - Optional rel attribute for link buttons (`kind: 'link'`) (default: `noopener noreferrer`).
 *
 * @returns A form text input item at `name` with a label and tooltip.
 */
const FieldTextWithButton: FC<FieldTextWithButtonProps> = ({
	name,
	required = false,

	colSize,

	label,
	labelSize,
	tooltip,

	type = 'text',
	defaultVal,
	placeholder,

	button,
}) => {
	const { control } = useFormContext();
	const value = useWatch({ name }) as string;
	const responsiveColSpan = getColSpanClasses(colSize, { sm: 6, xs: 12 });

	return (
		<FieldGroup className={cn(responsiveColSpan, 'py-1')}>
			<Controller
				control={control}
				defaultValue=""
				name={name}
				render={({ field, fieldState }) => {
					const showButton = Boolean(value && !fieldState.error && button);
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
								<FieldLabel
									htmlFor={name}
									required={required !== false}
									size={labelSize}
									text={label}
									tooltip={tooltip}
								/>
							)}
							<ButtonGroup>
								<Input
									{...fieldRest}
									aria-describedby={cn(tooltip && `${name}-tooltip`)}
									aria-invalid={fieldState.invalid}
									aria-label={`Value field for ${label ?? name}`}
									id={name}
									onBlur={handleBlur}
									placeholder={defaultVal || placeholder}
									type={type}
								/>
								{showButton &&
									button &&
									(button.kind === 'click' ? (
										<Button
											aria-label={button.ariaLabel}
											className="h-full"
											onClick={() => button.onClick(value)}
											variant="outline"
										>
											<button.Icon />
										</Button>
									) : (
										<Button
											aria-label={button.ariaLabel}
											asChild
											variant="outline"
										>
											<a
												aria-label={button.ariaLabel}
												href={button.href(value)}
												rel={button.rel ?? 'noopener noreferrer'}
												target={button.target ?? '_blank'}
											>
												<button.Icon />
											</a>
										</Button>
									))}
							</ButtonGroup>
							{fieldState.invalid && <FieldError errors={[fieldState.error]} />}
						</Field>
					);
				}}
			/>
		</FieldGroup>
	);
};

export default FieldTextWithButton;
