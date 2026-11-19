import { useMemo } from 'react';
import { FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyWeComSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `WeCom` notifier.
 *
 * @param name - The path to this `WeCom` in the form.
 * @param main - The main values.
 */
const WECOM = ({ name, main }: { name: string; main?: NotifyWeComSchema }) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.wecom,
		[main, typeDataDefaults?.notify.wecom],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.key}
					label="Webhook Key"
					name={`${name}.url_fields.key`}
					required
					tooltip={{
						content: 'The key=... value of the group bot webhook URL',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					defaultVal={defaults?.params?.mentioned_list}
					label="Mentioned Users"
					name={`${name}.params.mentioned_list`}
					tooltip={{
						content: 'Comma-separated user IDs to mention, or @all',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.mentioned_mobile_list}
					label="Mentioned Mobiles"
					name={`${name}.params.mentioned_mobile_list`}
					tooltip={{
						content: 'Comma-separated mobile numbers to mention, or @all',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default WECOM;
