import { FormLabel, FormText } from 'components/generic/form';

import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { NotifyPushbulletType } from 'types/config';
import { firstNonDefault } from 'utils';
import { useMemo } from 'react';

/**
 * The form fields for a PushBullet notifier.
 *
 * @param name - The path to this `PushBullet` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `PushBullet` notifier.
 */
const PUSHBULLET = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifyPushbulletType;
	defaults?: NotifyPushbulletType;
	hard_defaults?: NotifyPushbulletType;
}) => {
	const convertedDefaults = useMemo(
		() => ({
			// URL Fields
			url_fields: {
				targets: firstNonDefault(
					main?.url_fields?.targets,
					defaults?.url_fields?.targets,
					hard_defaults?.url_fields?.targets,
				),
				token: firstNonDefault(
					main?.url_fields?.token,
					defaults?.url_fields?.token,
					hard_defaults?.url_fields?.token,
				),
			},
			// Params
			params: {
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
					col_sm={12}
					label="Access Token"
					defaultVal={convertedDefaults.url_fields.token}
				/>
				<FormText
					name={`${name}.url_fields.targets`}
					required
					col_sm={12}
					label="Targets"
					tooltip="e.g. DEVICE1,DEVICE2..."
					defaultVal={convertedDefaults.url_fields.targets}
				/>
			</>
			<FormLabel text="Params" heading />
			<>
				<FormText
					name={`${name}.params.title`}
					col_sm={12}
					label="Title"
					defaultVal={convertedDefaults.params.title}
				/>
			</>
		</>
	);
};

export default PUSHBULLET;
