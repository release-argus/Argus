import { useMemo } from 'react';
import { BooleanWithDefault } from '@/components/generic';
import { FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyMatrixSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Matrix` notifier.
 *
 * @param name - The path to this `Matrix` in the form.
 * @param main - The main values.
 */
const MATRIX = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyMatrixSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.matrix,
		[main, typeDataDefaults?.notify.matrix],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ xs: 9 }}
					defaultVal={defaults?.url_fields?.host}
					label="Host"
					name={`${name}.url_fields.host`}
					required
					tooltip={{
						content: 'e.g. smtp.example.com',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ xs: 3 }}
					defaultVal={defaults?.url_fields?.port}
					label="Port"
					name={`${name}.url_fields.port`}
					tooltip={{
						content: 'e.g. 25/465/587/2525',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.username}
					label="Username"
					name={`${name}.url_fields.username`}
					tooltip={{
						content: 'e.g. something@example.com',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.password}
					label="Password"
					name={`${name}.url_fields.password`}
					required
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.rooms}
					label="Rooms"
					name={`${name}.params.rooms`}
					tooltip={{
						content: 'e.g. !ROOM_ID,ALIAS',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.disabletls}
					label="Disable TLS"
					name={`${name}.params.disabletls`}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default MATRIX;
