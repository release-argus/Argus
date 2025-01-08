import { Col, FormControl, FormGroup } from 'react-bootstrap';
import { FC, JSX } from 'react';

import FormLabel from './form-label';
import { Position } from 'types/config';
import { formPadding } from './util';
import { requiredTest } from './form-validate';
import { useError } from 'hooks/errors';
import { useFormContext } from 'react-hook-form';

interface Props {
	name: string;
	required?: boolean;

	col_xs?: number;
	col_sm?: number;
	label?: string;
	tooltip?: string | JSX.Element;

	defaultVal?: string;
	placeholder?: string;

	rows?: number;
	position?: Position;
	positionXS?: Position;
}

/**
 * A form textarea
 *
 * @param name - The name of the form item.
 * @param required - Whether the form item is required.
 * @param col_xs - The number of columns the item takes up on XS+ screens.
 * @param col_sm - The number of columns the item takes up on SM+ screens.
 * @param label - The label of the form item.
 * @param tooltip - The tooltip of the form item.
 * @param defaultVal - The default value of the form item.
 * @param placeholder - The placeholder of the form item.
 * @param rows - The number of rows for the textarea.
 * @param position - The position of the form item.
 * @param positionXS - The position of the form item on extra small screens.
 * @returns A form textarea with a label and tooltip.
 */
const FormTextArea: FC<Props> = ({
	name,
	required,

	col_xs = 12,
	col_sm = 6,
	label,
	tooltip,

	defaultVal,
	placeholder,

	rows,
	position = 'left',
	positionXS = position,
}) => {
	const { register, setError, clearErrors } = useFormContext();
	const error = useError(name, required);

	const padding = formPadding({ col_xs, col_sm, position, positionXS });

	return (
		<Col xs={col_xs} sm={col_sm} className={`${padding} pt-1 pb-1 col-form`}>
			<FormGroup>
				{label && (
					<FormLabel text={label} tooltip={tooltip} required={required} />
				)}
				<FormControl
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
					<small className="error-msg">{error['message'] || 'err'}</small>
				)}
			</FormGroup>
		</Col>
	);
};

export default FormTextArea;
