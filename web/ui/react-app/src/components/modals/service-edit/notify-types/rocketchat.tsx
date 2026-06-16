import { useMemo } from 'react';
import { FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyRocketChatSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * Form fields for a `Rocket.Chat` notifier.
 *
 * @param name - The path to this `Rocket.Chat` in the form.
 * @param main - The main values.
 */
const ROCKET_CHAT = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyRocketChatSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.rocketchat,
		[main, typeDataDefaults?.notify.rocketchat],
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
						content: 'e.g. rocketchat.example.io',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 3, xs: 3 }}
					defaultVal={defaults?.url_fields?.port}
					label="Port"
					name={`${name}.url_fields.port`}
					required
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.channel}
					label="Channel"
					name={`${name}.url_fields.channel`}
					required
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.username}
					label="Username"
					name={`${name}.url_fields.username`}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.tokena}
					label="Token A"
					name={`${name}.url_fields.tokena`}
					required
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.tokenb}
					label="Token B"
					name={`${name}.url_fields.tokenb`}
					required
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default ROCKET_CHAT;
