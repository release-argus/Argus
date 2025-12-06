import type { ComponentProps } from 'react';
import { Controller, useFormContext } from 'react-hook-form';
import type { MultiValue, SingleValue } from 'react-select';
import FieldLabelWithTooltip, {
	type FieldLabelSize,
} from '@/components/generic/field-label';
import {
	type ColSize,
	getColSpanClasses,
} from '@/components/generic/field-shared';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { Field, FieldError, FieldGroup } from '@/components/ui/field';
import type { OptionType } from '@/components/ui/react-select/custom-components';
import Select from '@/components/ui/react-select/select';
import { cn } from '@/lib/utils';

type OriginalOnChange<IsMulti extends boolean> = ComponentProps<
	typeof Select<OptionType, IsMulti>
>['onChange'];
type ModifiedOnChange<IsMulti extends boolean = false> = (
	...args: Parameters<NonNullable<OriginalOnChange<IsMulti>>>
) => IsMulti extends true
	? MultiValue<OptionType> | boolean
	: SingleValue<OptionType> | boolean;

type BaseProps<IsMulti extends boolean = false> = {
	/* The name of the field. */
	name: string;
	/* Whether the field is required. */
	required?: boolean;

	/* The key of the field. */
	key?: string;
	/* The column width on different screen sizes. */
	colSize?: ColSize;

	/* The form label to display. */
	label?: string;
	/* The size of the form label. */
	labelSize?: FieldLabelSize;

	/* Whether the select field should allow creating new options. */
	creatable?: boolean;
	/* Whether the select field should have a clear button. */
	isClearable?: boolean;
	/* Whether the select field accepts input to filter the options. */
	isSearchable?: boolean;

	/* The options for the select field. */
	options: OptionType[] | readonly OptionType[];
	/* The function to call when the form item changes. */
	onChange?: ModifiedOnChange<IsMulti>;
	/* Whether to show any error messages/styling. */
	showError?: boolean;
};

type FormSelectProps<IsMulti extends boolean = false> = BaseProps<IsMulti> & {
	/* Whether the select field should allow multiple values. */
	isMulti?: IsMulti;
	/* The tooltip on the field label. */
	tooltip?: TooltipWithAriaProps;
};

/**
 * A labelled select form item.
 *
 * @param name - The name of the form item.
 * @param required - Marks the form item as required; may provide the error text when missing.
 *
 * @param key - The key of the form item.
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
 * @param isMulti - Whether the select field should allow multiple values.
 * @param isClearable - Whether the select field should have a clear button.
 * @param isSearchable - Whether the select field should accept input to filter the available option.
 *
 * @param options - The options for the select field.
 * @param onChange - The function to call when the form item changes.
 * @param showError - Whether to show any error messages/styling.
 */
const FieldSelect = <IsMulti extends boolean = false>({
	name,
	required,

	key = name,
	colSize,

	label,
	labelSize,
	tooltip,

	isMulti,
	isClearable,
	isSearchable,

	options,
	onChange,
	showError = true,
}: FormSelectProps<IsMulti>) => {
	const { control } = useFormContext();
	const responsiveColSpan = getColSpanClasses(colSize, { sm: 6, xs: 12 });

	return (
		<FieldGroup className={cn(responsiveColSpan, 'py-1')} key={key}>
			<Controller
				control={control}
				name={name}
				render={({ field, fieldState }) => (
					<Field
						data-invalid={fieldState.invalid && showError}
						orientation="vertical"
					>
						{label && (
							<FieldLabelWithTooltip
								htmlFor={name}
								required={required}
								size={labelSize}
								text={label}
								tooltip={tooltip}
							/>
						)}
						<Select
							{...field}
							aria-describedby={cn(tooltip && `${name}-tooltip`)}
							aria-invalid={fieldState.invalid && showError}
							aria-label={`Select option${isMulti ? 's' : ''} for ${label ?? name}`}
							id={name}
							isClearable={isClearable}
							isMulti={isMulti}
							isSearchable={options.length > 5 || isSearchable}
							menuShouldScrollIntoView
							onChange={(newValue, actionMeta) => {
								let result = onChange?.(newValue, actionMeta);
								if (result === false) return;

								result = result ?? newValue;
								if (Array.isArray(result)) {
									// Multi-select case.
									field.onChange(
										result.map((option: OptionType) => option.value),
									);
								} else if (result === null) {
									// Clear case.
									field.onChange([]);
								} else {
									// Single-select case.
									field.onChange((result as OptionType).value);
								}
							}}
							options={options}
							// styles={customStyles}
							value={
								isMulti
									? options.filter((option) =>
											(
												field.value as OptionType['value'][] | undefined
											)?.includes(option.value),
										)
									: (options.find((option) => field.value === option.value) ??
										options?.[0])
							}
						/>
						{fieldState.invalid && showError && (
							<FieldError errors={[fieldState.error]} />
						)}
					</Field>
				)}
			/>
		</FieldGroup>
	);
};

export default FieldSelect;
