import { FormText } from 'components/generic/form';

/**
 * The form fields for the `Replace` url_command.
 *
 * @param name - The name of the field in the form.
 * @returns The form fields for this Replace url_command.
 */
const REPLACE = ({ name }: { name: string }) => (
	<>
		<FormText
			key="old"
			name={`${name}.old`}
			required
			col_xs={7}
			col_sm={4}
			label="Replace"
			smallLabel
			position="middle"
			positionXS="right"
		/>
		<FormText
			key="new"
			name={`${name}.new`}
			col_sm={4}
			label="With"
			smallLabel
			position="right"
		/>
	</>
);

export default REPLACE;
