import { Col, FormGroup } from 'react-bootstrap';
import { Controller, useFormContext } from 'react-hook-form';
import { DndContext, DragEndEvent, closestCenter } from '@dnd-kit/core';
import { FC, JSX, useEffect, useState } from 'react';
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
import { formPadding } from './util';
import { useError } from 'hooks/errors';

type FormSelectCreatableSortableProps = {
	name: string;

	key?: string;
	col_xs?: number;
	col_sm?: number;
	col_md?: number;

	label?: string;
	smallLabel?: boolean;
	tooltip?: string | JSX.Element;

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
	position?: Position;
	positionXS?: Position;
};

/**
 * FormSelectCreatableSortable is a labelled creatable select form item that can have new options
 * typed aed created, and is sorted by pick-order and can re-arranged by dragging.
 *
 *
 * @param name - The name of the form item.
 *
 * @param key - The key of the form item.
 * @param col_xs - The number of columns the item takes up on XS+ screens.
 * @param col_sm - The number of columns the item takes up on SM+ screens.
 * @param col_md - The number of columns the item takes up on MD+ screens.
 *
 * @param label - The label of the form item.
 * @param smallLabel - Whether the label should be small.
 * @param tooltip - The tooltip of the form item.
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
 * @param position - The position of the form item.
 * @param positionXS - The position of the form item on extra small screens.
 * @returns A labeled select form item.
 */
const FormSelectCreatableSortable: FC<FormSelectCreatableSortableProps> = ({
	name,

	key = name,
	col_xs = 12,
	col_sm = 6,
	col_md = col_sm,

	label,
	smallLabel,
	tooltip,

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
	position = 'left',
	positionXS = position,
}) => {
	const error = useError(name, customValidation !== undefined);
	const { setValue } = useFormContext();

	const [creatableOptions, setCreatableOptions] = useState<OptionType[]>([]);
	const [selectedOptions, setSelectedOptions] = useState<string[]>(
		initialValue ?? [],
	);
	useEffect(() => setSelectedOptions(initialValue ?? []), [initialValue]);

	useEffect(() => {
		if (options.length === 0) return;
		setCreatableOptions(convertStringArrayToOptionTypeArray(options, true));
	}, [options]);

	const padding = formPadding({ col_xs, col_sm, position, positionXS });

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

	const handleDragEnd = (event: DragEndEvent) => {
		const { active, over } = event;

		if (active.id !== over?.id) {
			const oldIndex = selectedOptions.findIndex((opt) => opt === active.id);
			const newIndex = selectedOptions.findIndex((opt) => opt === over?.id);
			const newOrder = arrayMove(selectedOptions, oldIndex, newIndex);

			setSelectedOptions(newOrder);
			setValue(
				name,
				newOrder.map((opt) => opt),
			);
		}
	};

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

	return (
		<Col
			xs={col_xs}
			sm={col_sm}
			md={col_md}
			className={`${padding} pt-1 pb-1 col-form`}
			key={key}
		>
			<FormGroup>
				{label && (
					<FormLabel text={label} tooltip={tooltip} small={!!smallLabel} />
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
									aria-label={`Select options for ${label}`}
									aria-describedby={`${name}-error ${name}-tooltip`}
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
					<small className="error-msg">{error['message'] || 'err'}</small>
				)}
			</FormGroup>
		</Col>
	);
};

export default FormSelectCreatableSortable;
