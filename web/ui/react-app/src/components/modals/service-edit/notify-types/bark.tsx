import {
	FormLabel,
	FormSelect,
	FormText,
	FormTextWithPreview,
} from 'components/generic/form';
import { useEffect, useMemo } from 'react';

import { NotifyBarkType } from 'types/config';
import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { firstNonDefault } from 'utils';
import { normaliseForSelect } from 'components/modals/service-edit/util';
import { useFormContext } from 'react-hook-form';

export const BarkSchemeOptions = [
	{ label: 'HTTPS', value: 'https' },
	{ label: 'HTTP', value: 'http' },
];

export const BarkSoundOptions = [
	{ label: '', value: '' },
	{ label: 'Alarm', value: 'alarm' },
	{ label: 'Anticipate', value: 'anticipate' },
	{ label: 'Bell', value: 'bell' },
	{ label: 'Birdsong', value: 'birdsong' },
	{ label: 'Bloom', value: 'bloom' },
	{ label: 'Calypso', value: 'calypso' },
	{ label: 'Chime', value: 'chime' },
	{ label: 'Choo', value: 'choo' },
	{ label: 'Descent', value: 'descent' },
	{ label: 'Electronic', value: 'electronic' },
	{ label: 'Fanfare', value: 'fanfare' },
	{ label: 'Glass', value: 'glass' },
	{ label: 'GoToSleep', value: 'gotosleep' },
	{ label: 'HealthNotification', value: 'healthnotification' },
	{ label: 'Horn', value: 'horn' },
	{ label: 'Ladder', value: 'ladder' },
	{ label: 'MailSent', value: 'mailsent' },
	{ label: 'Minuet', value: 'minuet' },
	{ label: 'MultiWayInvitation', value: 'multiwayinvitation' },
	{ label: 'NewMail', value: 'newmail' },
	{ label: 'NewsFlash', value: 'newsflash' },
	{ label: 'Noir', value: 'noir' },
	{ label: 'PaymentSuccess', value: 'paymentsuccess' },
	{ label: 'Shake', value: 'shake' },
	{ label: 'SherwoodForest', value: 'sherwoodforest' },
	{ label: 'Silence', value: 'silence' },
	{ label: 'Spell', value: 'spell' },
	{ label: 'Suspense', value: 'suspense' },
	{ label: 'Telegraph', value: 'telegraph' },
	{ label: 'Tiptoes', value: 'tiptoes' },
	{ label: 'Typewriters', value: 'typewriters' },
	{ label: 'Update', value: 'update' },
];

/**
 * The form fields for a Bark notifier.
 *
 * @param name - The path to this `Bark` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `Bark` notifier.
 */
