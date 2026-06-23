import { useMemo } from 'react';
import { FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyShoutrrrSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a raw `Shoutrrr` URL notifier.
 *
 * This type is a hidden/deprecated passthrough for legacy configs
 * (e.g. old-style Microsoft Teams Office 365 Connector webhooks).
 * New notifiers should use a specific type instead.
 *
 * @param name - The path to this `Shoutrrr` in the form.
 * @param main - The main values.
 */
const SHOUTRRR = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyShoutrrrSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.shoutrrr,
		[main, typeDataDefaults?.notify.shoutrrr],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.raw}
					label="Raw"
					name={`${name}.url_fields.raw`}
					required
					tooltip={{
						content:
							'Full shoutrrr-format URL, e.g. slack://TOKEN@CHANNEL. ' +
							'This type is for unsupported configs - consider raising a feature request to have your type added.',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default SHOUTRRR;
