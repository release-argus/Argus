import { FormColour, FormLabel, FormText } from 'components/generic/form';

import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { NotifyTeamsType } from 'types/config';
import { firstNonDefault } from 'utils';
import { useMemo } from 'react';

/**
 * The form fields for a Teams notifier.
 *
 * @param name - The path to this `Teams` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `Teams` notifier.
 */
const TEAMS = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifyTeamsType;
	defaults?: NotifyTeamsType;
	hard_defaults?: NotifyTeamsType;
}) => {
	const convertedDefaults = useMemo(
		() => ({
			// URL Fields
			url_fields: {
				altid: firstNonDefault(
					main?.url_fields?.altid,
					defaults?.url_fields?.altid,
					hard_defaults?.url_fields?.altid,
				),
				group: firstNonDefault(
					main?.url_fields?.group,
					defaults?.url_fields?.group,
					hard_defaults?.url_fields?.group,
				),
				groupowner: firstNonDefault(
					main?.url_fields?.groupowner,
					defaults?.url_fields?.groupowner,
					hard_defaults?.url_fields?.groupowner,
				),
				tenant: firstNonDefault(
					main?.url_fields?.tenant,
					defaults?.url_fields?.tenant,
					hard_defaults?.url_fields?.tenant,
				),
			},
			// Params
			params: {
				color: firstNonDefault(
					main?.params?.color,
					defaults?.params?.color,
					hard_defaults?.params?.color,
				),
				host: firstNonDefault(
					main?.params?.host,
					defaults?.params?.host,
					hard_defaults?.params?.host,
				),
				title: firstNonDefault(
					main?.params?.title,
					defaults?.params?.title,
					hard_defaults?.params?.title,
				),
			},
		}),
		[main, defaults, hard_defaults],
	);

	return (
		<>
			<NotifyOptions
				name={name}
				main={main?.options}
				defaults={defaults?.options}
				hard_defaults={hard_defaults?.options}
			/>
			<FormLabel text="URL Fields" heading />
			<>
				<FormText
					name={`${name}.url_fields.altid`}
					label="Alt ID"
					defaultVal={convertedDefaults.url_fields.altid}
				/>
				<FormText
					name={`${name}.url_fields.tenant`}
					label="Tenant"
					defaultVal={convertedDefaults.url_fields.tenant}
					position="right"
				/>
				<FormText
					name={`${name}.url_fields.group`}
					label="Group"
					defaultVal={convertedDefaults.url_fields.group}
				/>
				<FormText
					name={`${name}.url_fields.groupowner`}
					label="Group Owner"
					defaultVal={convertedDefaults.url_fields.groupowner}
					position="right"
				/>
			</>
			<FormLabel text="Params" heading />
			<>
				<FormText
					name={`${name}.params.host`}
					col_sm={6}
					label="Host"
					defaultVal={convertedDefaults.params.host}
				/>
				<FormText
					name={`${name}.params.title`}
					col_sm={6}
					label="Title"
					defaultVal={convertedDefaults.params.title}
					position="right"
				/>
				<FormColour
					name={`${name}.params.color`}
					col_sm={6}
					label="Color"
					defaultVal={convertedDefaults.params.color}
				/>
			</>
		</>
	);
};

export default TEAMS;
