import { useMemo } from 'react';
import { FieldTextArea } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyGoogleChatSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Google Chat` notifier.
 *
 * @param name - The path to this `Google Chat` in the form.
 * @param main - The main values.
 */
const GOOGLE_CHAT = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyGoogleChatSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.googlechat,
		[main, typeDataDefaults?.notify.googlechat],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldTextArea
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.raw}
					label="Raw"
					name={`${name}.url_fields.raw`}
					required
					rows={2}
					tooltip={{
						content:
							'e.g. chat.googleapis.com/v1/spaces/foo/messages?key=bar&token=baz',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default GOOGLE_CHAT;
