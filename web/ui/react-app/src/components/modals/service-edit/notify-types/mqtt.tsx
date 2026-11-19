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
import { mqttQoSOptions } from '@/utils/api/types/config/notify/mqtt';
import type { NotifyMQTTSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for an `MQTT` notifier.
 *
 * @param name - The path to this `MQTT` in the form.
 * @param main - The main values.
 */
const MQTT = ({ name, main }: { name: string; main?: NotifyMQTTSchema }) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.mqtt,
		[main, typeDataDefaults?.notify.mqtt],
	);

	const mqttQoSOptionsNormalised = useMemo(
		() => withDefaultOption(mqttQoSOptions, defaults?.params?.qos),
		[defaults?.params?.qos],
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
						content: 'e.g. mqtt.example.com',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 3, xs: 3 }}
					defaultVal={defaults?.url_fields?.port}
					label="Port"
					name={`${name}.url_fields.port`}
					tooltip={{
						content: 'Defaults to 8883 with TLS, 1883 without',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.username}
					label="Username"
					name={`${name}.url_fields.username`}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.password}
					label="Password"
					name={`${name}.url_fields.password`}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.topic}
					label="Topic"
					name={`${name}.url_fields.topic`}
					required
					tooltip={{
						content: 'e.g. home/argus/releases',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldSelect
					colSize={{ sm: 6 }}
					label="QoS"
					name={`${name}.params.qos`}
					options={mqttQoSOptionsNormalised}
					tooltip={{
						content: 'Quality of Service level',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.params?.clientid}
					label="Client ID"
					name={`${name}.params.clientid`}
					tooltip={{
						content: 'MQTT client identifier. Defaults to a random one',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.retained}
					label="Retained"
					name={`${name}.params.retained`}
					tooltip={{
						content: 'Retain the message on the broker',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.cleansession}
					label="Clean Session"
					name={`${name}.params.cleansession`}
					tooltip={{
						content: 'Start with a clean session',
						type: 'string',
					}}
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

export default MQTT;
