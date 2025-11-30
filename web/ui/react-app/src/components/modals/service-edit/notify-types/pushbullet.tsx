import { useMemo } from 'react';
import { FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyPushbulletSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `PushBullet` notifier.
 *
 * @param name - The path to this `PushBullet` in the form.
 * @param main - The main values.
 */
const PUSHBULLET = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyPushbulletSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.pushbullet,
		[main, typeDataDefaults?.notify.pushbullet],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.token}
					label="Access Token"
					name={`${name}.url_fields.token`}
					required
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.targets}
					label="Targets"
					name={`${name}.url_fields.targets`}
					required
					tooltip={{
						content: 'e.g. DEVICE1,DEVICE2...',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default PUSHBULLET;
