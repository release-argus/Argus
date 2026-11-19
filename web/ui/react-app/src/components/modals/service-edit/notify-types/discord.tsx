import { useMemo } from 'react';
import { BooleanWithDefault } from '@/components/generic';
import {
	FieldColour,
	FieldText,
	FieldTextWithPreview,
} from '@/components/generic/field';
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
					label="Webhook ID"
					name={`${name}.url_fields.webhookid`}
					required
					tooltip={{
						ariaLabel:
							'Format: https://discord.com/api/webhooks/WEBHOOK_ID/token',
						content: (
							<>
								<span className="text-muted-foreground">
									{'e.g. https://discord.com/api/webhooks/'}
								</span>
								<span className="bold underline">webhook_id</span>
								<span className="text-muted-foreground">/token</span>
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
								<span className="text-muted-foreground">
									{'e.g. https://discord.com/api/webhooks/webhook_id/'}
								</span>
								<span className="bold underline">token</span>
							</>
						),
						type: 'element',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldTextWithPreview
					colSize={{ sm: 6 }}
					defaultVal={defaults?.params?.avatar}
					label="Avatar"
					name={`${name}.params.avatar`}
					tooltip={{
						content: 'Override Webhook avatar with this URL',
						type: 'string',
					}}
				/>
				<FieldColour
					colSize={{ sm: 6 }}
					defaultVal={defaults?.params?.color}
					label="Color"
					name={`${name}.params.color`}
					tooltip={{
						content: 'Message left-hand border color in hex, e.g. #ffffff',
						type: 'string',
					}}
				/>
				<FieldColour
					colSize={{ sm: 6 }}
					defaultVal={defaults?.params?.colordebug}
					label="Color (Debug)"
					name={`${name}.params.colordebug`}
					tooltip={{
						content: 'Border color for debug-level messages',
						type: 'string',
					}}
				/>
				<FieldColour
					colSize={{ sm: 6 }}
					defaultVal={defaults?.params?.colorinfo}
					label="Color (Info)"
					name={`${name}.params.colorinfo`}
					tooltip={{
						content: 'Border color for info-level messages',
						type: 'string',
					}}
				/>
				<FieldColour
					colSize={{ sm: 6 }}
					defaultVal={defaults?.params?.colorwarn}
					label="Color (Warn)"
					name={`${name}.params.colorwarn`}
					tooltip={{
						content: 'Border color for warning-level messages',
						type: 'string',
					}}
				/>
				<FieldColour
					colSize={{ sm: 6 }}
					defaultVal={defaults?.params?.colorerror}
					label="Color (Error)"
					name={`${name}.params.colorerror`}
					tooltip={{
						content: 'Border color for error-level messages',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.thread_id}
					label="Thread ID"
					name={`${name}.params.thread_id`}
					tooltip={{
						content: 'The Thread ID to send the message to',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.thread_name}
					label="Thread Name"
					name={`${name}.params.thread_name`}
					tooltip={{
						content:
							'Name of a thread to create for each message in a forum or media channel',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.username}
					label="Username"
					name={`${name}.params.username`}
					tooltip={{
						content: 'Override the Webhook username',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.mentions}
					label="Mentions"
					name={`${name}.params.mentions`}
					tooltip={{
						content:
							'Comma-separated mentions that ping: everyone, here, user:ID or role:ID',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.splitlines}
					label="Split Lines"
					name={`${name}.params.splitlines`}
					tooltip={{
						content: 'Whether to send each line as a separate embedded item',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.json}
					label="JSON"
					name={`${name}.params.json`}
					tooltip={{
						content: 'Send the message as a raw JSON payload',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default DISCORD;
