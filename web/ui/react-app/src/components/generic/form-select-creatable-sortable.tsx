import { Col, FormGroup } from 'react-bootstrap';
import { Controller, useFormContext } from 'react-hook-form';
import { DndContext, DragEndEvent, closestCenter } from '@dnd-kit/core';
import { FC, useEffect, useState } from 'react';
import { MenuPlacement, MultiValue } from 'react-select';
import {
	MultiValueRemove,
	convertStringArrayToOptionTypeArray,
	createOption,
	customOnChange,
	customStyles,
	customStylesFixedHeight,
	handleSelectedChange,
	sortableMultiValue,
} from './form-select-shared';
import { SortableContext, arrayMove } from '@dnd-kit/sortable';

import CreatableSelect from 'react-select/creatable';
import FormLabel from './form-label';
import { OptionType } from 'types/util';
import { Position } from 'types/config';
import { TooltipWithAriaProps } from './tooltip';
import cx from 'classnames';
import { formPadding } from './util';
import { useError } from 'hooks/errors';

type Props = {
	name: string;
	key?: string;

	col_xs?: number;
	col_sm?: number;
	col_md?: number;
	col_lg?: number;

	label?: string;
	smallLabel?: boolean;

	isClearable?: boolean;
	noOptionsMessage?: string;

	placeholder?: string;
	initialValue?: string[];
	options: OptionType[] | string[];
	optionCounts?: boolean;
	customValidation?: (value: string) => string | boolean;
	onChange?: (newValue: MultiValue<OptionType>) => void;

	dynamicHeight?: boolean;
	menuPlacement?: MenuPlacement;
	positionXS?: Position;
	positionSM?: Position;
	positionMD?: Position;
	positionLG?: Position;
};

type FormSelectCreatableSortableProps = TooltipWithAriaProps & Props;

/**
 * FormSelectCreatableSortable is a labelled creatable select form item that can have new options
 * typed aed created, and is sorted by pick-order and can re-arranged by dragging.
 *
 *
 * @param name - The name of the form item.
 * @param key - The key of the form item.
 *
 * @param col_xs - The number of columns the item takes up on XS+ screens.
 * @param col_sm - The number of columns the item takes up on SM+ screens.
 * @param col_md - The number of columns the item takes up on MD+ screens.
 * @param col_lg - The number of columns the item takes up on LG+ screens.
 *
 * @param label - The label of the form item.
 * @param smallLabel - Whether the label should be small.
 * @param tooltip - The tooltip of the form item.
 * @param tooltipAriaLabel - The aria label for the tooltip (Defaults to the tooltip).
 *
 * @param isClearable - Whether the select field should have a clear button.
 * @param noOptionsMessage - The text to display when no options are available for selection.
 *
 * @param placeholder - Placeholder text.
 * @param initialValue - The initial value for the field.
 * @param options - The options for the select field.
 * @param optionCounts - Whether to show the selected count on the option labels.
 * @param customValidation - Custom validation function for the form item.
 * @param onChange - The function to call when the form item changes.
 *
 * @param dynamicHeight - Whether the field can expand downwards when filled.
 * @param menuPlacement - Positioning of the options dropdown.
 * @param positionXS - The position of the form item on XS+ screens.
 * @param positionSM - The position of the form item on SM+ screens.
 * @param positionMD - The position of the form item on MD+ screens.
 * @param positionLG - The position of the form item on LG+ screens.
 * @returns A labeled select form item.
 */
