import { useEffect, useMemo } from 'react';
import { useFormContext } from 'react-hook-form';
import { BooleanWithDefault } from '@/components/generic';
import { FieldSelect, FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { normaliseForSelect } from '@/components/modals/service-edit/util/normalise';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import {
	type SMTPAuth,
	type SMTPEncryption,
	smtpAuthOptions,
	smtpEncryptionOptions,
} from '@/utils/api/types/config/notify/smtp';
import type { NotifySMTPSchema } from '@/utils/api/types/config-edit/notify/schemas';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';
import { ensureValue } from '@/utils/form-utils';

/**
 * The form fields for an `SMTP` notifier.
 *
 * @param name - The path to this `SMTP` in the form.
 * @param main - The main values.
 */
const SMTP = ({ name, main }: { name: string; main?: NotifySMTPSchema }) => {
	const { getValues, setValue } = useFormContext();
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.smtp,
		[main, typeDataDefaults?.notify.smtp],
	);

	// Ensure selects have a valid value.
	// biome-ignore lint/correctness/useExhaustiveDependencies: fallback on first load.
	useEffect(() => {
		ensureValue<SMTPAuth>({
			defaultValue: defaults?.params?.auth,
			fallback: Object.values(smtpAuthOptions)[0].value,
			getValues,
			path: `${name}.params.auth`,
			setValue,
		});
		ensureValue<SMTPEncryption>({
			defaultValue: defaults?.params?.encryption,
			fallback: Object.values(smtpEncryptionOptions)[0].value,
			getValues,
			path: `${name}.params.encryption`,
			setValue,
		});
	}, [main]);

	const smtpAuthOptionsNormalised = useMemo(() => {
		const defaultParamsAuthLabel = normaliseForSelect(
			smtpAuthOptions,
			defaults?.params?.auth,
		);

		if (defaultParamsAuthLabel)
			return [
				{
					label: `${defaultParamsAuthLabel.label} (default)`,
					value: nullString,
				},
				...smtpAuthOptions,
			];

		return smtpAuthOptions;
	}, [defaults?.params?.auth]);

	const smtpEncryptionOptionsNormalised = useMemo(() => {
		const defaultParamsEncryptionLabel = normaliseForSelect(
			smtpEncryptionOptions,
			defaults?.params?.encryption,
		);

		if (defaultParamsEncryptionLabel)
			return [
				{
					label: `${defaultParamsEncryptionLabel.label} (default)`,
					value: nullString,
				},
				...smtpEncryptionOptions,
			];

		return smtpEncryptionOptions;
	}, [defaults?.params?.encryption]);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ xs: 9 }}
					defaultVal={defaults?.url_fields?.host}
					label="Host"
					name={`${name}.url_fields.host`}
					required
					tooltip={{
						content: 'e.g. smtp.example.com',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ xs: 3 }}
					defaultVal={defaults?.url_fields?.port}
					label="Port"
					name={`${name}.url_fields.port`}
					tooltip={{
						content: 'e.g. 25/465/587/2525',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.username}
					label="Username"
					name={`${name}.url_fields.username`}
					tooltip={{
						content: 'e.g. something@example.com',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.password}
					label="Password"
					name={`${name}.url_fields.password`}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ sm: 10, xs: 10 }}
					defaultVal={defaults?.params?.toaddresses}
					label="To Address(es)"
					name={`${name}.params.toaddresses`}
					required
					tooltip={{
						content: 'Emails to send to (Comma separated)',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 2, xs: 10 }}
					defaultVal={defaults?.params?.timeout}
					label="Timeout"
					name={`${name}.params.timeout`}
					required
					tooltip={{
						content: 'Timeout for send operations',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.fromaddress}
					label="From Address"
					name={`${name}.params.fromaddress`}
					required
					tooltip={{
						content: 'Email to send from',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.fromname}
					label="From Name"
					name={`${name}.params.fromname`}
					tooltip={{
						content: 'Name to send as',
						type: 'string',
					}}
				/>
				<FieldSelect
					colSize={{ sm: 4 }}
					label="Auth"
					name={`${name}.params.auth`}
					options={smtpAuthOptionsNormalised}
				/>
				<FieldText
					colSize={{ sm: 8 }}
					defaultVal={defaults?.params?.subject}
					label="Subject"
					name={`${name}.params.subject`}
					tooltip={{
						content: 'Email subject',
						type: 'string',
					}}
				/>
				<FieldSelect
					colSize={{ sm: 4 }}
					label="Encryption"
					name={`${name}.params.encryption`}
					options={smtpEncryptionOptionsNormalised}
					tooltip={{
						content: 'Encryption method',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 8 }}
					defaultVal={defaults?.params?.clienthost}
					label="Client Host"
					name={`${name}.params.clienthost`}
					tooltip={{
						content: `The client host name sent to the SMTP server during HELO phase.
						If set to "auto", it will use the OS hostname`,
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.usehtml}
					label="Use HTML"
					name={`${name}.params.usehtml`}
					tooltip={{
						content: "Whether 'message' is in HTML",
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.usestarttls}
					label="Use StartTLS"
					name={`${name}.params.usestarttls`}
					tooltip={{
						content: 'Use StartTLS encryption',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.requirestarttls}
					label="Require StartTLS"
					name={`${name}.params.requirestarttls`}
					tooltip={{
						content: 'Fail if StartTLS is enabled but unsupported',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.skiptlsverification}
					label="Skip TLS Verification"
					name={`${name}.params.skiptlsverify`}
					tooltip={{
						content: 'Skip TLS certificate verification',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default SMTP;
