import { useMemo } from 'react';
import { FieldSelect, FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { withDefaultOption } from '@/components/modals/service-edit/util';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import {
	pagerDutyActionOptions,
	pagerDutySeverityOptions,
} from '@/utils/api/types/config/notify/pagerduty';
import type { NotifyPagerDutySchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `PagerDuty` notifier.
 *
 * @param name - The path to this `PagerDuty` in the form.
 * @param main - The main values.
 */
const PAGERDUTY = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyPagerDutySchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.pagerduty,
		[main, typeDataDefaults?.notify.pagerduty],
	);

	const pagerDutyActionOptionsNormalised = useMemo(
		() => withDefaultOption(pagerDutyActionOptions, defaults?.params?.action),
		[defaults?.params?.action],
	);

	const pagerDutySeverityOptionsNormalised = useMemo(
		() =>
			withDefaultOption(pagerDutySeverityOptions, defaults?.params?.severity),
		[defaults?.params?.severity],
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
					tooltip={{
						content: 'e.g. events.pagerduty.com',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 3, xs: 3 }}
					defaultVal={defaults?.url_fields?.port}
					label="Port"
					name={`${name}.url_fields.port`}
					tooltip={{
						content: 'e.g. 443',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.integrationkey}
					label="Integration Key"
					name={`${name}.url_fields.integrationkey`}
					required
					tooltip={{
						content: 'Events API v2 integration key',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldSelect
					colSize={{ sm: 6 }}
					label="Action"
					name={`${name}.params.action`}
					options={pagerDutyActionOptionsNormalised}
					tooltip={{
						content: 'The type of event',
						type: 'string',
					}}
				/>
				<FieldSelect
					colSize={{ sm: 6 }}
					label="Severity"
					name={`${name}.params.severity`}
					options={pagerDutySeverityOptionsNormalised}
				/>
				<FieldText
					defaultVal={defaults?.params?.source}
					label="Source"
					name={`${name}.params.source`}
					tooltip={{
						content:
							'Unique location of the affected system, preferably a hostname or FQDN',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.dedup_key}
					label="Dedup Key"
					name={`${name}.params.dedup_key`}
					tooltip={{
						content:
							'Identifies the alert, at most 255 characters. Acknowledge and resolve events need it',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
					tooltip={{
						content: 'Sent as the event summary',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.client}
					label="Client"
					name={`${name}.params.client`}
					tooltip={{
						content: 'Name of the monitoring client triggering this event',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.client_url}
					label="Client URL"
					name={`${name}.params.client_url`}
					tooltip={{
						content: 'URL of the monitoring client triggering this event',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.details}
					label="Details"
					name={`${name}.params.details`}
					tooltip={{
						content: 'Additional details about the incident, as a JSON object',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.contexts}
					label="Contexts"
					name={`${name}.params.contexts`}
					tooltip={{
						content: 'Additional context links or images',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default PAGERDUTY;
