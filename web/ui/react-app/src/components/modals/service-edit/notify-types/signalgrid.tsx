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
import { signalgridTypeOptions } from '@/utils/api/types/config/notify/signalgrid';
import type { NotifySignalgridSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Signalgrid` notifier.
 *
 * @param name - The path to this `Signalgrid` in the form.
 * @param main - The main values.
 */
const SIGNALGRID = ({
	name,
	main,
}: {
	name: string;
	main?: NotifySignalgridSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.signalgrid,
		[main, typeDataDefaults?.notify.signalgrid],
	);

	const signalgridTypeOptionsNormalised = useMemo(
		() => withDefaultOption(signalgridTypeOptions, defaults?.params?.type),
		[defaults?.params?.type],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.clientkey}
					label="Client Key"
					name={`${name}.url_fields.clientkey`}
					required
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.channel}
					label="Channel"
					name={`${name}.url_fields.channel`}
					required
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldSelect
					colSize={{ sm: 4 }}
					label="Type"
					name={`${name}.params.type`}
					options={signalgridTypeOptionsNormalised}
					tooltip={{
						content: 'Notification severity',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 8 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.critical}
					label="Critical"
					name={`${name}.params.critical`}
					tooltip={{
						content: 'Deliver as a critical alert',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default SIGNALGRID;
