import { useEffect, useMemo } from 'react';
import { useFormContext } from 'react-hook-form';
import { FieldSelect, FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { normaliseForSelect } from '@/components/modals/service-edit/util';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import {
	type IFTTTMessageValue,
	type IFTTTTitleValue,
	iftttMessageValueOptions,
	iftttTitleValueOptions,
} from '@/utils/api/types/config/notify/ifttt';
import type { NotifyIFTTTSchema } from '@/utils/api/types/config-edit/notify/schemas';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';
import { ensureValue } from '@/utils/form-utils';

/**
 * The form fields for an `IFTTT` notifier.
 *
 * @param name - The path to this `IFTTT` in the form.
 * @param main - The main values.
 */
const IFTTT = ({ name, main }: { name: string; main?: NotifyIFTTTSchema }) => {
	const { getValues, setValue } = useFormContext();
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.ifttt,
		[main, typeDataDefaults?.notify.ifttt],
	);

	// Ensure selects have a valid value.
	// biome-ignore lint/correctness/useExhaustiveDependencies: fallback on first load.
	useEffect(() => {
		ensureValue<IFTTTMessageValue>({
			defaultValue: defaults?.params?.messagevalue,
			fallback: iftttMessageValueOptions[0].value,
			getValues,
			path: `${name}.params.messagevalue`,
			setValue,
		});
		ensureValue<IFTTTTitleValue>({
			defaultValue: defaults?.params?.titlevalue,
			fallback: iftttTitleValueOptions[0].value,
			getValues,
			path: `${name}.params.titlevalue`,
			setValue,
		});
	}, [main]);

	const messageValueOptionsNormalised = useMemo(() => {
		const defaultMessageValue = normaliseForSelect(
			iftttMessageValueOptions,
			defaults?.params?.messagevalue,
		);

		if (defaultMessageValue)
			return [
				{
					label: `${defaultMessageValue.label} (default)`,
					value: nullString,
				},
				...iftttMessageValueOptions,
			];

		return iftttMessageValueOptions;
	}, [defaults?.params?.messagevalue]);

	const titleValueOptionsNormalised = useMemo(() => {
		const defaultTitleValue = normaliseForSelect(
			iftttTitleValueOptions,
			defaults?.params?.titlevalue,
		);

		if (defaultTitleValue)
			return [
				{
					label: `${defaultTitleValue.label} (default)`,
					value: nullString,
				},
				...iftttTitleValueOptions,
			];

		return iftttTitleValueOptions;
	}, [defaults?.params?.titlevalue]);

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
					colSize={{ lg: 6, sm: 12 }}
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
					colSize={{ lg: 6, sm: 12 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
					tooltip={{
						content: 'Optional notification title',
						type: 'string',
					}}
				/>
				<FieldSelect
					label="Message Value"
					name={`${name}.params.messagevalue`}
					options={messageValueOptionsNormalised}
					tooltip={{
						content: 'Value field (1-3) to use as the notification message',
						type: 'string',
					}}
				/>
				<FieldSelect
					label="Title Value"
					name={`${name}.params.titlevalue`}
					options={titleValueOptionsNormalised}
					tooltip={{
						content:
							'Value field (1-3) to use as the notification title, or None to disable',
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
