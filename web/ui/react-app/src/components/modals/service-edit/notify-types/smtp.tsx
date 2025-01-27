import { FormLabel, FormSelect, FormText } from 'components/generic/form';
import { firstNonDefault, strToBool } from 'utils';
import { useEffect, useMemo } from 'react';

import { BooleanWithDefault } from 'components/generic';
import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { NotifySMTPType } from 'types/config';
import { normaliseForSelect } from 'components/modals/service-edit/util/normalise-selects';
import { useFormContext } from 'react-hook-form';

export const SMTPAuthOptions = [
	{ label: 'None', value: 'None' },
	{ label: 'Plain', value: 'Plain' },
	{ label: 'CRAM-MD5', value: 'CRAMMD5' },
	{ label: 'Unknown', value: 'Unknown' },
	{ label: 'OAuth2', value: 'OAuth2' },
];
export const SMTPEncryptionOptions = [
	{ label: 'Auto', value: 'Auto' },
	{ label: 'ExplicitTLS', value: 'ExplicitTLS' },
	{ label: 'ImplicitTLS', value: 'ImplicitTLS' },
	{ label: 'None', value: 'None' },
];

/**
 * The form fields for an SMTP notifier.
 *
 * @param name - The path to this `SMTP` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `SMTP` notifier.
 */
const SMTP = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifySMTPType;
	defaults?: NotifySMTPType;
	hard_defaults?: NotifySMTPType;
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
				username: firstNonDefault(
					main?.url_fields?.username,
					defaults?.url_fields?.username,
					hard_defaults?.url_fields?.username,
				),
			},
			// Params
			params: {
				auth: firstNonDefault(
					main?.params?.auth,
					defaults?.params?.auth,
					hard_defaults?.params?.auth,
				).toLowerCase(),
				clienthost: firstNonDefault(
					main?.params?.clienthost,
					defaults?.params?.clienthost,
					hard_defaults?.params?.clienthost,
				),
				encryption: firstNonDefault(
					main?.params?.encryption,
					defaults?.params?.encryption,
					hard_defaults?.params?.encryption,
				).toLowerCase(),
				fromaddress: firstNonDefault(
					main?.params?.fromaddress,
					defaults?.params?.fromaddress,
					hard_defaults?.params?.fromaddress,
				),
				fromname: firstNonDefault(
					main?.params?.fromname,
					defaults?.params?.fromname,
					hard_defaults?.params?.fromname,
				),
				subject: firstNonDefault(
					main?.params?.subject,
					defaults?.params?.subject,
					hard_defaults?.params?.subject,
				),
				toaddresses: firstNonDefault(
					main?.params?.toaddresses,
					defaults?.params?.toaddresses,
					hard_defaults?.params?.toaddresses,
				),
				usehtml:
					strToBool(
						firstNonDefault(
							main?.params?.usehtml,
							defaults?.params?.usehtml,
							hard_defaults?.params?.usehtml,
						),
					) ?? false,
				usestarttls:
					strToBool(
						firstNonDefault(
							main?.params?.usestarttls,
							defaults?.params?.usestarttls,
							hard_defaults?.params?.usestarttls,
						),
					) ?? true,
			},
		}),
		[main, defaults, hard_defaults],
	);

	const smtpAuthOptions = useMemo(() => {
		const defaultParamsAuthLabel = normaliseForSelect(
			SMTPAuthOptions,
			convertedDefaults.params.auth,
		);

		if (defaultParamsAuthLabel)
			return [
				{ value: '', label: `${defaultParamsAuthLabel.label} (default)` },
				...SMTPAuthOptions,
			];

		return SMTPAuthOptions;
	}, [convertedDefaults.params.auth]);

	const smtpEncryptionOptions = useMemo(() => {
		const defaultParamsEncryptionLabel = normaliseForSelect(
			SMTPEncryptionOptions,
			convertedDefaults.params.encryption,
		);

		if (defaultParamsEncryptionLabel)
			return [
				{ value: '', label: `${defaultParamsEncryptionLabel.label} (default)` },
				...SMTPEncryptionOptions,
			];

		return SMTPEncryptionOptions;
	}, [convertedDefaults.params.encryption]);

	useEffect(() => {
		const currentAuth = getValues(`${name}.params.auth`);
		// Normalise selected auth, or default it.
		if (convertedDefaults.params.auth === '')
			setValue(
				`${name}.params.auth`,
				normaliseForSelect(SMTPAuthOptions, currentAuth)?.value || 'Unknown',
			);

		// Normalise selected encryption, or default it.
		if (convertedDefaults.params.encryption === '')
			setValue(
				`${name}.params.encryption`,
				normaliseForSelect(
					SMTPEncryptionOptions,
					getValues(`${name}.params.encryption`),
				)?.value || 'Auto',
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
					tooltip="e.g. smtp.example.com"
					defaultVal={convertedDefaults.url_fields.host}
				/>
				<FormText
					name={`${name}.url_fields.port`}
					col_sm={3}
					label="Port"
					tooltip="e.g. 25/465/587/2525"
					isNumber
					defaultVal={convertedDefaults.url_fields.port}
					position="right"
				/>
				<FormText
					name={`${name}.url_fields.username`}
					label="Username"
					tooltip="e.g. something@example.com"
					defaultVal={convertedDefaults.url_fields.username}
				/>
				<FormText
					name={`${name}.url_fields.password`}
					label="Password"
					defaultVal={convertedDefaults.url_fields.password}
					position="right"
				/>
			</>
			<FormLabel text="Params" heading />
			<>
				<FormText
					name={`${name}.params.toaddresses`}
					required
					col_sm={12}
					label="To Address(es)"
					tooltip="Email(s) to send to (Comma separated)"
					defaultVal={convertedDefaults.params.toaddresses}
				/>
				<FormText
					name={`${name}.params.fromaddress`}
					required
					label="From Address"
					tooltip="Email to send from"
					defaultVal={convertedDefaults.params.fromaddress}
				/>
				<FormText
					name={`${name}.params.fromname`}
					label="From Name"
					tooltip="Name to send as"
					defaultVal={convertedDefaults.params.fromname}
					position="right"
				/>
				<FormSelect
					name={`${name}.params.auth`}
					col_sm={4}
					label="Auth"
					options={smtpAuthOptions}
				/>
				<FormText
					name={`${name}.params.subject`}
					col_sm={8}
					label="Subject"
					tooltip="Email subject"
					defaultVal={convertedDefaults.params.subject}
					position="right"
				/>
				<FormSelect
					name={`${name}.params.encryption`}
					col_sm={4}
					label="Encryption"
					tooltip="Encryption method"
					options={smtpEncryptionOptions}
				/>
				<FormText
					name={`${name}.params.clienthost`}
					col_sm={8}
					label="Client Host"
					tooltip={`The client host name sent to the SMTP server during HELO phase.
						If set to "auto", it will use the OS hostname`}
					defaultVal={convertedDefaults.params.clienthost}
					position="right"
				/>
				<BooleanWithDefault
					name={`${name}.params.usehtml`}
					label="Use HTML"
					tooltip="Whether 'message' is in HTML"
					defaultValue={convertedDefaults.params.usehtml}
				/>
				<BooleanWithDefault
					name={`${name}.params.usestarttls`}
					label="Use StartTLS"
					tooltip="Use StartTLS encryption"
					defaultValue={convertedDefaults.params.usestarttls}
				/>
			</>
		</>
	);
};

export default SMTP;
