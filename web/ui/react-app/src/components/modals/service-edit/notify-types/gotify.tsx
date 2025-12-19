import { useMemo } from 'react';
import { BooleanWithDefault } from '@/components/generic';
import { FieldText } from '@/components/generic/field';
import { GotifyExtras } from '@/components/modals/service-edit/notify-types/extra';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyGotifySchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `Gotify` notifier.
 *
 * @param name - The path to this `Gotify` in the form.
 * @param main - The main values.
 */
const GOTIFY = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyGotifySchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.gotify,
		[main, typeDataDefaults?.notify.gotify],
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
						content: 'e.g. gotify.example.com',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ xs: 3 }}
					defaultVal={defaults?.url_fields?.port}
					label="Port"
					name={`${name}.url_fields.port`}
					tooltip={{
						content: 'e.g. 443',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.path}
					label="Path"
					name={`${name}.url_fields.path`}
					tooltip={{
						ariaLabel: 'Format: gotify.example.io/PATH',
						content: (
							<>
								<span className="text-muted-foreground">
									{'e.g. gotify.example.io/'}
								</span>
								<span className="bold underline">path</span>
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
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldText
					colSize={{ sm: 2 }}
					defaultVal={defaults?.params?.priority}
					label="Priority"
					name={`${name}.params.priority`}
				/>
				<FieldText
					colSize={{ sm: 10 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<GotifyExtras
					defaults={defaults?.params?.extras}
					label="Extras"
					name={`${name}.params.extras`}
					tooltip={{
						content:
							'Additional metadata in the notification payload sent to Gotify',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.disabletls}
					label="Disable TLS"
					name={`${name}.params.disabletls`}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.disabletls}
					label="Insecure Skip Verify"
					name={`${name}.params.insecureskipverify`}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.useheader}
					label="Use Header-based Authentication"
					name={`${name}.params.useheader`}
				/>
			</FieldSet>
		</FieldSet>
	);
};

export default GOTIFY;
