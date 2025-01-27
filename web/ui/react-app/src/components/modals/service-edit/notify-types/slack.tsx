import {
	FormColour,
	FormLabel,
	FormText,
	FormTextWithPreview,
} from 'components/generic/form';

import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { NotifySlackType } from 'types/config';
import { firstNonDefault } from 'utils';
import { useMemo } from 'react';

/**
 * The form fields for a Slack notifier.
 *
 * @param name - The path to this `Slack` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `Slack` notifier.
 */
const SLACK = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifySlackType;
	defaults?: NotifySlackType;
	hard_defaults?: NotifySlackType;
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
				token: firstNonDefault(
					main?.url_fields?.token,
					defaults?.url_fields?.token,
					hard_defaults?.url_fields?.token,
				),
			},
			// Params
			params: {
				botname: firstNonDefault(
					main?.params?.botname,
					defaults?.params?.botname,
					hard_defaults?.params?.botname,
				),
				color: firstNonDefault(
					main?.params?.color,
					defaults?.params?.color,
					hard_defaults?.params?.color,
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
					name={`${name}.url_fields.token`}
					required
					label="Token"
					tooltip={
						<>
							{'xoxb:'}
							<span className="bold-underline">BOT-OAUTH-TOKEN</span>
							{' or '}
							<span className="bold-underline">WEBHOOK</span>
						</>
					}
					tooltipAriaLabel="Format: xoxb:BOT-OAUTH-TOKEN or WEBHOOK"
					defaultVal={convertedDefaults.url_fields.token}
				/>
				<FormText
					name={`${name}.url_fields.channel`}
					required
					label="Channel"
					defaultVal={convertedDefaults.url_fields.channel}
					position="right"
				/>
			</>
			<FormLabel text="Params" heading />
			<>
				<FormText
					name={`${name}.params.botname`}
					label="Bot Name"
					defaultVal={convertedDefaults.params.botname}
				/>
				<FormColour
					name={`${name}.params.color`}
					label="Color"
					tooltip="Message left-hand border color in hex, e.g. #ffffff"
					defaultVal={convertedDefaults.params.color}
					position="right"
				/>
				<FormTextWithPreview
					name={`${name}.params.icon`}
					label="Icon"
					tooltip="Use emoji or URL as icon (based on presence of http(s):// prefix)"
					isURL={false}
					defaultVal={convertedDefaults.params.icon}
				/>
				<FormText
					name={`${name}.params.title`}
					col_sm={12}
					type="text"
					label="Title"
					tooltip="Text prepended to the message"
					defaultVal={convertedDefaults.params.title}
				/>
			</>
		</>
	);
};

export default SLACK;
