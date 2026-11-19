import { useMemo } from 'react';
import { FieldText } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyTwilioSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Twilio` notifier.
 *
 * @param name - The path to this `Twilio` in the form.
 * @param main - The main values.
 */
const TWILIO = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyTwilioSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.twilio,
		[main, typeDataDefaults?.notify.twilio],
	);

	return (
		<FieldSet className="col-span-full grid grid-cols-subgrid">
			<NotifyOptions defaults={defaults?.options} name={name} />
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="URL Fields" />
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.accountsid}
					label="Account SID"
					name={`${name}.url_fields.accountsid`}
					required
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.authtoken}
					label="Auth Token"
					name={`${name}.url_fields.authtoken`}
					required
					tooltip={{
						content: 'Twilio Auth Token, or the API key secret with API Key',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.fromnumber}
					label="From"
					name={`${name}.url_fields.fromnumber`}
					required
					tooltip={{
						content:
							'Sender phone number, alphanumeric sender ID, or Messaging Service SID',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.tonumbers}
					label="To"
					name={`${name}.url_fields.tonumbers`}
					required
					tooltip={{
						content: 'Comma-separated recipient phone numbers',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					defaultVal={defaults?.params?.apikey}
					label="API Key"
					name={`${name}.params.apikey`}
					tooltip={{
						content:
							'API key SID, which authenticates with the Auth Token as its secret',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default TWILIO;
