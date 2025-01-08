import {
	FormLabel,
	FormText,
	FormTextWithPreview,
} from 'components/generic/form';

import { NotifyJoinType } from 'types/config';
import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { firstNonDefault } from 'utils';
import { useMemo } from 'react';

/**
 * The form fields for a Join notifier.
 *
 * @param name - The path to this `Join` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `Join` notifier.
 */
const JOIN = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifyJoinType;
	defaults?: NotifyJoinType;
	hard_defaults?: NotifyJoinType;
}) => {
	const convertedDefaults = useMemo(
		() => ({
			// URL Fields
			url_fields: {
				apikey: firstNonDefault(
					main?.url_fields?.apikey,
					defaults?.url_fields?.apikey,
					hard_defaults?.url_fields?.apikey,
				),
			},
			// Params
			params: {
				devices: firstNonDefault(
					main?.params?.devices,
					defaults?.params?.devices,
					hard_defaults?.params?.devices,
				),
				icon: firstNonDefault(
					main?.params?.icon,
					defaults?.params?.icon,
					hard_defaults?.params?.icon,
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
					name={`${name}.url_fields.apikey`}
					required
					col_sm={12}
					label="API Key"
					defaultVal={convertedDefaults.url_fields.apikey}
				/>
			</>
			<FormLabel text="Params" heading />
			<>
				<FormText
					name={`${name}.params.devices`}
					required
					col_sm={12}
					label="Devices"
					tooltip="e.g. ID1,ID2..."
					defaultVal={convertedDefaults.params.devices}
				/>
				<FormTextWithPreview
					name={`${name}.params.icon`}
					label="Icon"
					tooltip="URL of icon to use"
					defaultVal={convertedDefaults.params.icon}
				/>
				<FormText
					name={`${name}.params.title`}
					col_sm={12}
					label="Title"
					tooltip="e.g. 'Release - {{ service_name | default:service_id }}'"
					defaultVal={convertedDefaults.params.title}
				/>
			</>
		</>
	);
};

export default JOIN;
