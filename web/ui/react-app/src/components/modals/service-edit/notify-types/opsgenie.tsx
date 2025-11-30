import { useMemo } from 'react';
import {
	FieldKeyValMap,
	FieldList,
	FieldText,
} from '@/components/generic/field';
import { OpsGenieTargets } from '@/components/modals/service-edit/notify-types/extra';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyOpsGenieSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for an `OpsGenie` notifier.
 *
 * @param name - The path to this `OpsGenie` in the form.
 * @param main - The main values.
 */
const OPSGENIE = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyOpsGenieSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.opsgenie,
		[main, typeDataDefaults?.notify.opsgenie],
	);

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
					tooltip={{
						content:
							"The OpsGenie API host. Use 'api.eu.opsgenie.com' for EU instances",
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ xs: 3 }}
					defaultVal={defaults?.url_fields?.port}
					label="Port"
					name={`${name}.url_fields.port`}
				/>
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
				<FieldList
					defaults={defaults?.params?.actions}
					label="Actions"
					name={`${name}.params.actions`}
					tooltip={{
						content: 'Custom actions that will be available for the alert',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.alias}
					label="Alias"
					name={`${name}.params.alias`}
					tooltip={{
						content: 'Client-defined identifier of the alert',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.description}
					label="Description"
					name={`${name}.params.description`}
					tooltip={{
						content: 'Description field of the alert',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.note}
					label="Note"
					name={`${name}.params.note`}
					tooltip={{
						content:
							'Additional note that will be added while creating the alert',
						type: 'string',
					}}
				/>
				<FieldKeyValMap
					defaults={defaults?.params?.details}
					label="Details"
					name={`${name}.params.details`}
					placeholders={{
						key: 'e.g. X-Authorization',
						value: "e.g. 'Bearer TOKEN'",
					}}
					tooltip={{
						content: 'Map of key-val custom props of the alert',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.entity}
					label="Entity"
					name={`${name}.params.entity`}
					tooltip={{
						content:
							'Entity field of the alert that is generally used to specify which domain the Source field of the alert',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.priority}
					label="Priority"
					name={`${name}.params.priority`}
					tooltip={{
						content: 'Priority level of the alert. 1/2/3/4/5',
						type: 'string',
					}}
				/>
				<OpsGenieTargets
					defaults={defaults?.params?.responders}
					label="Responders"
					name={`${name}.params.responders`}
					tooltip={{
						content:
							'Teams, users, escalations and schedules that the alert will be routed to',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.source}
					label="Source"
					name={`${name}.params.source`}
					tooltip={{
						content: 'Source field of the alert',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.tags}
					label="Tags"
					name={`${name}.params.tags`}
					tooltip={{
						content: 'Tags of the alert',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
					tooltip={{
						content: 'Notification title, optionally set by the sender',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.user}
					label="User"
					name={`${name}.params.user`}
					tooltip={{
						content: 'Display name of the request owner',
						type: 'string',
					}}
				/>
				<OpsGenieTargets
					defaults={defaults?.params?.visibleto}
					label="Visible To"
					name={`${name}.params.visibleto`}
					tooltip={{
						content:
							'Teams and users that the alert will become visible to without sending any notification',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default OPSGENIE;
