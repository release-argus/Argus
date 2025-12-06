import { useMemo } from 'react';
import { FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyIFTTTSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for an `IFTTT` notifier.
 *
 * @param name - The path to this `IFTTT` in the form.
 * @param main - The main values.
 */
const IFTTT = ({ name, main }: { name: string; main?: NotifyIFTTTSchema }) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.ifttt,
		[main, typeDataDefaults?.notify.ifttt],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.webhookid}
					label="WebHook ID"
					name={`${name}.url_fields.webhookid`}
					required
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.events}
					label="Events"
					name={`${name}.params.events`}
					required
					tooltip={{
						content: 'e.g. event1,event2...',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
					tooltip={{
						content: 'Optional notification title',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.usemessageasvalue}
					label="Use Message As Value"
					name={`${name}.params.usemessageasvalue`}
					tooltip={{
						content: 'Set the corresponding value field to the message',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.usetitleasvalue}
					label="Use Title As Value"
					name={`${name}.params.usetitleasvalue`}
					tooltip={{
						content: 'Set the corresponding value field to the title',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 4 }}
					defaultVal={defaults?.params?.value1}
					label="Value1"
					name={`${name}.params.value1`}
				/>
				<FieldText
					colSize={{ sm: 4 }}
					defaultVal={defaults?.params?.value2}
					label="Value2"
					name={`${name}.params.value2`}
				/>
				<FieldText
					colSize={{ sm: 4 }}
					defaultVal={defaults?.params?.value3}
					label="Value3"
					name={`${name}.params.value3`}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default IFTTT;
