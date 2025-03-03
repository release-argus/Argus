import { Col, FormCheck as FormCheckRB, FormGroup } from 'react-bootstrap';

import { FC } from 'react';
import { FormCheckType } from 'react-bootstrap/esm/FormCheck';
import FormLabel from './form-label';
import { Position } from 'types/config';
import { TooltipWithAriaProps } from './tooltip';
import cx from 'classnames';
import { formPadding } from './form-shared';
import { useFormContext } from 'react-hook-form';

type Props = {
	name: string;
	className?: string;

	col_xs?: number;
	col_sm?: number;
	col_md?: number;
	col_lg?: number;
	size?: 'sm' | 'lg';

	label?: string;
	smallLabel?: boolean;
	type?: FormCheckType;

	positionXS?: Position;
	positionSM?: Position;
	positionMD?: Position;
	positionLG?: Position;
};

type FormCheckProps = TooltipWithAriaProps & Props;

/**
 * A form checkbox
 *
 * @param name - The name of the field.
 * @param className - Additional classes for the form item.
 *
 * @param col_xs - The number of columns the item takes up on XS+ screens.
 * @param col_sm - The number of columns the item takes up on SM+ screens.
 * @param col_md - The number of columns the item takes up on MD+ screens.
 * @param col_lg - The number of columns the item takes up on LG+ screens.
 * @param size - The size of the checkbox.
 *
 * @param label - The form label to display.
 * @param smallLabel - Whether the label should be small.
 * @param tooltip - The tooltip to display.
 * @param tooltipAriaLabel - The aria label for the tooltip (Defaults to the tooltip).
 * @param type - The type of the checkbox.
 * @param positionXS - The position of the field on XS+ screens.
 * @param positionSM - The position of the field on SM+ screens.
 * @param positionMD - The position of the field on MD+ screens.
 * @param positionLG - The position of the field on LG+ screens.
 * @returns A form checkbox with a label and tooltip.
 */
const FormCheck: FC<FormCheckProps> = ({
	name,
	className,

	col_xs = 12,
	col_sm = 6,
	col_md,
	col_lg,
	size = 'sm',

	label,
	smallLabel,
	tooltip,
	tooltipAriaLabel,
	type = 'checkbox',

	positionXS = 'left',
	positionSM,
	positionMD,
	positionLG,
}) => {
	const { register } = useFormContext();

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
