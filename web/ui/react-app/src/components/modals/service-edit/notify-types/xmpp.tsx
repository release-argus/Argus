import { useMemo } from 'react';
import { BooleanWithDefault } from '@/components/generic';
import { FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyXMPPSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for an `XMPP` notifier.
 *
 * @param name - The path to this `XMPP` in the form.
 * @param main - The main values.
 */
const XMPP = ({ name, main }: { name: string; main?: NotifyXMPPSchema }) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.xmpp,
		[main, typeDataDefaults?.notify.xmpp],
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
						content: 'e.g. xmpp.example.com',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 3, xs: 3 }}
					defaultVal={defaults?.url_fields?.port}
					label="Port"
					name={`${name}.url_fields.port`}
					tooltip={{
						content: 'e.g. 5222',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.user}
					label="User"
					name={`${name}.url_fields.user`}
					required
					tooltip={{
						content: 'JID localpart or full JID',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.password}
					label="Password"
					name={`${name}.url_fields.password`}
					required
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					defaultVal={defaults?.params?.to}
					label="To"
					name={`${name}.params.to`}
					tooltip={{
						content: 'Comma-separated recipient JIDs',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.rooms}
					label="Rooms"
					name={`${name}.params.rooms`}
					tooltip={{
						content: 'Comma-separated MUC room JIDs',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.nick}
					label="Nickname"
					name={`${name}.params.nick`}
					tooltip={{
						content: 'MUC nickname',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.roompassword}
					label="Room Password"
					name={`${name}.params.roompassword`}
					tooltip={{
						content: 'Password for protected MUC rooms',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.disabletls}
					label="Disable TLS"
					name={`${name}.params.disabletls`}
					tooltip={{
						content: 'Connect without TLS',
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

export default XMPP;
