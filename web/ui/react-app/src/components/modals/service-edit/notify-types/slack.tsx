import { useMemo } from 'react';
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
import type { NotifySlackSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Slack` notifier.
 *
 * @param name - The path to this `Slack` in the form.
 * @param main - The main values.
 */
const SLACK = ({ name, main }: { name: string; main?: NotifySlackSchema }) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.slack,
		[main, typeDataDefaults?.notify.slack],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					defaultVal={defaults?.url_fields?.token}
					label="Token"
					name={`${name}.url_fields.token`}
					required
					tooltip={{
						ariaLabel: 'Format: xoxb:BOT-OAUTH-TOKEN or WEBHOOK',
						content: (
							<>
								{'xoxb:'}
								<span className="bold underline">BOT-OAUTH-TOKEN</span>
								{' or '}
								<span className="bold underline">WEBHOOK</span>
							</>
						),
						type: 'element',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.channel}
					label="Channel"
					name={`${name}.url_fields.channel`}
					required
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					defaultVal={defaults?.params?.botname}
					label="Bot Name"
					name={`${name}.params.botname`}
				/>
				<FieldColour
					defaultVal={defaults?.params?.color}
					label="Color"
					name={`${name}.params.color`}
					tooltip={{
						content: 'Message left-hand border color in hex, e.g. #ffffff',
						type: 'string',
					}}
				/>
				<FieldTextWithPreview
					defaultVal={defaults?.params?.icon}
					label="Icon"
					name={`${name}.params.icon`}
					tooltip={{
						content:
							'Use emoji or URL as icon (based on presence of http(s):// prefix)',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.threadts}
					label="Thread TS"
					name={`${name}.params.threadts`}
					tooltip={{
						content:
							'TS value of the parent message (to send message as reply in thread)',
						type: 'string',
					}}
					type="text"
				/>
				<FieldText
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
					tooltip={{
						content: 'Text prepended to the message',
						type: 'string',
					}}
					type="text"
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default SLACK;
