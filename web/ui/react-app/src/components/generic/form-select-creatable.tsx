import { Col, FormGroup } from 'react-bootstrap';
import {
	ConditionalOnChangeProps,
	convertStringArrayToOptionTypeArray,
	createOption,
	customComponents,
	customOnChange,
	customStyles,
	handleSelectedChange,
} from './form-select-shared';
import { FC, JSX, useEffect, useState } from 'react';
import { MultiValue, SingleValue } from 'react-select';

import { Controller } from 'react-hook-form';
import CreatableSelect from 'react-select/creatable';
import FormLabel from './form-label';
import { OptionType } from 'types/util';
import { Position } from 'types/config';
import { formPadding } from './util';
import { useError } from 'hooks/errors';

type Props = {
	name: string;

	key?: string;
	col_xs?: number;
	col_sm?: number;
	col_md?: number;

	label?: string;
	smallLabel?: boolean;
	tooltip?: string | JSX.Element;

	isMulti?: boolean;
	isClearable?: boolean;
	noOptionsMessage?: string;

	options: OptionType[] | string[];
	optionCounts?: boolean;
	customValidation?: (value: string) => string | boolean;

	position?: Position;
	positionXS?: Position;
};

type FormSelectCreatableProps = Props & ConditionalOnChangeProps;
/**
 * FormSelectCreatable is a labelled select form item that can have new options typed and created.
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
 * @param isMulti - Whether the select field should allow multiple values.
 * @param isClearable - Whether the select field should have a clear button.
 * @param noOptionsMessage - The text to display when no options are available for selection.
 *
 * @param options - The options for the select field.
 * @param optionCounts - Whether to show the selected count on the option labels.
 * @param customValidation - Custom validation function for the form item.
 * @param onChange - The function to call when the form item changes.
 *
 * @param position - The position of the form item.
 * @param positionXS - The position of the form item on extra small screens.
 * @returns A labeled select form item.
 */
const FormSelectCreatable: FC<FormSelectCreatableProps> = ({
	name,

	key = name,
	col_xs = 12,
	col_sm = 6,
	col_md = col_sm,

	label,
	smallLabel,
	tooltip,

	isMulti,
	isClearable,
	noOptionsMessage,

	options,
	optionCounts,
	customValidation,
	onChange,

	position = 'left',
	positionXS = position,
}) => {
	const error = useError(name, customValidation !== undefined);

	const [creatableOptions, setCreatableOptions] = useState<OptionType[]>([]);
	useEffect(() => {
		setCreatableOptions(convertStringArrayToOptionTypeArray(options, true));
	}, [options]);

	const padding = formPadding({ col_xs, col_sm, position, positionXS });

	const handleCreate = (
		inputValue: string,
		onChange: (...event: any[]) => void,
		currentValues: string[],
	) => {
		const newOption = createOption(inputValue, 1);
		setCreatableOptions((prev) =>
			[...prev, newOption].toSorted((a, b) => a.label.localeCompare(b.label)),
		);

		onChange([...currentValues, inputValue]);
	};

	const handleOnSelect = (
		currentValue: string | string[],
		newValue: SingleValue<OptionType> | MultiValue<OptionType>,
		fieldOnChange: (...event: any[]) => void,
	) => {
		// Update counts on option labels.
		handleSelectedChange(
			currentValue,
			newValue,
			creatableOptions,
			optionCounts,
			setCreatableOptions,
		);
		if (onChange) {
			if (isMulti) customOnChange(newValue, { isMulti: true, onChange });
			else customOnChange(newValue, { isMulti: false, onChange });
			return;
		}

		// Multi-select case.
		if (Array.isArray(newValue) || newValue === null)
			fieldOnChange((newValue ?? []).map((option) => option.value));
		// Single-select case.
		else fieldOnChange([(newValue as OptionType).value]);
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
					render={({ field }) => (
						<CreatableSelect
							{...field}
							aria-label={`Select options for ${label}`}
							className="form-select-creatable"
							options={creatableOptions ?? []}
							onCreateOption={(inputValue: string) =>
								handleCreate(inputValue, field.onChange, field.value)
							}
							onChange={(newValue) =>
								handleOnSelect(field.value, newValue, field.onChange)
							}
							value={
								isMulti
									? creatableOptions.find((option) =>
											field.value?.includes(option?.value),
									  )
									: creatableOptions.find(
											(option) => field.value === option?.value,
									  )
							}
							isClearable={isClearable}
							isMulti={isMulti}
							closeMenuOnSelect={!isMulti}
							hideSelectedOptions={false}
							noOptionsMessage={
								noOptionsMessage ? () => noOptionsMessage : undefined
							}
							components={customComponents}
							styles={customStyles}
						/>
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

export default FormSelectCreatable;