const BARK = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifyBarkType;
	defaults?: NotifyBarkType;
	hard_defaults?: NotifyBarkType;
}) => {
	const { getValues, setValue } = useFormContext();
	const convertedDefaults = useMemo(
		() => ({
			// URL Fields
			url_fields: {
				devicekey: firstNonDefault(
					main?.url_fields?.devicekey,
					defaults?.url_fields?.devicekey,
					hard_defaults?.url_fields?.devicekey,
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
			},
			// Params
			params: {
				badge: firstNonDefault(
					main?.params?.badge,
					defaults?.params?.badge,
					hard_defaults?.params?.badge,
				),
				copy: firstNonDefault(
					main?.params?.copy,
					defaults?.params?.copy,
					hard_defaults?.params?.copy,
				),
				group: firstNonDefault(
					main?.params?.group,
					defaults?.params?.group,
					hard_defaults?.params?.group,
				),
				icon: firstNonDefault(
					main?.params?.icon,
					defaults?.params?.icon,
					hard_defaults?.params?.icon,
				),
				scheme: firstNonDefault(
					main?.params?.scheme,
					defaults?.params?.scheme,
					hard_defaults?.params?.scheme,
				).toLowerCase(),
				sound: firstNonDefault(
					main?.params?.sound,
					defaults?.params?.sound,
					hard_defaults?.params?.sound,
				).toLowerCase(),
				title: firstNonDefault(
					main?.params?.title,
					defaults?.params?.title,
					hard_defaults?.params?.title,
				),
				url: firstNonDefault(
					main?.params?.url,
					defaults?.params?.url,
					hard_defaults?.params?.url,
				),
			},
		}),
		[main, defaults, hard_defaults],
	);

	const barkSchemeOptions = useMemo(() => {
		const defaultScheme = normaliseForSelect(
			BarkSchemeOptions,
			convertedDefaults.params.scheme,
		);

		if (defaultScheme)
			return [
				{ value: '', label: `${defaultScheme.label} (default)` },
				...BarkSchemeOptions,
			];

		return BarkSchemeOptions;
	}, [convertedDefaults.params.scheme]);

	const barkSoundOptions = useMemo(() => {
		const defaultSound = normaliseForSelect(
			BarkSoundOptions,
			convertedDefaults.params.sound,
		);

		if (defaultSound)
			return [
				{ value: '', label: `${defaultSound.label} (default)` },
				...BarkSoundOptions.filter((option) => option.value !== ''),
			];

		return BarkSoundOptions;
	}, [convertedDefaults.params.sound]);

	useEffect(() => {
		// Normalise selected scheme, or default it.
		if (convertedDefaults.params.scheme === '')
			setValue(
				`${name}.params.scheme`,
				normaliseForSelect(
					BarkSchemeOptions,
					getValues(`${name}.params.scheme`),
				)?.value || 'https',
			);

		// Normalise selected sound, or default it.
		if (
			convertedDefaults.params.sound === '' &&
			getValues(`${name}.params.sound`) !== undefined
		)
			setValue(
				`${name}.params.sound`,
				normaliseForSelect(BarkSoundOptions, getValues(`${name}.params.sound`))
					?.value ?? '',
			);
	}, []);

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
					col_sm={9}
					required
					label="Host"
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
					tooltip="Server path"
					defaultVal={convertedDefaults.url_fields.path}
				/>
				<FormText
					name={`${name}.url_fields.devicekey`}
					required
					label="Device Key"
					defaultVal={convertedDefaults.url_fields.devicekey}
					positionXS="right"
				/>
			</>
			<FormLabel text="Params" heading />
			<>
				<FormSelect
					name={`${name}.params.scheme`}
					col_sm={3}
					label="Scheme"
					tooltip="Server protocol"
					options={barkSchemeOptions}
				/>
				<FormText
					name={`${name}.params.badge`}
					col_sm={3}
					label="Badge"
					tooltip="The number displayed next to the App icon"
					isNumber
					defaultVal={convertedDefaults.params.badge}
					positionXS="middle"
				/>
				<FormText
					name={`${name}.params.copy`}
					label="Copy"
					tooltip="The value to be copied"
					defaultVal={convertedDefaults.params.copy}
					positionXS="right"
				/>
				<FormText
					name={`${name}.params.group`}
					label="Group"
					tooltip="The group of the notification"
					defaultVal={convertedDefaults.params.group}
				/>
				<FormSelect
					name={`${name}.params.sound`}
					label="Sound"
					options={barkSoundOptions}
					positionXS="right"
				/>
				<FormText
					name={`${name}.params.title`}
					label="Title"
					defaultVal={convertedDefaults.params.title}
				/>
				<FormText
					name={`${name}.params.url`}
					label="URL"
					tooltip="URL to open when notification is tapped"
					defaultVal={convertedDefaults.params.url}
					positionXS="right"
				/>
				<FormTextWithPreview
					name={`${name}.params.icon`}
					label="Icon"
					tooltip="URL to an icon"
					defaultVal={convertedDefaults.params.icon}
				/>
			</>
		</>
	);
};

export default BARK;
