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
	type TelegramParsemode,
	telegramParsemodeOptions,
} from '@/utils/api/types/config/notify/telegram';
import type { NotifyTelegramSchema } from '@/utils/api/types/config-edit/notify/schemas';
import { ensureValue } from '@/utils/form-utils';

/**
 * The form fields for a `Telegram` notifier.
 *
 * @param name - The path to this `Telegram` in the form.
 * @param defaults - The default values.
 */
const TELEGRAM = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyTelegramSchema;
}) => {
	const { getValues, setValue } = useFormContext();
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.telegram,
		[main, typeDataDefaults?.notify.telegram],
	);

	// Ensure selects have a valid value.
	// biome-ignore lint/correctness/useExhaustiveDependencies: fallback on first load.
	useEffect(() => {
		ensureValue<TelegramParsemode>({
			defaultValue: defaults?.params?.parsemode,
			fallback: Object.values(telegramParsemodeOptions)[0].value,
			getValues,
			path: `${name}.params.parsemode`,
			setValue,
		});
	}, [main]);

	const telegramParseModeOptions = useMemo(() => {
		const defaultParseMode = normaliseForSelect(
			telegramParsemodeOptions,
			defaults?.params?.parsemode,
		);

		if (defaultParseMode)
			return [
				{ label: `${defaultParseMode.label} (default)`, value: '' },
				...telegramParsemodeOptions,
			];

		return telegramParsemodeOptions;
	}, [defaults?.params?.parsemode]);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.token}
					label="Token"
					name={`${name}.url_fields.token`}
					required
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ sm: 8 }}
					defaultVal={defaults?.params?.chats}
					label="Chats"
					name={`${name}.params.chats`}
					required
					tooltip={{
						content: 'Chat IDs or Channel names, e.g. -123,@bar',
						type: 'string',
					}}
				/>
				<FieldSelect
					colSize={{ sm: 4 }}
					label="Parse Mode"
					name={`${name}.params.parsemode`}
					options={telegramParseModeOptions}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.notification}
					label="Notification"
					name={`${name}.params.notification`}
					tooltip={{
						content: 'Disable for silent messages',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.preview}
					label="Preview"
					name={`${name}.params.preview`}
					tooltip={{
						content: 'Enable web page previews on messages',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default TELEGRAM;
