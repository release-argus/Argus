import { FormLabel, FormText } from 'components/generic/form';

import { NotifyIFTTTType } from 'types/config';
import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { firstNonDefault } from 'utils';
import { useMemo } from 'react';

/**
 * The form fields for an IFTTT notifier.
 *
 * @param name - The path to this `IFTTT` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `IFTTT` notifier.
 */
const IFTTT = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifyIFTTTType;
	defaults?: NotifyIFTTTType;
	hard_defaults?: NotifyIFTTTType;
}) => {
	const convertedDefaults = useMemo(
		() => ({
			// URL Fields
			url_fields: {
				webhookid: firstNonDefault(
					main?.url_fields?.webhookid,
					defaults?.url_fields?.webhookid,
					hard_defaults?.url_fields?.webhookid,
				),
			},
			// Params
			params: {
				events: firstNonDefault(
					main?.params?.events,
					defaults?.params?.events,
					hard_defaults?.params?.events,
				),
				title: firstNonDefault(
					main?.params?.title,
					defaults?.params?.title,
					hard_defaults?.params?.title,
				),
				usemessageasvalue: firstNonDefault(
					main?.params?.usemessageasvalue,
					defaults?.params?.usemessageasvalue,
					hard_defaults?.params?.usemessageasvalue,
				),
				usetitleasvalue: firstNonDefault(
					main?.params?.usetitleasvalue,
					defaults?.params?.usetitleasvalue,
					hard_defaults?.params?.usetitleasvalue,
				),
				value1: firstNonDefault(
					main?.params?.value1,
					defaults?.params?.value1,
					hard_defaults?.params?.value1,
				),
				value2: firstNonDefault(
					main?.params?.value2,
					defaults?.params?.value2,
					hard_defaults?.params?.value2,
				),
				value3: firstNonDefault(
					main?.params?.value3,
					defaults?.params?.value3,
					hard_defaults?.params?.value3,
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
					name={`${name}.url_fields.webhookid`}
					required
					col_sm={12}
					label="WebHook ID"
					defaultVal={convertedDefaults.url_fields.webhookid}
				/>
			</>
			<FormLabel text="Params" heading />
			<>
				<FormText
					name={`${name}.params.events`}
					required
					col_sm={12}
					label="Events"
					tooltip="e.g. event1,event2..."
					defaultVal={convertedDefaults.params.events}
				/>
				<FormText
					name={`${name}.params.title`}
					col_sm={12}
					label="Title"
					tooltip="Optional notification title"
					defaultVal={convertedDefaults.params.title}
				/>
				<FormText
					name={`${name}.params.usemessageasvalue`}
					label="Use Message As Value"
					tooltip="Set the corresponding value field to the message"
					isNumber
					defaultVal={convertedDefaults.params.usemessageasvalue}
				/>
				<FormText
					name={`${name}.params.usetitleasvalue`}
					label="Use Title As Value"
					tooltip="Set the corresponding value field to the title"
					isNumber
					defaultVal={convertedDefaults.params.usetitleasvalue}
					positionXS="right"
				/>
				<FormText
					name={`${name}.params.value1`}
					col_sm={4}
					label="Value1"
					defaultVal={convertedDefaults.params.value1}
				/>
				<FormText
					name={`${name}.params.value2`}
					col_sm={4}
					label="Value2"
					defaultVal={convertedDefaults.params.value2}
					positionXS="middle"
				/>
				<FormText
					name={`${name}.params.value3`}
					col_sm={4}
					label="Value3"
					defaultVal={convertedDefaults.params.value3}
					positionXS="right"
				/>
			</>
		</>
	);
};

export default IFTTT;
