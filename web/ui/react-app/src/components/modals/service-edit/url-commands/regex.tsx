import { useEffect } from 'react';
import { useFormContext, useWatch } from 'react-hook-form';
import { FieldCheck, FieldText } from '@/components/generic/field';

/**
 * Form fields for a `RegEx` `url_command`.
 *
 * @param name - The name of the field in the form.
 */
const REGEX = ({ name }: { name: string }) => {
	const { setValue } = useFormContext();

	// Template toggle.
	const templateToggle = useWatch({
		name: `${name}.template_toggle`,
	}) as boolean | undefined;
	// biome-ignore lint/correctness/useExhaustiveDependencies: name stable.
	useEffect(() => {
		// Clear the template if toggle false.
		if (templateToggle === false) {
			setValue(`${name}.template`, '');
			setValue(`${name}.template_toggle`, false);
		}
	}, [templateToggle]);

	return (
		<>
			<FieldText
				colSize={{ sm: 5, xs: 7 }}
				label="RegEx"
				labelSize="sm"
				name={`${name}.regex`}
				required
			/>
			<FieldText
				className="order-2 sm:order-1"
				colSize={{ sm: 2, xs: 3 }}
				label="Index"
				labelSize="sm"
				name={`${name}.index`}
				placeholder="0"
				tooltip={{
					content: 'Index of the RegEx match to use (starting at 0).',
					type: 'string',
				}}
			/>
			<FieldCheck
				colSize={{ sm: 1, xs: 2 }}
				label="T"
				labelSize="sm"
				name={`${name}.template_toggle`}
				tooltip={{
					content: 'Use the RegEx to create a template',
					type: 'string',
				}}
			/>
			{templateToggle && (
				<FieldText
					className="order-1 sm:order-2"
					colSize={{ sm: 12, xs: 6 }}
					label="RegEx Template"
					labelSize="sm"
					name={`${name}.template`}
					tooltip={{
						content: String.raw`e.g. RegEx of 'v(\d)-(\d)-(\d)' on 'v4-0-1' with template '$1.$2.$3' would give '4.0.1'`,
						type: 'string',
					}}
				/>
			)}
		</>
	);
};

export default REGEX;
