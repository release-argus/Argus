import {
	FormLabel,
	FormText,
	FormTextWithPreview,
} from 'components/generic/form';

import { NotifyMatterMostType } from 'types/config';
import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { firstNonDefault } from 'utils';
import { useMemo } from 'react';

/**
 * The form fields for a MatterMost notifier.
 *
 * @param name - The path to this `MatterMost` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `MatterMost` notifier.
 */
const MATTERMOST = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifyMatterMostType;
	defaults?: NotifyMatterMostType;
	hard_defaults?: NotifyMatterMostType;
}) => {
	const convertedDefaults = useMemo(
		() => ({
			// URL Fields
			url_fields: {
				channel: firstNonDefault(
					main?.url_fields?.channel,
					defaults?.url_fields?.channel,
					hard_defaults?.url_fields?.channel,
				),
				host: firstNonDefault(
					main?.url_fields?.host,
					defaults?.url_fields?.host,
					hard_defaults?.url_fields?.host,
				),
				path: firstNonDefault(
					main?.url_fields?.path,
					defaults?.url_fields?.path,
					hard_defaults?.url_fields?.path,
				),
				port: firstNonDefault(
					main?.url_fields?.port,
					defaults?.url_fields?.port,
					hard_defaults?.url_fields?.port,
				),
				token: firstNonDefault(
					main?.url_fields?.token,
					defaults?.url_fields?.token,
					hard_defaults?.url_fields?.token,
				),
				username: firstNonDefault(
					main?.url_fields?.username,
					defaults?.url_fields?.username,
					hard_defaults?.url_fields?.username,
				),
			},
			// Params
			params: {
				icon: firstNonDefault(
					main?.params?.icon,
					defaults?.params?.icon,
					hard_defaults?.params?.icon,
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
					name={`${name}.url_fields.host`}
					required
					col_sm={9}
					label="Host"
					tooltip="e.g. gotify.example.com"
					defaultVal={convertedDefaults.url_fields.host}
				/>
				<FormText
					name={`${name}.url_fields.port`}
					col_sm={3}
					label="Port"
					tooltip="e.g. 443"
					isNumber
					defaultVal={convertedDefaults.url_fields.port}
					position="right"
				/>
				<FormText
					name={`${name}.url_fields.path`}
					label="Path"
					tooltip={
						<>
							{'e.g. mattermost.example.io/'}
							<span className="bold-underline">path</span>
						</>
					}
					tooltipAriaLabel="Format: mattermost.example.io/PATH"
					defaultVal={convertedDefaults.url_fields.path}
				/>
				<FormText
					name={`${name}.url_fields.channel`}
					label="Channel"
					tooltip="e.g. releases"
					defaultVal={convertedDefaults.url_fields.channel}
					position="right"
				/>
				<FormText
					name={`${name}.url_fields.username`}
					label="Username"
					defaultVal={convertedDefaults.url_fields.username}
				/>
				<FormText
					name={`${name}.url_fields.token`}
					required
					label="Token"
					tooltip="WebHook token"
					defaultVal={convertedDefaults.url_fields.token}
					position="right"
				/>
			</>
			<FormLabel text="Params" heading />
			<>
				<FormTextWithPreview
					name={`${name}.params.icon`}
					label="Icon"
					tooltip="URL of icon to use"
					isURL={false}
					defaultVal={convertedDefaults.params.icon}
				/>
			</>
		</>
	);
};

export default MATTERMOST;
