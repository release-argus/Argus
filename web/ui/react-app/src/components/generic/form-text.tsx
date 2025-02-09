import { Col, FormControl, FormGroup } from 'react-bootstrap';
import { FC, JSX } from 'react';
import {
	numberTest,
	regexTest,
	requiredTest,
	uniqueTest,
	urlTest,
} from './form-validate';

import FormLabel from './form-label';
import { Position } from 'types/config';
import { TooltipWithAriaProps } from './tooltip';
import cx from 'classnames';
import { formPadding } from './util';
import { useError } from 'hooks/errors';
import { useFormContext } from 'react-hook-form';

interface Props {
	name: string;
	registerParams?: Record<string, unknown>;
	required?: boolean | string;
	unique?: boolean;

	col_xs?: number;
	col_sm?: number;
	label?: string;
	smallLabel?: boolean;
	type?: 'text' | 'url';
	labelButton?: JSX.Element;

	isNumber?: boolean;
	isRegex?: boolean;
	isURL?: boolean;
	validationFunc?: (value: string) => boolean | string;
	defaultVal?: string;
	placeholder?: string;

	position?: Position;
	positionXS?: Position;
}

type FormTextProps = TooltipWithAriaProps & Props;

/**
 * A form text input item.
 *
 * @param name - The name of the form item.
 * @param registerParams - Additional parameters for the form item.
 * @param required - Whether the form item is required.
 * @param unique - Whether the form item should be unique.
 * @param col_xs - The number of columns the item takes up on XS+ screens.
 * @param col_sm - The number of columns the item takes up on SM+ screens.
 * @param label - The label of the form item.
 * @param smallLabel - Whether the label should be small.
 * @param tooltip - The tooltip of the form item.
 * @param tooltipAriaLabel - The aria label for the tooltip (Defaults to the tooltip).
 * @param type - The type of the form item.
 * @param isNumber - Whether the form item should be a number.
 * @param isRegex - Whether the form item should be a regex.
 * @param isURL - Whether the form item should be a URL.
 * @param validationFunc - The validation function for the form item.
 * @param defaultVal - The default value of the form item.
 * @param placeholder - The placeholder of the form item.
 * @param position - The position of the form item.
 * @param positionXS - The position of the form item on extra small screens.
 * @returns A form text input item at name with a label and tooltip.
 */
const FormText: FC<FormTextProps> = ({
	name,
	registerParams = {},
	required = false,
	unique,

	col_xs = 12,
	col_sm = 6,
	label,
	smallLabel,
	tooltip,
	tooltipAriaLabel,
	type = 'text',
	labelButton,

	isNumber,
	isRegex,
	isURL,
	validationFunc,
	defaultVal,
	placeholder,

	position = 'left',
	positionXS = position,
}) => {
	const { getValues, register, setError, clearErrors } = useFormContext();
	const error = useError(
		name,
		!!required ||
			isNumber ||
			isRegex ||
			isURL ||
			registerParams['validate'] !== undefined,
	);

	const padding = formPadding({ col_xs, col_sm, position, positionXS });
	const getTooltipProps = () => {
		if (!tooltip) return {};
		if (typeof tooltip === 'string') return { tooltip, tooltipAriaLabel };
		return { tooltip, tooltipAriaLabel };
	};

	return (
		<Col xs={col_xs} sm={col_sm} className={`${padding} pt-1 pb-1 col-form`}>
			<FormGroup>
				{label && (
					<div
						className={
							labelButton &&
							'd-flex justify-content-between align-items-center w-100'
						}
					>
						<FormLabel
							htmlFor={name}
							text={label}
							{...getTooltipProps()}
							required={required !== false}
							small={!!smallLabel}
						/>
						{labelButton}
					</div>
				)}
				<FormControl
					id={name}
					aria-label={`Value field for ${label}`}
					aria-describedby={cx(error && name + '-error')}
					type={type}
					placeholder={defaultVal || placeholder}
					autoFocus={false}
					{...register(name, {
						validate: {
							required: (value) =>
								requiredTest(
									value || defaultVal || '',
									name,
									setError,
									clearErrors,
									required,
								),
							isRegex: (value) => regexTest(value || defaultVal || '', isRegex),
							isNumber: (value) =>
								numberTest(value || defaultVal || '', isNumber),
							isUnique: (value) =>
								uniqueTest(value || defaultVal || '', name, getValues, unique),
							isURL: (value) => urlTest(value || defaultVal || '', isURL),
							validationFunc: (value) =>
								!validationFunc || validationFunc(value || defaultVal || ''),
						},
						...registerParams,
					})}
					isInvalid={!!error}
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

export default FormText;