const FormSelectCreatableSortable: FC<FormSelectCreatableSortableProps> = ({
	name,
	key = name,

	col_xs = 12,
	col_sm = 6,
	col_md,
	col_lg,

	label,
	smallLabel,
	tooltip,
	tooltipAriaLabel,

	isClearable,
	noOptionsMessage,

	initialValue,
	placeholder,
	options,
	optionCounts,
	customValidation,
	onChange,

	dynamicHeight,
	menuPlacement = 'auto',
	positionXS = 'left',
	positionSM,
	positionMD,
	positionLG,
}) => {
	const error = useError(name, customValidation !== undefined);
	const { setValue } = useFormContext();

	const [creatableOptions, setCreatableOptions] = useState<OptionType[]>([]);
	// Convert options to the correct object.
	useEffect(() => {
		if (options.length === 0) return;
		setCreatableOptions(convertStringArrayToOptionTypeArray(options, true));
	}, [options]);

	const [selectedOptions, setSelectedOptions] = useState<string[]>(
		initialValue ?? [],
	);
	useEffect(() => setSelectedOptions(initialValue ?? []), [initialValue]);

	const padding = formPadding({
		col_xs,
		col_sm,
		col_md,
		col_lg,
		positionXS,
		positionSM,
		positionMD,
		positionLG,
	});

	// Create a new option.
	const handleCreate = (
		inputValue: string,
		onChange: (...event: any[]) => void,
		currentValues: string[],
	) => {
		// Create the option.
		setCreatableOptions((prev) => [...prev, createOption(inputValue, 1)]);
		// Select the new option.
		setSelectedOptions((prev) => [...prev, inputValue]);

		// Update the form value.
		onChange([...currentValues, inputValue]);
	};

	// Swap the dragged option.
	const handleDragEnd = (event: DragEndEvent) => {
		const { active, over } = event;

		// If it hasn't moved, exit.
		if (active.id === over?.id) return;

		// Swap the indices.
		const oldIndex = selectedOptions.findIndex((opt) => opt === active.id);
		const newIndex = selectedOptions.findIndex((opt) => opt === over?.id);
		const newOrder = arrayMove(selectedOptions, oldIndex, newIndex);

		// Update the displayed ordering.
		setSelectedOptions(newOrder);
		// Update the field value with this new ordering.
		setValue(name, newOrder);
	};

	// Select/De-select an option.
	const handleOnSelect = (
		currentValue: string | string[],
		newValue: MultiValue<OptionType>,
		fieldOnChange: (...event: any[]) => void,
	): void => {
		// Update selected counts on option labels,
		// and ordering of selections.
		handleSelectedChange(
			currentValue,
			newValue,
			creatableOptions,
			optionCounts,
			setCreatableOptions,
			setSelectedOptions,
		);
		// Custom onChange.
		if (onChange) return customOnChange(newValue, { isMulti: true, onChange });

		// Multi-select case.
		fieldOnChange((newValue ?? []).map((option) => option.value));
	};

	const getTooltipProps = () => {
		if (!tooltip) return {};
		if (typeof tooltip === 'string') return { tooltip, tooltipAriaLabel };
		return { tooltip, tooltipAriaLabel };
	};

	return (
		<Col
			xs={col_xs}
			sm={col_sm}
			md={col_md}
			lg={col_lg}
			className={`${padding} pt-1 pb-1 col-form`}
			key={key}
		>
			<FormGroup>
				{label && (
					<FormLabel
						htmlFor={name}
						text={label}
						{...getTooltipProps()}
						small={!!smallLabel}
					/>
				)}
				<Controller
					name={name}
					defaultValue={initialValue ?? []}
					render={({ field }) => (
						<DndContext
							collisionDetection={closestCenter}
							onDragEnd={handleDragEnd}
						>
							<SortableContext items={field.value ?? []}>
								<CreatableSelect
									{...field}
									id={name}
									aria-label={`Select options for ${label}`}
									aria-describedby={cx(
										error && name + '-error',
										tooltip && name + '-tooltip',
									)}
									className="form-select-creatable"
									options={creatableOptions}
									onCreateOption={(inputValue: string) =>
										handleCreate(inputValue, field.onChange, field.value)
									}
									onChange={(newValue) =>
										handleOnSelect(field.value, newValue, field.onChange)
									}
									placeholder={placeholder}
									value={(field.value ?? []).map((option: string) =>
										createOption(option),
									)}
									isMulti
									isClearable={isClearable}
									closeMenuOnSelect={false}
									hideSelectedOptions={false}
									noOptionsMessage={
										noOptionsMessage ? () => noOptionsMessage : undefined
									}
									menuPlacement={menuPlacement}
									components={{
										MultiValue: sortableMultiValue,
										MultiValueRemove: MultiValueRemove,
									}}
									styles={
										dynamicHeight ? customStyles : customStylesFixedHeight
									}
								/>
							</SortableContext>
						</DndContext>
					)}
					rules={{
						validate: {
							customValidation: (value) =>
								customValidation ? customValidation(value) : undefined,
						},
					}}
				/>
				{error && (
					<small id={name + '-error'} className="error-msg" role="alert">
						{error['message'] || 'err'}
					</small>
				)}
			</FormGroup>
		</Col>
	);
};

export default FormSelectCreatableSortable;
