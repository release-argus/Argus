import { FormCheck, FormText } from 'components/generic/form';
import { useFormContext, useWatch } from 'react-hook-form';

import { useEffect } from 'react';

/**
 * The form fields for a `RegEx` url_command.
 *
 * @param name - The name of the field in the form.
 * @returns The form fields for this RegEx url_command.
 */
const REGEX = ({ name }: { name: string }) => {
	const { setValue } = useFormContext();

	// Template toggle.
	const templateToggle: boolean | undefined = useWatch({
		name: `${name}.template_toggle`,
	});
	useEffect(() => {
		// Clear the template if toggle false.
		if (templateToggle === false) {
			setValue(`${name}.template`, '');
			setValue(`${name}.template_toggle`, false);
		}
	}, [templateToggle]);

	return (
		<>
			<FormText
				name={`${name}.regex`}
				required
				col_sm={5}
				col_xs={7}
				label="RegEx"
				smallLabel
				isRegex
				position="middle"
				positionXS="right"
			/>
			<FormText
				name={`${name}.index`}
				col_sm={2}
				col_xs={2}
				label="Index"
				smallLabel
				isNumber
				isRegex
				placeholder="0"
				position="middle"
				positionXS="left"
			/>
			<FormCheck
				name={`${name}.template_toggle`}
				col_sm={1}
				col_xs={2}
				size="lg"
				label="T"
				smallLabel
				tooltip="Use the RegEx to create a template"
				position="right"
				positionXS="middle"
			/>
			{templateToggle && (
				<FormText
					name={`${name}.template`}
					col_sm={12}
					col_xs={8}
					label="Template"
					smallLabel
					tooltip="e.g. RegEx of 'v(\d)-(\d)-(\d)' on 'v4-0-1' with template '$1.$2.$3' would give '4.0.1'"
					position="right"
				/>
			)}
		</>
	);
};

export default REGEX;
