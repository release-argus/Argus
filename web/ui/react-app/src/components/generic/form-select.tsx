import { Col, FormGroup } from 'react-bootstrap';
import {
	ConditionalOnChangeProps,
	customOnChange,
	customStyles,
} from './form-select-shared';

import { Controller } from 'react-hook-form';
import { FC } from 'react';
import FormLabel from './form-label';
import { OptionType } from 'types/util';
import { Position } from 'types/config';
import Select from 'react-select';
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

	label?: string;
	smallLabel?: boolean;

	creatable?: boolean;
	isClearable?: boolean;

	options: OptionType[];
	customValidation?: (value: string) => string | boolean;

	position?: Position;
	positionXS?: Position;
};

type FormSelectProps = TooltipWithAriaProps & Props & ConditionalOnChangeProps;

/**
 * FormSelect is a labelled select form item.
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
 * @param tooltipAriaLabel - The aria label for the tooltip (Defaults to the tooltip).
 *
 * @param isMulti - Whether the select field should allow multiple values.
 * @param isClearable - Whether the select field should have a clear button.
 *
 * @param options - The options for the select field.
 * @param customValidation - Custom validation function for the form item.
 * @param onChange - The function to call when the form item changes.
 *
 * @param position - The position of the form item.
 * @param positionXS - The position of the form item on extra small screens.
 * @returns A labeled select form item.
 */
const FormSelect: FC<FormSelectProps> = ({
	name,

	key = name,
	col_xs = 12,
	col_sm = 6,
	col_md = col_sm,

	label,
	smallLabel,
	tooltip,
	tooltipAriaLabel,

	isMulti,
	isClearable,

	options,
	customValidation,
	onChange,

	position = 'left',
	positionXS = position,
}) => {
	const error = useError(name, customValidation !== undefined);

	const padding = formPadding({ col_xs, col_sm, position, positionXS });
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
					render={({ field }) => (
						<Select
							{...field}
							id={name}
							aria-label={`Select option${isMulti ? 's' : ''} for ${label}`}
							aria-describedby={cx(
								error && name + '-error',
								tooltip && name + '-tooltip',
							)}
							options={options}
							value={
								isMulti
									? options.find((option) =>
											field.value?.includes(option?.value),
									  )
									: options.find((option) => field.value === option?.value) ??
									  options?.[0]
							}
							onChange={(newValue) => {
								if (onChange) {
									if (isMulti)
										customOnChange(newValue, { isMulti: true, onChange });
									else customOnChange(newValue, { isMulti: false, onChange });
									return;
								}

								if (Array.isArray(newValue)) {
									// Multi-select case.
									field.onChange(newValue.map((option) => option.value));
								} else if (newValue === null) {
									// Clear case.
									field.onChange([]);
								} else {
									// Single-select case.
									field.onChange((newValue as OptionType).value);
								}
							}}
							isSearchable
							isMulti={isMulti}
							isClearable={isClearable}
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
					<small id={name + '-error'} className="error-msg" role="alert">
						{error['message'] || 'err'}
					</small>
				)}
			</FormGroup>
		</Col>
	);
};

export default FormSelect;
