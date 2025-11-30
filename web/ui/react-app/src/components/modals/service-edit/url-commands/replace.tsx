import { FieldText } from '@/components/generic/field';

/**
 * Form fields for a `Replace` `url_command`.
 *
 * @param name - The name of the field in the form.
 */
const REPLACE = ({ name }: { name: string }) => (
	<>
		<FieldText
			colSize={{ sm: 4, xs: 7 }}
			key="old"
			label="Replace"
			labelSize="sm"
			name={`${name}.old`}
			required
		/>
		<FieldText
			colSize={{ sm: 4 }}
			key="new"
			label="With"
			labelSize="sm"
			name={`${name}.new`}
		/>
	</>
);

export default REPLACE;
