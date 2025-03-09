import {
	Button,
	Col,
	FormControl,
	FormGroup,
	InputGroup,
} from 'react-bootstrap';
import { requiredTest, urlTest } from './form-validate';
import { useFormContext, useWatch } from 'react-hook-form';

import { FC } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import FormLabel from './form-label';
import { IconDefinition } from '@fortawesome/fontawesome-svg-core';
import { Position } from 'types/config';
import { TooltipWithAriaProps } from './tooltip';
import cx from 'classnames';
import { formPadding } from './form-shared';
import { useError } from 'hooks/errors';

interface Props {
	name: string;
	required?: boolean | string;

	col_xs?: number;
	col_sm?: number;
	col_md?: number;
	col_lg?: number;

	label?: string;
	smallLabel?: boolean;

	type?: 'text' | 'url';
	isURL?: boolean;
	validationFunc?: (value: string) => boolean | string;
	defaultVal?: string;
	placeholder?: string;

	buttonIcon: IconDefinition;
	buttonAriaLabel: string;
	buttonVariant?: string;
	buttonOnClick?: (value: string) => void;
	buttonHref?: (value: string) => string;
	showButtonCondition?: (value: string, hasError: boolean) => boolean;

	positionXS?: Position;
	positionSM?: Position;
	positionMD?: Position;
	positionLG?: Position;
}

type FormTextWithButtonProps = TooltipWithAriaProps & Props;

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
 * @param isURL - Whether the form item should be a URL.
 * @param validationFunc - The validation function for the form item.
 * @param defaultVal - The default value of the form item.
 * @param placeholder - The placeholder of the form item.
 *
 * @param buttonIcon - The icon for the button.
 * @param buttonAriaLabel - Aria Label for the button.
 * @param buttonVariant - Variant for the button.
 * @param buttonOnClick - Function to run when the button is clicked.
 * @param buttonHref - Hyperlink to navigate to when the button is clicked.
 * @param showButtonCondition - Condition behind whether to render the button in the UI.
 *
 * @param positionXS - The position of the form item on XS+ screens.
 * @param positionSM - The position of the form item on SM+ screens.
 * @param positionMD - The position of the form item on MD+ screens.
 * @param positionLG - The position of the form item on LG+ screens.
 * @returns A form text input item at name with a label and tooltip.
 */
const FormTextWithButton: FC<FormTextWithButtonProps> = ({
	name,
	required = false,

	col_xs = 12,
	col_sm = 6,
	col_md,
	col_lg,

	label,
	smallLabel,
	tooltip,
	tooltipAriaLabel,

	type = 'text',
	isURL,
	validationFunc,
	defaultVal,
	placeholder,

	buttonIcon,
	buttonAriaLabel,
	buttonVariant = 'secondary',
	buttonOnClick,
	buttonHref,
	showButtonCondition = (value, hasError) => value && !hasError,

	positionXS = 'left',
	positionSM,
	positionMD,
	positionLG,
}) => {
	const { register, setError, clearErrors } = useFormContext();
	const value = useWatch({ name });
	const error = useError(name, !!required || isURL);

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

	const ButtonWrapper = ({ children }: { children: React.ReactNode }) =>
		buttonHref ? (
			<a href={buttonHref(value)} target="_blank" rel="noopener noreferrer">
				{children}
			</a>
		) : (
			<>{children}</>
		);

	return (
		<Col
			xs={col_xs}
			sm={col_sm}
			md={col_md}
			lg={col_lg}
			className={cx(padding, 'pt-1 pb-1', 'col-form')}
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
				<InputGroup className="me-3">
					<FormControl
						id={name}
						aria-label={`Value field for ${label}`}
						aria-describedby={cx(
							error && `${name}-error`,
							tooltip && `${name}-tooltip`,
						)}
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
								isURL: (value) => urlTest(value || defaultVal || '', isURL),
								validationFunc: (value) =>
									!validationFunc || validationFunc(value || defaultVal || ''),
							},
						})}
						isInvalid={!!error}
					/>
					{showButtonCondition(value, !!error) && (
						<ButtonWrapper>
							<Button
								aria-label={buttonAriaLabel}
								variant={buttonVariant}
								className="curved-right-only"
								onClick={() => buttonOnClick?.(value)}
							>
								<FontAwesomeIcon icon={buttonIcon} />
							</Button>
						</ButtonWrapper>
					)}
				</InputGroup>
				{error && (
					<small id={`${name}-error`} className="error-msg" role="alert">
						{error['message'] || 'err'}
					</small>
				)}
			</FormGroup>
		</Col>
	);
};

export default FormTextWithButton;
