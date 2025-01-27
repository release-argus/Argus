import { FC } from 'react';
import { Form } from 'react-bootstrap';
import { HelpTooltip } from 'components/generic';
import { TooltipWithAriaProps } from './tooltip';

type Props = {
	id?: string;
	htmlFor?: string;
	text: string;
	heading?: boolean;
	required?: boolean;
	small?: boolean;
};

type FormLabelProps = TooltipWithAriaProps & Props;

/**
 * A label for a form item.
 *
 * @param id - The ID of this label.
 * @param htmlFor - The ID of the item this label is for.
 * @param text - The text of the label.
 * @param tooltip - The tooltip of the label.
 * @param tooltipAriaLabel - The aria label for the tooltip (Defaults to the tooltip).
 * @param heading - Whether the label is a heading.
 * @param required - Whether the label is required.
 * @param small - Whether the label is small.
 * @returns A label for a form item.
 */
const FormLabel: FC<FormLabelProps> = ({
	id,
	htmlFor,
	text,
	tooltip,
	tooltipAriaLabel,
	heading,
	required,
	small,
}) => {
	const style = () => {
		if (heading)
			return {
				fontSize: '1.25rem',
				textDecorationLine: 'underline',
				paddingTop: '1.5rem',
			};
		if (small)
			return {
				fontSize: '0.8rem',
			};
		return undefined;
	};

	return (
		<Form.Label id={id} htmlFor={htmlFor} style={style()}>
			{text}
			{required && <span className="text-danger">*</span>}
			{tooltip && (
				<HelpTooltip
					id={`${htmlFor}-tooltip`}
					tooltip={tooltip}
					tooltipAriaLabel={tooltipAriaLabel ?? tooltip}
				/>
			)}
		</Form.Label>
	);
};

export default FormLabel;
