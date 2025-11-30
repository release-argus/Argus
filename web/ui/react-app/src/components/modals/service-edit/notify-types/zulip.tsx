import { useMemo } from 'react';
import { FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyZulipSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Zulip Chat` notifier.
 *
 * @param name - The path to this `Zulip Chat` in the form.
 * @param main - The main values.
 */
const ZULIP_CHAT = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyZulipSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.zulip,
		[main, typeDataDefaults?.notify.zulip],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					defaultVal={defaults?.url_fields?.botmail}
					label="Bot Mail"
					name={`${name}.url_fields.botmail`}
					required
					tooltip={{
						content: 'e.g. something@example.com',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.botkey}
					label="Bot Key"
					name={`${name}.url_fields.botkey`}
					required
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.host}
					label="Host"
					name={`${name}.url_fields.host`}
					required
					tooltip={{
						content: 'e.g. zulip.example.com',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					defaultVal={defaults?.params?.stream}
					label="Stream"
					name={`${name}.params.stream`}
				/>
				<FieldText
					defaultVal={defaults?.params?.topic}
					label="Topic"
					name={`${name}.params.topic`}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default ZULIP_CHAT;
