import { Col, FormCheck as FormCheckRB, FormGroup } from 'react-bootstrap';

import { FC } from 'react';
import { FormCheckType } from 'react-bootstrap/esm/FormCheck';
import FormLabel from './form-label';
import { Position } from 'types/config';
import { TooltipWithAriaProps } from './tooltip';
import { formPadding } from './util';
import { useFormContext } from 'react-hook-form';

type Props = {
	name: string;

	col_xs?: number;
	col_sm?: number;
	size?: 'sm' | 'lg';
	label?: string;
	smallLabel?: boolean;
	type?: FormCheckType;

	position?: Position;
	positionXS?: Position;
};

type FormCheckProps = TooltipWithAriaProps & Props;

/**
 * A form checkbox
 *
 * @param name - The name of the field.
 * @param col_xs - The number of columns the item takes up on XS+ screens.
 * @param col_sm - The number of columns the item takes up on SM+ screens.
 * @param size - The size of the checkbox.
 * @param label - The form label to display.
 * @param smallLabel - Whether the label should be small.
 * @param tooltip - The tooltip to display.
 * @param tooltipAriaLabel - The aria label for the tooltip (Defaults to the tooltip).
 * @param type - The type of the checkbox.
 * @param position - The position of the field.
 * @param positionXS - The position of the field on extra small screens.
 * @returns A form checkbox with a label and tooltip.
 */
const FormCheck: FC<FormCheckProps> = ({
	name,

	col_xs = 12,
	col_sm = 6,
	size = 'sm',
	label,
	smallLabel,
	tooltip,
	tooltipAriaLabel,
	type = 'checkbox',

	position = 'left',
	positionXS = position,
}) => {
	const { register } = useFormContext();

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
					<FormLabel
						htmlFor={name}
						text={label}
						{...getTooltipProps()}
						small={!!smallLabel}
					/>
				)}
				<FormCheckRB
					id={name}
					className={`form-check${size === 'lg' ? '-large' : ''}`}
					type={type}
					autoFocus={false}
					{...register(name)}
				/>
			</FormGroup>
		</Col>
	);
};

export default FormCheck;
