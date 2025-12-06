import { useMemo } from 'react';
import { FieldText, FieldTextWithPreview } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyJoinSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Join` notifier.
 *
 * @param name - The `path` to this `Join` in the form.
 * @param main - The main values.
 */
const JOIN = ({ name, main }: { name: string; main?: NotifyJoinSchema }) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.join,
		[main, typeDataDefaults?.notify.join],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.apikey}
					label="API Key"
					name={`${name}.url_fields.apikey`}
					required
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.devices}
					label="Devices"
					name={`${name}.params.devices`}
					required
					tooltip={{
						content: 'e.g. ID1,ID2...',
						type: 'string',
					}}
				/>
				<FieldTextWithPreview
					defaultVal={defaults?.params?.icon}
					label="Icon"
					name={`${name}.params.icon`}
					tooltip={{
						content: 'URL of icon to use',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
					tooltip={{
						content: "e.g. 'Release - {{ service_name | default:service_id }}'",
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default JOIN;
