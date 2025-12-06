import {
	closestCenter,
	DndContext,
	type DragEndEvent,
	KeyboardSensor,
	PointerSensor,
	useSensor,
	useSensors,
} from '@dnd-kit/core';
import { arrayMove, SortableContext } from '@dnd-kit/sortable';
import { CheckIcon } from '@radix-ui/react-icons';
import { type FC, useEffect, useState } from 'react';
import { Controller, useFormContext } from 'react-hook-form';
import {
	components,
	type GroupBase,
	type MenuPlacement,
	type MultiValue,
	type OptionProps,
} from 'react-select';
import FieldLabelWithTooltip, {
	type FieldLabelSize,
} from '@/components/generic/field-label';
import { createOption } from '@/components/generic/field-select-shared';
import {
	type ColSize,
	getColSpanClasses,
} from '@/components/generic/field-shared';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import { Field, FieldError, FieldGroup } from '@/components/ui/field';
import {
	type OptionType,
	SortableMultiValue,
} from '@/components/ui/react-select/custom-components';
import Select from '@/components/ui/react-select/select';
import { cn } from '@/lib/utils';

type BaseProps = {
	/* The name of the field. */
	name: string;

	/* The column width on different screen sizes. */
	colSize?: ColSize;

	/* The form label to display. */
	label?: string;
	/* The size of the form label. */
	labelSize?: FieldLabelSize;

	/* Whether the select field should have a clear button. */
	isClearable?: boolean;
	/* The text to display when no options available for selection. */
	noOptionsMessage?: string;

	/* Placeholder text. */
	placeholder?: string;
	/* The initial value for the field. */
	initialValue?: string[];
	/* The options for the select field. */
	options: OptionType[];
	/* The counts of the selected options outside this field. */
	counts?: Record<string, number>;
	/* The function to call when the form item changes. */
	onChange?: (newValue: MultiValue<OptionType>) => void;

	/* Positioning of the dropdown. */
	menuPlacement?: MenuPlacement;
};

type FieldSelectCreatableSortableProps = BaseProps & {
	/* The tooltip on the field label. */
	tooltip?: TooltipWithAriaProps;
};

/**
 * A labelled select form item supporting creation of new options,
 * sorted by pick-order and re-arrangeable by dragging.
 *
 *
 * @param name - The name of the form item.
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
 * @param isClearable - Whether the select field should have a clear button.
 * @param noOptionsMessage - The text to display when no options available for selection.
 *
 * @param placeholder - Placeholder text.
 * @param initialValue - The initial value for the field.
 * @param options - The options for the select field.
 * @param optionCounts - Whether to show the selected count on the option labels.
 * @param onChange - The function to call when the form item changes.
 *
 * @param menuPlacement - Positioning of the dropdown.
 * @returns A labelled select form item.
 */
