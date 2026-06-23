import { useEffect, useMemo } from 'react';
import { useFormContext } from 'react-hook-form';
import { BooleanWithDefault } from '@/components/generic';
import { FieldSelect, FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { normaliseForSelect } from '@/components/modals/service-edit/util';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import {
	type ZulipType,
	zulipTypeOptions,
} from '@/utils/api/types/config/notify/zulip';
import type { NotifyZulipSchema } from '@/utils/api/types/config-edit/notify/schemas';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';
import { ensureValue } from '@/utils/form-utils';

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
	const { getValues, setValue } = useFormContext();
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.zulip,
		[main, typeDataDefaults?.notify.zulip],
	);

	// Ensure selects have a valid value.
	// biome-ignore lint/correctness/useExhaustiveDependencies: fallback on first load.
	useEffect(() => {
		ensureValue<ZulipType>({
			defaultValue: defaults?.params?.type,
			fallback: zulipTypeOptions[0].value,
			getValues,
			path: `${name}.params.type`,
			setValue,
		});
	}, [main]);

	const zulipTypeOptionsNormalised = useMemo(() => {
		const defaultType = normaliseForSelect(
			zulipTypeOptions,
			defaults?.params?.type,
		);

		if (defaultType)
			return [
				{ label: `${defaultType.label} (default)`, value: nullString },
				...zulipTypeOptions,
			];

		return zulipTypeOptions;
	}, [defaults?.params?.type]);

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
						content: 'e.g. zulip.example.com',
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
					defaultVal={defaults?.url_fields?.botkey}
					label="Bot Key"
					name={`${name}.url_fields.botkey`}
					required
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.botmail}
					label="Bot Mail"
					name={`${name}.url_fields.botmail`}
					required
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldSelect
					colSize={{ xs: 6 }}
					label="Message Type"
					name={`${name}.params.type`}
					options={zulipTypeOptionsNormalised}
				/>
				<FieldText
					colSize={{ xs: 6 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
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
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.to}
					label="To (DM)"
					name={`${name}.params.to`}
					tooltip={{
						content: 'Comma-separated user IDs or emails for direct messages',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.read_by_sender}
					label="Mark Read by Sender"
					name={`${name}.params.read_by_sender`}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default ZULIP_CHAT;
