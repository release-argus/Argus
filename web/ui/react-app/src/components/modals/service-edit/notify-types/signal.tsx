import { useMemo } from 'react';
import { BooleanWithDefault } from '@/components/generic';
import { FieldSelect, FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { withDefaultOption } from '@/components/modals/service-edit/util';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import { signalTextModeOptions } from '@/utils/api/types/config/notify/signal';
import type { NotifySignalSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Signal` notifier.
 *
 * @param name - The path to this `Signal` in the form.
 * @param main - The main values.
 */
const SIGNAL = ({
	name,
	main,
}: {
	name: string;
	main?: NotifySignalSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.signal,
		[main, typeDataDefaults?.notify.signal],
	);

	const signalTextModeOptionsNormalised = useMemo(
		() => withDefaultOption(signalTextModeOptions, defaults?.params?.textmode),
		[defaults?.params?.textmode],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ sm: 9, xs: 9 }}
					defaultVal={defaults?.url_fields?.host}
					label="Host"
					name={`${name}.url_fields.host`}
					required
					tooltip={{
						content: 'Signal REST API server, e.g. signal-api.example.com',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 3, xs: 3 }}
					defaultVal={defaults?.url_fields?.port}
					label="Port"
					name={`${name}.url_fields.port`}
					tooltip={{
						content: 'e.g. 8080',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.user}
					label="User"
					name={`${name}.url_fields.user`}
					tooltip={{
						content: 'For HTTP Basic Auth',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.password}
					label="Password"
					name={`${name}.url_fields.password`}
					tooltip={{
						content: 'For HTTP Basic Auth',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 4 }}
					defaultVal={defaults?.url_fields?.source}
					label="Source"
					name={`${name}.url_fields.source`}
					required
					tooltip={{
						content: 'Source phone number, with country code, e.g. +1234567890',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 8 }}
					defaultVal={defaults?.url_fields?.recipients}
					label="Recipients"
					name={`${name}.url_fields.recipients`}
					required
					tooltip={{
						content:
							'Comma-separated phone numbers, group IDs, or u: usernames',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.token}
					label="API Token"
					name={`${name}.params.token`}
					tooltip={{
						content: 'For Bearer authentication',
						type: 'string',
					}}
				/>
				<FieldSelect
					colSize={{ sm: 4 }}
					label="Text Mode"
					name={`${name}.params.textmode`}
					options={signalTextModeOptionsNormalised}
					tooltip={{
						content: 'Styled enables markup',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 8 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.attachments}
					label="Attachments"
					name={`${name}.params.attachments`}
					tooltip={{
						content:
							'Base64-encoded files, comma-separated. Or one data: URI, e.g. data:image/png;base64,...',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.notifyself}
					label="Notify Self"
					name={`${name}.params.notifyself`}
					tooltip={{
						content: "Notify the sender's devices",
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.disabletls}
					label="Disable TLS"
					name={`${name}.params.disabletls`}
					tooltip={{
						content: 'Send over HTTP instead of HTTPS',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.skiptlsverify}
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

export default SIGNAL;
