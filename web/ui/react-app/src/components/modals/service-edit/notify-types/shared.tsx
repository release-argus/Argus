import { FormLabel, FormText, FormTextArea } from 'components/generic/form';
import { memo, useMemo } from 'react';

import { NotifyOptionsType } from 'types/config';
import { firstNonDefault } from 'utils';
import { numberRangeTest } from 'components/generic/form-validate';

/**
 * The form fields for the `notify.X.options` section.
 *
 * @param name - The path to these `options` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for the `options` section of this `Notify`.
 */
export const NotifyOptions = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifyOptionsType;
	defaults?: NotifyOptionsType;
	hard_defaults?: NotifyOptionsType;
}) => {
	const convertedDefaults = useMemo(
		() => ({
			// Options
			delay: firstNonDefault(
				main?.delay,
				defaults?.delay,
				hard_defaults?.delay,
			),
			max_tries: firstNonDefault(
				main?.max_tries,
				defaults?.max_tries,
				hard_defaults?.max_tries,
			),
			message: firstNonDefault(
				main?.message,
				defaults?.message,
				hard_defaults?.message,
			),
		}),
		[main, defaults, hard_defaults],
	);

	return (
		<>
			<FormLabel text="Options" heading />
			<>
				<FormText
					name={`${name}.options.delay`}
					col_xs={6}
					label="Delay"
					tooltip="e.g. 1h2m3s = 1 hour, 2 minutes and 3 seconds"
					defaultVal={convertedDefaults.delay}
				/>
				<FormText
					name={`${name}.options.max_tries`}
					col_xs={6}
					label="Max tries"
					isNumber
					validationFunc={(value: string) => numberRangeTest(value, 0, 255)}
					defaultVal={convertedDefaults.max_tries}
					positionXS="right"
				/>
				<FormTextArea
					name={`${name}.options.message`}
					col_sm={12}
					rows={3}
					label="Message"
					defaultVal={convertedDefaults.message}
				/>
			</>
		</>
	);
};

export default memo(NotifyOptions);
