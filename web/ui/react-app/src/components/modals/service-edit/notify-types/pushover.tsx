import { FormLabel, FormText } from 'components/generic/form';

import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { NotifyPushoverType } from 'types/config';
import { firstNonDefault } from 'utils';
import { useMemo } from 'react';

/**
 * The form fields for a PushOver notifier.
 *
 * @param name - The path to this `PushOver` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `PushOver` notifier.
 */
const PUSHOVER = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifyPushoverType;
	defaults?: NotifyPushoverType;
	hard_defaults?: NotifyPushoverType;
}) => {
	const convertedDefaults = useMemo(
		() => ({
			// URL Fields
			url_fields: {
				token: firstNonDefault(
					main?.url_fields?.token,
					defaults?.url_fields?.token,
					hard_defaults?.url_fields?.token,
				),
				user: firstNonDefault(
					main?.url_fields?.user,
					defaults?.url_fields?.user,
					hard_defaults?.url_fields?.user,
				),
			},
			// Params
			params: {
				devices: firstNonDefault(
					main?.params?.devices,
					defaults?.params?.devices,
					hard_defaults?.params?.devices,
				),
				priority: firstNonDefault(
					main?.params?.priority,
					defaults?.params?.priority,
					hard_defaults?.params?.priority,
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
					name={`${name}.url_fields.token`}
					required
					col_sm={6}
					label="API Token/Key"
					tooltip="'Create an Application/API Token' on the Pushover dashboard'"
					defaultVal={convertedDefaults.url_fields.token}
				/>
				<FormText
					name={`${name}.url_fields.user`}
					required
					col_sm={6}
					label="User Key"
					tooltip="Top right of Pushover dashboard"
					defaultVal={convertedDefaults.url_fields.user}
					position="right"
				/>
			</>
			<FormLabel text="Params" heading />
			<>
				<FormText
					name={`${name}.params.devices`}
					col_sm={12}
					label="Devices"
					tooltip="e.g. device1,device2... (deviceX=Name column in the 'Your Devices' list)"
					defaultVal={convertedDefaults.params.devices}
				/>
				<FormText
					name={`${name}.params.title`}
					col_sm={9}
					label="Title"
					defaultVal={convertedDefaults.params.title}
				/>
				<FormText
					name={`${name}.params.priority`}
					col_sm={3}
					label="Priority"
					tooltip="Only supply priority values between -1 and 1, since 2 requires additional parameters that are not supported yet"
					isNumber
					defaultVal={convertedDefaults.params.priority}
					position="right"
				/>
			</>
		</>
	);
};

export default PUSHOVER;
