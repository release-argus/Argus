import { FieldText } from '@/components/generic/field';

/**
 * Form fields for a `Split` `url_command`.
 *
 * @param name - The name of the field in the form.
 */
const SPLIT = ({ name }: { name: string }) => {
	return (
		<>
			<FieldText
				colSize={{ sm: 6, xs: 5 }}
				key="text"
				label="Text"
				labelSize="sm"
				name={`${name}.text`}
				required
			/>
			<FieldText
				colSize={{ sm: 2, xs: 2 }}
				key="index"
				label="Index"
				labelSize="sm"
				name={`${name}.index`}
				required
			/>
		</>
	);
};

export default SPLIT;
