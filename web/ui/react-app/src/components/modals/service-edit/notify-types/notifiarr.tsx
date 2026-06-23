import { useMemo } from 'react';
import { FieldColour, FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyNotifiarrSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Notifiarr` notifier.
 *
 * @param name - The path to this `Notifiarr` in the form.
 * @param main - The main values.
 */
const NOTIFIARR = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyNotifiarrSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.notifiarr,
		[main, typeDataDefaults?.notify.notifiarr],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.apikey}
					label="API Key"
					name={`${name}.url_fields.apikey`}
					required
					tooltip={{
						content: 'Found on your Notifiarr account settings page',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					defaultVal={defaults?.params?.name}
					label="Name"
					name={`${name}.params.name`}
					tooltip={{
						content: 'App/script name shown in the notification',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.channel}
					label="Channel"
					name={`${name}.params.channel`}
					tooltip={{
						content: 'Discord channel ID for the notification',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.thumbnail}
					label="Thumbnail URL"
					name={`${name}.params.thumbnail`}
				/>
				<FieldText
					defaultVal={defaults?.params?.image}
					label="Image URL"
					name={`${name}.params.image`}
				/>
				<FieldColour
					defaultVal={defaults?.params?.color}
					label="Color"
					name={`${name}.params.color`}
					tooltip={{
						content: 'Color for embed',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default NOTIFIARR;
