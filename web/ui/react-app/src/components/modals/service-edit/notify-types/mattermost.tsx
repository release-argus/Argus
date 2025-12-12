import { useMemo } from 'react';
import { BooleanWithDefault } from '@/components/generic';
import { FieldText, FieldTextWithPreview } from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NotifyMatterMostSchema } from '@/utils/api/types/config-edit/notify/schemas';

/**
 * The form fields for a `MatterMost` notifier.
 *
 * @param name - The path to this `MatterMost` in the form.
 * @param main - The main values.
 */
const MATTERMOST = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyMatterMostSchema;
}) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.mattermost,
		[main, typeDataDefaults?.notify.mattermost],
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
						ariaLabel: 'Format: mattermost.example.io/PATH',
						content: (
							<>
								<span className="text-muted-foreground">
									{'e.g. mattermost.example.io/'}
								</span>
								<span className="bold underline">path</span>
							</>
						),
						type: 'element',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.channel}
					label="Channel"
					name={`${name}.url_fields.channel`}
					tooltip={{
						content: 'e.g. releases',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.username}
					label="Username"
					name={`${name}.url_fields.username`}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.token}
					label="Token"
					name={`${name}.url_fields.token`}
					required
					tooltip={{
						content: 'WebHook token',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldTextWithPreview
					defaultVal={defaults?.params?.icon}
					label="Icon"
					name={`${name}.params.icon`}
					tooltip={{
						content: 'URL of icon to use',
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

export default MATTERMOST;
