import { useMemo } from 'react';
import { BooleanWithDefault } from '@/components/generic';
import {
	FieldSelect,
	FieldText,
	FieldTextWithPreview,
} from '@/components/generic/field';
import { NtfyActions } from '@/components/modals/service-edit/notify-types/extra';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { normaliseForSelect } from '@/components/modals/service-edit/util';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import {
	ntfyPriorityOptions,
	ntfySchemeOptions,
} from '@/utils/api/types/config/notify/ntfy';
import type { NotifyNtfySchema } from '@/utils/api/types/config-edit/notify/schemas';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';
import { applyDefaultsRecursive } from '@/utils/api/types/config-edit/util';

/**
 * The form fields for a `NTFY` notifier.
 *
 * @param name - The path to this `NTFY` in the form.
 * @param main - The main values.
 */
const NTFY = ({ name, main }: { name: string; main?: NotifyNtfySchema }) => {
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => applyDefaultsRecursive(main ?? null, typeDataDefaults?.notify.ntfy),
		[main, typeDataDefaults?.notify.ntfy],
	);

	const ntfyPriorityOptionsNormalised = useMemo(() => {
		const defaultPriority = normaliseForSelect(
			ntfyPriorityOptions,
			defaults?.params?.priority,
		);

		if (defaultPriority)
			return [
				{ label: `${defaultPriority.label} (default)`, value: nullString },
				...ntfyPriorityOptions,
			];

		return ntfyPriorityOptions;
	}, [defaults?.params?.priority]);

	const ntfySchemeOptionsNormalised = useMemo(() => {
		const defaultScheme = normaliseForSelect(
			ntfySchemeOptions,
			defaults?.params?.scheme,
		);

		if (defaultScheme)
			return [
				{ label: `${defaultScheme.label} (default)`, value: nullString },
				...ntfySchemeOptions,
			];

		return ntfySchemeOptions;
	}, [defaults?.params?.scheme]);

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
						content: 'e.g. ntfy.example.com',
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
					defaultVal={defaults?.url_fields?.username}
					label="Username"
					name={`${name}.url_fields.username`}
				/>
				<FieldText
					defaultVal={defaults?.url_fields?.password}
					label="Password"
					name={`${name}.url_fields.password`}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.url_fields?.topic}
					label="Topic"
					name={`${name}.url_fields.topic`}
					required
					tooltip={{
						content: 'Target topic',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldSelect
					colSize={{ lg: 3, sm: 6 }}
					label="Scheme"
					name={`${name}.params.scheme`}
					options={ntfySchemeOptionsNormalised}
					tooltip={{
						content: 'Server protocol',
						type: 'string',
					}}
				/>
				<FieldSelect
					colSize={{ lg: 3, sm: 6 }}
					label="Priority"
					name={`${name}.params.priority`}
					options={ntfyPriorityOptionsNormalised}
				/>
				<FieldText
					colSize={{ lg: 6, sm: 12 }}
					defaultVal={defaults?.params?.tags}
					label="Tags"
					name={`${name}.params.tags`}
					tooltip={{
						content:
							'Comma-separated list of tags that may or may not map to emojis',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 8 }}
					defaultVal={defaults?.params?.attach}
					label="Attach"
					name={`${name}.params.attach`}
					tooltip={{
						content: 'URL of an attachment',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 4 }}
					defaultVal={defaults?.params?.filename}
					label="Filename"
					name={`${name}.params.filename`}
					tooltip={{
						content: 'File name of the attachment',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.email}
					label="E-mail"
					name={`${name}.params.email`}
					tooltip={{
						content: 'E-mail address to send to',
						type: 'string',
					}}
				/>
				<FieldText
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.click}
					label="Click"
					name={`${name}.params.click`}
					tooltip={{
						content: 'URL to open when notification is clicked',
						type: 'string',
					}}
				/>
				<FieldTextWithPreview
					defaultVal={defaults?.params?.icon}
					label="Icon"
					name={`${name}.params.icon`}
					tooltip={{
						content: 'URL to an icon',
						type: 'string',
					}}
				/>
				<NtfyActions
					defaults={defaults?.params?.actions}
					label="Actions"
					name={`${name}.params.actions`}
					tooltip={{
						content: 'Custom action buttons for notifications',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.cache}
					label="Cache"
					name={`${name}.params.cache`}
					tooltip={{
						content: 'Cache messages',
						type: 'string',
					}}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.disabletls}
					label="Disable TLS"
					name={`${name}.params.disabletls`}
				/>
				<BooleanWithDefault
					defaultValue={defaults?.params?.firebase}
					label="Firebase"
					name={`${name}.params.firebase`}
					tooltip={{
						content: 'Send to Firebase Cloud Messaging',
						type: 'string',
					}}
				/>
			</FieldSet>
		</FieldSet>
	);
};
export default NTFY;