const FieldSelectCreatableSortable: FC<FieldSelectCreatableSortableProps> = ({
	name,

	colSize,

	label,
	labelSize = 'sm',
	tooltip,

	isClearable,
	noOptionsMessage,

	initialValue,
	placeholder,
	options,
	counts,
	onChange,

	menuPlacement = 'auto',
}) => {
	const { control, setValue } = useFormContext();

	const [creatableOptions, setCreatableOptions] =
		useState<OptionType[]>(options);
	// Convert options to the correct object.
	useEffect(() => {
		if (options.length === 0) return;
		setCreatableOptions(options);
	}, [options]);

	// Setup dnd-kit sensors to detect drag events.
	const sensors = useSensors(
		useSensor(PointerSensor, {
			// Don't start dragging immediately to allow for clicks.
			activationConstraint: {
				distance: 5,
			},
		}),
		useSensor(KeyboardSensor),
	);

	const responsiveColSpan = getColSpanClasses(colSize, { sm: 6, xs: 12 });

	return (
		<FieldGroup className={cn(responsiveColSpan, 'py-1')}>
			<Controller
				control={control}
				defaultValue={initialValue ?? []}
				name={name}
				render={({ field, fieldState }) => {
					const selectedValues = (
						Array.isArray(field.value) ? field.value : []
					) as string[];

					const handleCreateOption = (inputValue: string) => {
						const newOption = { label: inputValue, value: inputValue };
						const updatedOptions = [...creatableOptions, newOption];
						setCreatableOptions(updatedOptions);

						const newValues = [...selectedValues, inputValue];

						if (onChange) {
							onChange(updatedOptions);
						} else {
							field.onChange(newValues);
						}
					};

					// Swap the dragged option.
					const handleDragEnd = (event: DragEndEvent) => {
						const { active, over } = event;
						// If it hasn't moved, exit.
						if (active.id === over?.id) return;

						// Swap the indexes.
						const oldIndex = selectedValues.indexOf(String(active.id));
						const newIndex = over
							? selectedValues.indexOf(String(over.id))
							: -1;
						const newOrder = arrayMove(selectedValues, oldIndex, newIndex);

						// Update the field value with this new ordering.
						setValue(name, newOrder);
					};

					// Select/De-select an option.
					const handleOnChange = (newValue: MultiValue<OptionType>) => {
						// Custom onChange.
						if (onChange) {
							onChange(newValue);
						} else {
							// Multi-select case.
							field.onChange(newValue.map((option) => option.value));
						}
					};

					const handleFormatOptionLabel = (
						data: OptionType,
						_formatOptionLabelMeta?: unknown,
					) => {
						const count =
							(selectedValues.includes(data.value) ? 1 : 0) +
							(counts && data.value in counts ? counts[data.value] : 0);
						return count ? `${data.label} (${count})` : data.label;
					};

					const FormattedOption = <
						FinalOptionType extends OptionType,
						IsMulti extends boolean,
						Group extends GroupBase<FinalOptionType>,
					>(
						props: OptionProps<FinalOptionType, IsMulti, Group>,
					) => (
						<components.Option {...props}>
							<div className="flex items-center justify-between">
								<div>
									{counts
										? handleFormatOptionLabel(props.data)
										: props.data.label}
								</div>
								{props.isSelected && <CheckIcon className="shrink-0" />}
							</div>
						</components.Option>
					);

					const valueAsOptions = selectedValues.map(
						(val) =>
							creatableOptions.find((opt) => opt.value === val) ??
							createOption(val),
					);

					return (
						<Field data-invalid={fieldState.invalid} orientation="vertical">
							{label && (
								<FieldLabelWithTooltip
									{...field}
									htmlFor={name}
									required={false}
									size={labelSize}
									text={label}
									tooltip={tooltip}
								/>
							)}
							<DndContext
								collisionDetection={closestCenter}
								onDragEnd={handleDragEnd}
								sensors={sensors}
							>
								<SortableContext items={selectedValues.map(String)}>
									<Select
										{...field}
										aria-describedby={cn(tooltip && `${name}-tooltip`)}
										aria-invalid={fieldState.invalid}
										aria-label={`Select options for ${label ?? name}`}
										closeMenuOnSelect={false}
										components={{
											MultiValue: SortableMultiValue,
											Option: FormattedOption,
										}}
										formatOptionLabel={counts && handleFormatOptionLabel}
										hideSelectedOptions={false}
										id={name}
										isClearable={isClearable}
										isCreatable
										isMulti
										menuPlacement={menuPlacement}
										menuShouldScrollIntoView
										noOptionsMessage={
											noOptionsMessage ? () => noOptionsMessage : undefined
										}
										onChange={handleOnChange}
										onCreateOption={handleCreateOption}
										options={creatableOptions}
										placeholder={placeholder}
										value={valueAsOptions}
									/>
								</SortableContext>
							</DndContext>
							{fieldState.invalid && <FieldError errors={[fieldState.error]} />}
						</Field>
					);
				}}
			/>
		</FieldGroup>
	);
};

export default FieldSelectCreatableSortable;
