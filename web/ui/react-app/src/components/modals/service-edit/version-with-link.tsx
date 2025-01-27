import {
	Button,
	Col,
	FormControl,
	FormGroup,
	InputGroup,
} from 'react-bootstrap';
import { FC, useState } from 'react';
import {
	repoTest,
	requiredTest,
	urlTest,
} from 'components/generic/form-validate';
import { useFormContext, useWatch } from 'react-hook-form';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { FormLabel } from 'components/generic/form';
import { Position } from 'types/config';
import { TooltipWithAriaProps } from 'components/generic/tooltip';
import cx from 'classnames';
import { faLink } from '@fortawesome/free-solid-svg-icons';
import { formPadding } from 'components/generic/util';
import { useError } from 'hooks/errors';

interface Props {
	name: string;
	type: 'github' | 'url';
	required?: boolean;

	col_xs?: number;
	col_sm?: number;
	col_md?: number;
	position?: Position;
}

type VersionWithLinkProps = TooltipWithAriaProps & Props;

/**
 * The version field with a link to the source being monitored.
 *
 * @param name - The name of the field in the form.
 * @param type - The type of version field.
 * @param required - Whether the field is required.
 * @param col_xs - The amount of columns the item takes up on XS+ screens.
 * @param col_sm - The amount of columns the item takes up on SM+ screens.
 * @param col_md - The amount of columns the item takes up on MD+ screens.
 * @param tooltip - The tooltip for the field.
 * @param position - The position of the field.
 * @returns The version field with a link to the source being monitored.
 */
const VersionWithLink: FC<VersionWithLinkProps> = ({
	name,
	type,
	required,
	col_xs = 12,
	col_sm = 6,
	col_md = col_sm,
	tooltip,
	tooltipAriaLabel,
	position,
}) => {
	const { register, setError, clearErrors } = useFormContext();
	const value: string = useWatch({ name: name });
	const error = useError(name, true);

	const [isUnfocused, setIsUnfocused] = useState(true);
	const handleFocus = () => setIsUnfocused(false);
	const handleBlur = () => setIsUnfocused(true);

	const link = (type: 'github' | 'url') =>
		type === 'github' ? `https://github.com/${value}` : value;
	const padding = formPadding({ col_xs, col_sm, position });

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
		>
			<FormGroup>
				<FormLabel
					htmlFor={name}
					text={type === 'github' ? 'Repository' : 'URL'}
					{...getTooltipProps()}
					required={required}
				/>
				<InputGroup className="me-3">
					<FormControl
						id={name}
						aria-label={type === 'github' ? 'GitHub repository' : 'URL'}
						aria-describedby={cx(
							error && name + '-error',
							tooltip && name + '-tooltip',
						)}
						defaultValue={value}
						onFocus={handleFocus}
						{...register(name, {
							validate: {
								required: (value) =>
									requiredTest(value, name, setError, clearErrors, required),
								isType: (value) =>
									type === 'url' ? urlTest(value, true) : repoTest(value, true),
							},
							onBlur: handleBlur,
						})}
						isInvalid={!!error}
					/>
					{isUnfocused && value && !error && (
						<a href={link(type)} target="_blank">
							<Button
								aria-label={
									type === 'github' ? 'Open GitHub repository' : 'Open URL'
								}
								variant="secondary"
								className="curved-right-only"
							>
								<FontAwesomeIcon icon={faLink} />
							</Button>
						</a>
					)}
				</InputGroup>
				{error && (
					<small id={name + '-error'} className="error-msg" role="alert">
						{error['message'] || 'err'}
					</small>
				)}
			</FormGroup>
		</Col>
	);
};

export default VersionWithLink;
