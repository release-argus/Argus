import { Col, FormControl, FormGroup } from 'react-bootstrap';
import {
	numberTest,
	regexTest,
	requiredTest,
	uniqueTest,
	urlTest,
} from './form-validate';

import { FC } from 'react';
import FormLabel from './form-label';
import { Position } from 'types/config';
import { TooltipWithAriaProps } from './tooltip';
import cx from 'classnames';
import { formPadding } from './util';
import { useError } from 'hooks/errors';
import { useFormContext } from 'react-hook-form';

interface Props {
	name: string;
	className?: string;
	registerParams?: Record<string, unknown>;
	required?: boolean | string;
	unique?: boolean;

	col_xs?: number;
	col_sm?: number;
	col_md?: number;
	col_lg?: number;

	label?: string;
	smallLabel?: boolean;

	type?: 'text' | 'url';
	isNumber?: boolean;
	isRegex?: boolean;
	isURL?: boolean;
	validationFunc?: (value: string) => boolean | string;
	defaultVal?: string;
	placeholder?: string;

	positionXS?: Position;
	positionSM?: Position;
	positionMD?: Position;
	positionLG?: Position;
}

type FormTextProps = TooltipWithAriaProps & Props;

/**
 * A form text input item.
 *
 * @param name - The name of the form item.
 * @param className - Additional classes for the form item.
 * @param registerParams - Additional parameters for the form item.
 * @param required - Whether the form item is required.
 * @param unique - Whether the form item should be unique.
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
 * @param type - The type of the form item.
 * @param isNumber - Whether the form item should be a number.
 * @param isRegex - Whether the form item should be a regex.
 * @param isURL - Whether the form item should be a URL.
 * @param validationFunc - The validation function for the form item.
 * @param defaultVal - The default value of the form item.
 * @param placeholder - The placeholder of the form item.
 *
 * @param positionXS - The position of the form item on XS+ screens.
 * @param positionSM - The position of the form item on SM+ screens.
 * @param positionMD - The position of the form item on MD+ screens.
 * @param positionLG - The position of the form item on LG+ screens.
 * @returns A form text input item at name with a label and tooltip.
 */
const FormText: FC<FormTextProps> = ({
	name,
	className,
	registerParams = {},
	required = false,
	unique,

	col_xs = 12,
	col_sm = 6,
	col_md,
	col_lg,

	label,
	smallLabel,
	tooltip,
	tooltipAriaLabel,

	type = 'text',
	isNumber,
	isRegex,
	isURL,
	validationFunc,
	defaultVal,
	placeholder,

	positionXS = 'left',
	positionSM,
	positionMD,
	positionLG,
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
			className={cx(`${padding} pt-1 pb-1 col-form`, className)}
		>
			<FormGroup>
				{label && (
					<FormLabel
						htmlFor={name}
						text={label}
						{...getTooltipProps()}
						required={required !== false}
						small={!!smallLabel}
					/>
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
