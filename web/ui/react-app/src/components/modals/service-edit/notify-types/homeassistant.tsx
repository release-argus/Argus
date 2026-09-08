import { useMemo } from 'react';
import { BooleanWithDefault } from '@/components/generic';
import { FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyHomeAssistantSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Home Assistant` notifier.
 *
 * @param name - The path to this `Home Assistant` in the form.
 * @param main - The main values.
 */
const HOME_ASSISTANT = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyHomeAssistantSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.homeassistant,
		[main, typeDataDefaults?.notify.homeassistant],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ lg: 6, sm: 9, xs: 9 }}
					defaultVal={defaults?.url_fields?.host}
					label="Host"
					name={`${name}.url_fields.host`}
					required
					tooltip={{
						content: 'e.g. homeassistant.example.com',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ lg: 2, sm: 3, xs: 3 }}
					defaultVal={defaults?.url_fields?.port}
					label="Port"
					name={`${name}.url_fields.port`}
					tooltip={{
						content: 'e.g. 443',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ lg: 4, sm: 12 }}
					defaultVal={defaults?.url_fields?.path}
					label="Path"
					name={`${name}.url_fields.path`}
					tooltip={{
						ariaLabel: 'Format: homeassistant.example.com/PATH',
						content: (
							<>
								<span className="text-muted-foreground">
									{'e.g. homeassistant.example.com/'}
								</span>
								<span className="bold underline">path</span>
							</>
						),
						type: 'element',
					}}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.token}
					label="Token"
					name={`${name}.url_fields.token`}
					required
					tooltip={{
						content: 'Long-lived access token',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.params?.service}
					label="Action"
					name={`${name}.params.service`}
					tooltip={{
						content:
							"e.g. 'notify.mobile_app_phone'. Empty creates a persistent notification",
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.params?.targets}
					label="Targets"
					name={`${name}.params.targets`}
					tooltip={{
						content: 'Notify targets, e.g. device1,device2',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.nid}
					label="Notification ID"
					name={`${name}.params.nid`}
					tooltip={{
						content:
							'Persistent notification ID, replacing any with the same ID',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.disabletls}
					label="Disable TLS"
					name={`${name}.params.disabletls`}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.skiptlsverify}
					label="Skip TLS Verify"
					name={`${name}.params.skiptlsverify`}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default HOME_ASSISTANT;
