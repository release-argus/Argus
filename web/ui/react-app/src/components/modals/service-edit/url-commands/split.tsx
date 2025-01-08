import { FormText } from 'components/generic/form';

/**
 * The form fields for a `Split` url_command.
 *
 * @param name - The name of the field in the form.
 * @returns The form fields for this Split url_command.
 */
const SPLIT = ({ name }: { name: string }) => (
	<>
		<FormText
			key="text"
			name={`${name}.text`}
			required
			col_xs={5}
			col_sm={6}
			label="Text"
			smallLabel
			position="middle"
		/>
		<FormText
			key="index"
			name={`${name}.index`}
			required
			col_xs={2}
			col_sm={2}
			label="Index"
			smallLabel
			isNumber
			position="right"
		/>
	</>
);

export default SPLIT;
