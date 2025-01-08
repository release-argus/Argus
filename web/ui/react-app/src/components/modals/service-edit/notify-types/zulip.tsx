import { FormLabel, FormText } from 'components/generic/form';

import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { NotifyZulipType } from 'types/config';
import { firstNonDefault } from 'utils';
import { useMemo } from 'react';

/**
 * The form fields for a Zulip Chat notifier.
 *
 * @param name - The path to this `Zulip Chat` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `Zulip Chat` notifier.
 */
const ZULIP_CHAT = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifyZulipType;
	defaults?: NotifyZulipType;
	hard_defaults?: NotifyZulipType;
}) => {
	const convertedDefaults = useMemo(
		() => ({
			// URL Fields
			url_fields: {
				botkey: firstNonDefault(
					main?.url_fields?.botkey,
					defaults?.url_fields?.botkey,
					hard_defaults?.url_fields?.botkey,
				),
				botmail: firstNonDefault(
					main?.url_fields?.botmail,
					defaults?.url_fields?.botmail,
					hard_defaults?.url_fields?.botmail,
				),
				host: firstNonDefault(
					main?.url_fields?.host,
					defaults?.url_fields?.host,
					hard_defaults?.url_fields?.host,
				),
			},
			// Params
			params: {
				stream: firstNonDefault(
					main?.params?.stream,
					defaults?.params?.stream,
					hard_defaults?.params?.stream,
				),
				topic: firstNonDefault(
					main?.params?.topic,
					defaults?.params?.topic,
					hard_defaults?.params?.topic,
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
					name={`${name}.url_fields.botmail`}
					required
					label="Bot Mail"
					tooltip="e.g. something@example.com"
					defaultVal={convertedDefaults.url_fields.botmail}
				/>
				<FormText
					name={`${name}.url_fields.botkey`}
					required
					label="Bot Key"
					defaultVal={convertedDefaults.url_fields.botkey}
					position="right"
				/>
				<FormText
					name={`${name}.url_fields.host`}
					required
					col_sm={12}
					label="Host"
					tooltip="e.g. zulip.example.com"
					defaultVal={convertedDefaults.url_fields.host}
				/>
			</>
			<FormLabel text="Params" heading />
			<>
				<FormText
					name={`${name}.params.stream`}
					label="Stream"
					defaultVal={convertedDefaults.params.stream}
				/>
				<FormText
					name={`${name}.params.topic`}
					label="Topic"
					defaultVal={convertedDefaults.params.topic}
					position="right"
				/>
			</>
		</>
	);
};

export default ZULIP_CHAT;
