import { useMemo } from 'react';
import { FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyLarkSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Lark` notifier.
 *
 * @param name - The path to this `Lark` in the form.
 * @param main - The main values.
 */
const LARK = ({ name, main }: { name: string; main?: NotifyLarkSchema }) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.lark,
		[main, typeDataDefaults?.notify.lark],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.host}
					label="Host"
					name={`${name}.url_fields.host`}
					tooltip={{
						content: 'open.larksuite.com (Lark) or open.feishu.cn (Feishu)',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.token}
					label="Token"
					name={`${name}.url_fields.token`}
					required
					tooltip={{
						ariaLabel:
							'Custom bot token, from the end of the webhook URL (.../hook/TOKEN)',
						content: (
							<>
								{'Custom bot token, from the end of the webhook URL (.../hook/'}
								<span className="bold underline">TOKEN</span>
								{')'}
							</>
						),
						type: 'element',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.secret}
					label="Secret"
					name={`${name}.params.secret`}
					tooltip={{
						content: 'Custom bot secret, for signature verification',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<FieldText
					defaultVal={defaults?.params?.link}
					label="Link"
					name={`${name}.params.link`}
					tooltip={{
						content: 'URL to link to from the message',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default LARK;
