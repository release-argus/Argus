import { useMemo } from 'react';
import { FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyPushoverSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `PushOver` notifier.
 *
 * @param name - The path to this `PushOver` in the form.
 * @param main - The main values.
 */
const PUSHOVER = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyPushoverSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.pushover,
		[main, typeDataDefaults?.notify.pushover],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.token}
					label="API Token/Key"
					name={`${name}.url_fields.token`}
					required
					tooltip={{
						content:
							"'Create an Application/API Token' on the Pushover dashboard'",
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.user}
					label="User Key"
					name={`${name}.url_fields.user`}
					required
					tooltip={{
						content: 'Top right of Pushover dashboard',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.devices}
					label="Devices"
					name={`${name}.params.devices`}
					tooltip={{
						content:
							"e.g. device1,device2... (deviceX=Name column in the 'Your Devices' list)",
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ xs: 9 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<FieldText
					colSize={{ xs: 3 }}
					defaultVal={defaults?.params?.priority}
					label="Priority"
					name={`${name}.params.priority`}
					tooltip={{
						content:
							'Only supply priority values between -1 and 1, since 2 requires additional parameters that are not supported yet',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default PUSHOVER;
