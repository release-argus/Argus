import { FormLabel, FormText } from 'components/generic/form';

import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { NotifyRocketChatType } from 'types/config';
import { firstNonDefault } from 'utils';
import { useMemo } from 'react';

/**
 * The form fields for a Rocket.Chat notifier.
 *
 * @param name - The path to this `Rocket.Chat` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `Rocket.Chat` notifier.
 */
const ROCKET_CHAT = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifyRocketChatType;
	defaults?: NotifyRocketChatType;
	hard_defaults?: NotifyRocketChatType;
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
				tokena: firstNonDefault(
					main?.url_fields?.tokena,
					defaults?.url_fields?.tokena,
					hard_defaults?.url_fields?.tokena,
				),
				tokenb: firstNonDefault(
					main?.url_fields?.tokenb,
					defaults?.url_fields?.tokenb,
					hard_defaults?.url_fields?.tokenb,
				),
				username: firstNonDefault(
					main?.url_fields?.username,
					defaults?.url_fields?.username,
					hard_defaults?.url_fields?.username,
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
					tooltip="e.g. rocketchat.example.io"
					defaultVal={convertedDefaults.url_fields.host}
				/>
				<FormText
					name={`${name}.url_fields.port`}
					required
					col_sm={3}
					label="Port"
					isNumber
					defaultVal={convertedDefaults.url_fields.port}
					positionXS="right"
				/>
				<FormText
					name={`${name}.url_fields.path`}
					label="Path"
					tooltip={
						<>
							e.g. rocketchat.example.io/{''}
							<span className="bold-underline">path</span>
						</>
					}
					tooltipAriaLabel="Format: rocketchat.example.io/PATH"
					defaultVal={convertedDefaults.url_fields.path}
				/>
				<FormText
					name={`${name}.url_fields.channel`}
					required
					label="Channel"
					defaultVal={convertedDefaults.url_fields.channel}
					positionXS="right"
				/>
				<FormText
					name={`${name}.url_fields.username`}
					col_sm={12}
					label="Username"
					defaultVal={convertedDefaults.url_fields.username}
				/>
				<FormText
					name={`${name}.url_fields.tokena`}
					required
					label="Token A"
					defaultVal={convertedDefaults.url_fields.tokena}
				/>
				<FormText
					name={`${name}.url_fields.tokenb`}
					required
					label="Token B"
					defaultVal={convertedDefaults.url_fields.tokenb}
					positionXS="right"
				/>
			</>
		</>
	);
};

export default ROCKET_CHAT;
