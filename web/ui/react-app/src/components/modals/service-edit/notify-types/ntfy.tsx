import {
	FormLabel,
	FormSelect,
	FormText,
	FormTextWithPreview,
} from 'components/generic/form';
import {
	convertNtfyActionsFromString,
	normaliseForSelect,
} from 'components/modals/service-edit/util';
import { firstNonDefault, strToBool } from 'utils';
import { useEffect, useMemo } from 'react';

import { BooleanWithDefault } from 'components/generic';
import { NotifyNtfyType } from 'types/config';
import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { NtfyActions } from 'components/modals/service-edit/notify-types/extra';
import { useFormContext } from 'react-hook-form';

export const NtfySchemeOptions = [
	{ label: 'HTTPS', value: 'https' },
	{ label: 'HTTP', value: 'http' },
];

export const NtfyPriorityOptions = [
	{ label: 'Min', value: 'min' },
	{ label: 'Low', value: 'low' },
	{ label: 'Default', value: 'default' },
	{ label: 'High', value: 'high' },
	{ label: 'Max', value: 'max' },
];

/**
 * The form fields for a NTFY notifier.
 *
 * @param name - The path to this `NTFY` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `NTFY` notifier.
 */
const NTFY = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifyNtfyType;
	defaults?: NotifyNtfyType;
	hard_defaults?: NotifyNtfyType;
}) => {
	const { getValues, setValue } = useFormContext();

	const convertedDefaults = useMemo(
		() => ({
			// URL Fields
			url_fields: {
				host: firstNonDefault(
					main?.url_fields?.host,
					defaults?.url_fields?.host,
					hard_defaults?.url_fields?.host,
				),
				password: firstNonDefault(
					main?.url_fields?.password,
					defaults?.url_fields?.password,
					hard_defaults?.url_fields?.password,
				),
				port: firstNonDefault(
					main?.url_fields?.port,
					defaults?.url_fields?.port,
					hard_defaults?.url_fields?.port,
				),
				topic: firstNonDefault(
					main?.url_fields?.topic,
					defaults?.url_fields?.topic,
					hard_defaults?.url_fields?.topic,
				),
				username: firstNonDefault(
					main?.url_fields?.username,
					defaults?.url_fields?.username,
					hard_defaults?.url_fields?.username,
				),
			},
			// Params
			params: {
				actions: convertNtfyActionsFromString(
					firstNonDefault(
						main?.params?.actions as string | undefined,
						defaults?.params?.actions as string | undefined,
						hard_defaults?.params?.actions as string | undefined,
					),
				),
				attach: firstNonDefault(
					main?.params?.attach,
					defaults?.params?.attach,
					hard_defaults?.params?.attach,
				),
				cache:
					strToBool(
						firstNonDefault(
							main?.params?.cache,
							defaults?.params?.cache,
							hard_defaults?.params?.cache,
						),
					) ?? true,
				click: firstNonDefault(
					main?.params?.click,
					defaults?.params?.click,
					hard_defaults?.params?.click,
				),
				email: firstNonDefault(
					main?.params?.email,
					defaults?.params?.email,
					hard_defaults?.params?.email,
				),
				filename: firstNonDefault(
					main?.params?.filename,
					defaults?.params?.filename,
					hard_defaults?.params?.filename,
				),
				firebase:
					strToBool(
						firstNonDefault(
							main?.params?.firebase,
							defaults?.params?.firebase,
							hard_defaults?.params?.firebase,
						),
					) ?? true,
				icon: firstNonDefault(
					main?.params?.icon,
					defaults?.params?.icon,
					hard_defaults?.params?.icon,
				),
				priority: firstNonDefault(
					main?.params?.priority,
					defaults?.params?.priority,
					hard_defaults?.params?.priority,
				).toLowerCase(),
				scheme: firstNonDefault(
					main?.params?.scheme,
					defaults?.params?.scheme,
					hard_defaults?.params?.scheme,
				).toLowerCase(),
				tags: firstNonDefault(
					main?.params?.tags,
					defaults?.params?.tags,
					hard_defaults?.params?.tags,
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

	const ntfyPriorityOptions = useMemo(() => {
		const defaultPriority = normaliseForSelect(
			NtfyPriorityOptions,
			convertedDefaults.params.priority,
		);

		if (defaultPriority)
			return [
				{ value: '', label: `${defaultPriority.label} (default)` },
				...NtfyPriorityOptions,
			];

		return NtfyPriorityOptions;
	}, [convertedDefaults.params.priority]);

	const ntfySchemeOptions = useMemo(() => {
		const defaultScheme = normaliseForSelect(
			NtfySchemeOptions,
			convertedDefaults.params.scheme,
		);

		if (defaultScheme)
			return [
				{ value: '', label: `${defaultScheme.label} (default)` },
				...NtfySchemeOptions,
			];

		return NtfySchemeOptions;
	}, [convertedDefaults.params.scheme]);

	useEffect(() => {
		// Normalise selected priority, or default it.
		if (convertedDefaults.params.priority === '')
			setValue(
				`${name}.params.priority`,
				normaliseForSelect(
					NtfyPriorityOptions,
					getValues(`${name}.params.priority`),
				)?.value || 'default',
			);

		// Normalise selected scheme, or default it.
		if (convertedDefaults.params.scheme === '')
			setValue(
				`${name}.params.scheme`,
				normaliseForSelect(
					NtfySchemeOptions,
					getValues(`${name}.params.scheme`),
				)?.value || 'https',
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
					required
					col_sm={9}
					label="Host"
					tooltip="e.g. ntfy.example.com"
					defaultVal={convertedDefaults.url_fields.host}
				/>
				<FormText
					name={`${name}.url_fields.port`}
					col_sm={3}
					label="Port"
					tooltip="e.g. 443"
					isNumber
					defaultVal={convertedDefaults.url_fields.port}
					positionXS="right"
				/>
				<FormText
					name={`${name}.url_fields.username`}
					label="Username"
					defaultVal={convertedDefaults.url_fields.username}
				/>
				<FormText
					name={`${name}.url_fields.password`}
					label="Password"
					defaultVal={convertedDefaults.url_fields.password}
					positionXS="right"
				/>
				<FormText
					name={`${name}.url_fields.topic`}
					required
					col_sm={12}
					label="Topic"
					tooltip="Target topic"
					defaultVal={convertedDefaults.url_fields.topic}
				/>
			</>
			<FormLabel text="Params" heading />
			<>
				<FormSelect
					name={`${name}.params.scheme`}
					col_sm={6}
					col_lg={3}
					label="Scheme"
					tooltip="Server protocol"
					options={ntfySchemeOptions}
				/>
				<FormSelect
					name={`${name}.params.priority`}
					col_sm={6}
					col_lg={3}
					label="Priority"
					options={ntfyPriorityOptions}
					positionXS="right"
					positionLG="middle"
				/>
				<FormText
					name={`${name}.params.tags`}
					col_sm={12}
					col_lg={6}
					label="Tags"
					tooltip="Comma-separated list of tags that may or may not map to emojis"
					defaultVal={convertedDefaults.params.tags}
					positionLG="right"
				/>
				<FormText
					name={`${name}.params.attach`}
					col_sm={8}
					label="Attach"
					tooltip="URL of an attachment"
					defaultVal={convertedDefaults.params.attach}
				/>
				<FormText
					name={`${name}.params.filename`}
					col_sm={4}
					label="Filename"
					tooltip="File name of the attachment"
					defaultVal={convertedDefaults.params.filename}
					positionXS="right"
				/>
				<FormText
					name={`${name}.params.email`}
					label="E-mail"
					tooltip="E-mail address to send to"
					defaultVal={convertedDefaults.params.email}
				/>
				<FormText
					name={`${name}.params.title`}
					label="Title"
					defaultVal={convertedDefaults.params.title}
					positionXS="right"
				/>
				<FormText
					name={`${name}.params.click`}
					col_sm={12}
					label="Click"
					tooltip="URL to open when notification is clicked"
					defaultVal={convertedDefaults.params.click}
				/>
				<FormTextWithPreview
					name={`${name}.params.icon`}
					label="Icon"
					tooltip="URL to an icon"
					defaultVal={convertedDefaults.params.icon}
				/>
				<NtfyActions
					name={`${name}.params.actions`}
					label="Actions"
					tooltip="Custom action buttons for notifications"
					defaults={convertedDefaults.params.actions}
				/>
				<BooleanWithDefault
					name={`${name}.params.cache`}
					label="Cache"
					tooltip="Cache messages"
					defaultValue={convertedDefaults.params.cache}
				/>
				<BooleanWithDefault
					name={`${name}.params.firebase`}
					label="Firebase"
					tooltip="Send to Firebase Cloud Messaging"
					defaultValue={convertedDefaults.params.firebase}
				/>
			</>
		</>
	);
};
export default NTFY;
