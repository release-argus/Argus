import { Col, FormControl, FormGroup } from 'react-bootstrap';

import { FC } from 'react';
import FormLabel from './form-label';
import { Position } from 'types/config';
import { TooltipWithAriaProps } from './tooltip';
import cx from 'classnames';
import { formPadding } from './form-shared';
import { requiredTest } from './form-validate';
import { useError } from 'hooks/errors';
import { useFormContext } from 'react-hook-form';

interface Props {
	name: string;
	required?: boolean;

	col_xs?: number;
	col_sm?: number;
	col_md?: number;
	col_lg?: number;

	label?: string;

	defaultVal?: string;
	placeholder?: string;

	rows?: number;
	positionXS?: Position;
	positionSM?: Position;
	positionMD?: Position;
	positionLG?: Position;
}

type FormTextAreaProps = TooltipWithAriaProps & Props;

/**
 * A form textarea
 *
 * @param name - The name of the form item.
 * @param required - Whether the form item is required.
 *
 * @param col_xs - The number of columns the item takes up on XS+ screens.
 * @param col_sm - The number of columns the item takes up on SM+ screens.
 * @param col_md - The number of columns the item takes up on MD+ screens.
 * @param col_lg - The number of columns the item takes up on LG+ screens.
 *
 * @param label - The label of the form item.
 * @param tooltip - The tooltip of the form item.
 * @param tooltipAriaLabel - The aria label for the tooltip (Defaults to the tooltip).
 *
 * @param defaultVal - The default value of the form item.
 * @param placeholder - The placeholder of the form item.
 *
 * @param rows - The number of rows for the textarea.
 * @param positionXS - The position of the form item on XS+ screens.
 * @param positionSM - The position of the form item on SM+ screens.
 * @param positionMD - The position of the form item on MD+ screens.
 * @param positionLG - The position of the form item on LG+ screens.
 * @returns A form textarea with a label and tooltip.
 */
const FormTextArea: FC<FormTextAreaProps> = ({
	name,
	required,

	col_xs = 12,
	col_sm = 6,
	col_md,
	col_lg,

	label,
	tooltip,
	tooltipAriaLabel,

	defaultVal,
	placeholder,

	rows,
	positionXS = 'left',
	positionSM,
	positionMD,
	positionLG,
}) => {
	const { register, setError, clearErrors } = useFormContext();
	const error = useError(name, required);

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
			className={`${padding} pt-1 pb-1 col-form`}
		>
			<FormGroup>
				{label && (
					<FormLabel
						htmlFor={name}
						text={label}
						{...getTooltipProps()}
						required={required}
					/>
				)}
				<FormControl
					id={name}
					aria-describedby={cx(
						error && name + '-error',
						tooltip && name + '-tooltip',
					)}
					type={'textarea'}
					as="textarea"
					rows={rows}
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
						},
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

export default FormTextArea;
