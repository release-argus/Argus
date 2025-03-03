import { Col, FormControl, FormGroup, InputGroup } from 'react-bootstrap';
import { useFormContext, useWatch } from 'react-hook-form';

import { FC } from 'react';
import FormLabel from './form-label';
import { Position } from 'types/config';
import cx from 'classnames';
import { formPadding } from './form-shared';
import { useError } from 'hooks/errors';

interface Props {
	name: string;

	col_xs?: number;
	col_sm?: number;
	col_md?: number;
	col_lg?: number;

	label: string;
	tooltip?: string;

	defaultVal?: string;

	positionXS?: Position;
	positionSM?: Position;
	positionMD?: Position;
	positionLG?: Position;
}

/**
 * A form item for a hex colour with a colour picker.
 *
 * @param name - The name of the field.
 *
 * @param col_xs - The number of columns the item takes up on XS+ screens.
 * @param col_sm - The number of columns the item takes up on SM+ screens.
 * @param col_md - The number of columns the item takes up on MD+ screens.
 * @param col_lg - The number of columns the item takes up on LG+ screens.
 *
 * @param label - The form label to display.
 * @param tooltip - The tooltip to display.
 *
 * @param defaultVal - The default value of the field.
 *
 * @param positionXS - The position of the form item on XS+ screens.
 * @param positionSM - The position of the form item on SM+ screens.
 * @param positionMD - The position of the form item on MD+ screens.
 * @param positionLG - The position of the form item on LG+ screens.
 * @returns A form item for a hex colour with a colour picker, label and tooltip.
 */
const FormColour: FC<Props> = ({
	name,

	col_xs = 12,
	col_sm = 6,
	col_md,
	col_lg,

	label,
	tooltip,

	defaultVal,

	positionXS = 'left',
	positionSM,
	positionMD,
	positionLG,
}) => {
	const { register, setValue } = useFormContext();
	const hexColour: string = useWatch({ name: name });
	const trimmedHex = hexColour?.replace('#', '');
	const error = useError(name, true);

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
	const setColour = (hex: string) =>
		setValue(name, hex.substring(1), { shouldDirty: true });

	return (
		<Col
			xs={col_xs}
			sm={col_sm}
			md={col_md}
			lg={col_lg}
			className={`${padding} pt-1 pb-1 col-form`}
		>
			<FormGroup style={{ display: 'flex', flexDirection: 'column' }}>
				<div>
					<FormLabel htmlFor={name} text={label} tooltip={tooltip} />
				</div>
				<div style={{ display: 'flex', flexWrap: 'nowrap' }}>
					<InputGroup className="mb-2">
						<InputGroup.Text aria-hidden="true">#</InputGroup.Text>
						<FormControl
							id={name}
							aria-describedby={cx(
								error && name + '-error',
								tooltip && name + '-tooltip',
							)}
							style={{ width: '25%' }}
							type="text"
							defaultValue={trimmedHex}
							placeholder={defaultVal}
							maxLength={6}
							autoFocus={false}
							{...register(name, {
								pattern: {
									value: /^[\da-f]{6}$|^$/i,
									message: 'Invalid colour hex',
								},
							})}
							isInvalid={error !== undefined}
						/>
						<FormControl
							aria-label="Select a colour"
							className="form-control-color"
							style={{ width: '30%' }}
							type="color"
							title="Choose your colour"
							value={`#${trimmedHex || defaultVal?.replace('#', '')}`}
							onChange={(event) => setColour(event.target.value)}
							autoFocus={false}
						/>
					</InputGroup>
				</div>
			</FormGroup>
			{error && (
				<small id={name + '-error'} className="error-msg" role="alert">
					{error['message'] || 'err'}
				</small>
			)}
		</Col>
	);
};

export default FormColour;
