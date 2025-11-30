import { useMemo } from 'react';
import { FieldColour, FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyTeamsSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Teams` notifier.
 *
 * @param name - The path to this `Teams` in the form.
 * @param main - The main values.
 */
const TEAMS = ({ name, main }: { name: string; main?: NotifyTeamsSchema }) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.teams,
		[main, typeDataDefaults?.notify.teams],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					defaultVal={defaults?.url_fields?.altid}
					label="Alt ID"
					name={`${name}.url_fields.altid`}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.tenant}
					label="Tenant"
					name={`${name}.url_fields.tenant`}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.group}
					label="Group"
					name={`${name}.url_fields.group`}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.groupowner}
					label="Group Owner"
					name={`${name}.url_fields.groupowner`}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ xs: 6 }}
					defaultVal={defaults?.params?.host}
					label="Host"
					name={`${name}.params.host`}
				/>
				<FieldText
					colSize={{ xs: 6 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<FieldColour
					colSize={{ sm: 6 }}
					defaultVal={defaults?.params?.color}
					label="Color"
					name={`${name}.params.color`}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default TEAMS;
