import { useMemo } from 'react';
import { BooleanWithDefault } from '@/components/generic';
import { FieldText, FieldTextWithPreview } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyDiscordSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Discord` notifier.
 *
 * @param name - The path to this `Discord` in the form.
 * @param main - The main values.
 */
const DISCORD = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyDiscordSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.discord,
		[main, typeDataDefaults?.notify.discord],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					defaultVal={defaults?.url_fields?.webhookid}
					label="WebHook ID"
					name={`${name}.url_fields.webhookid`}
					required
					tooltip={{
						ariaLabel:
							'Format: https://discord.com/api/webhooks/WEBHOOK_ID/token',
						content: (
							<>
								e.g. https://discord.com/api/webhooks/{''}
								<span className="bold-underline">webhook_id</span>
								{''}
								/token
							</>
						),
						type: 'element',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.token}
					label="Token"
					name={`${name}.url_fields.token`}
					required
					tooltip={{
						ariaLabel:
							'Format: https://discord.com/api/webhooks/webhook_id/TOKEN',
						content: (
							<>
								e.g. https://discord.com/api/webhooks/webhook_id/{''}
								<span className="bold-underline">token</span>
							</>
						),
						type: 'element',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldTextWithPreview
					defaultVal={defaults?.params?.avatar}
					label="Avatar"
					name={`${name}.params.avatar`}
					tooltip={{
						content: 'Override WebHook avatar with this URL',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.username}
					label="Username"
					name={`${name}.params.username`}
					tooltip={{
						content: 'Override the WebHook username',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.splitlines}
					label="Split Lines"
					name={`${name}.params.splitlines}`}
					tooltip={{
						content: 'Whether to send each line as a separate embedded item',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default DISCORD;
