import { useEffect, useMemo } from 'react';
import { useFormContext, useWatch } from 'react-hook-form';
import { BooleanWithDefault } from '@/components/generic';
import {
	FieldKeyValMap,
	FieldSelect,
	FieldText,
} from '@/components/generic/field';
import {
	Heading,
	NotifyOptions,
} from '@/components/modals/service-edit/notify-types/shared';
import { normaliseForSelect } from '@/components/modals/service-edit/util';
import { FieldSet } from '@/components/ui/field';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import {
	type GenericRequestMethod,
	genericRequestMethodOptions,
} from '@/utils/api/types/config/notify/generic';
import type { NotifyGenericSchema } from '@/utils/api/types/config-edit/notify/schemas';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';
import { ensureValue } from '@/utils/form-utils';

/**
 * The form fields for a `Generic` notifier.
 *
 * @param name - The path to this `Generic` in the form.
 * @param main - The main values.
 * @returns The form fields for this `Generic` notifier.
 */
const GENERIC = ({
	name,
	main,
}: {
	name: string;
	main?: NotifyGenericSchema;
}) => {
	const { getValues, setValue } = useFormContext();
	const selectedTemplate = useWatch({
		name: `${name}.params.template`,
	}) as string;
	const { typeDataDefaults } = useSchemaContext();
	const defaults = useMemo(
		() => main ?? typeDataDefaults?.notify.generic,
		[main, typeDataDefaults?.notify.generic],
	);

	// Ensure selects have a valid value.
	// biome-ignore lint/correctness/useExhaustiveDependencies: fallback on first load.
	useEffect(() => {
		ensureValue<GenericRequestMethod>({
			defaultValue: defaults?.params?.requestmethod,
			fallback: Object.values(genericRequestMethodOptions)[0].value,
			getValues,
			path: `${name}.params.requestmethod`,
			setValue,
		});
	}, [main]);

	const genericRequestMethodOptionsNormalised = useMemo(() => {
		const defaultRequestMethod = normaliseForSelect(
			genericRequestMethodOptions,
			defaults?.params?.requestmethod,
		);

		if (defaultRequestMethod)
			return [
				{ label: `${defaultRequestMethod.label} (default)`, value: nullString },
				...genericRequestMethodOptions,
			];

		return genericRequestMethodOptions;
	}, [defaults?.params?.requestmethod]);

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
					colSize={{ sm: 6 }}
					defaultVal={defaults?.url_fields?.path}
					label="Path"
					name={`${name}.url_fields.path`}
					tooltip={{
						ariaLabel: 'Format: mattermost.example.io/PATH',
						content: (
							<>
								{'e.g. mattermost.example.io/'}
								<span className="bold-underline">path</span>
							</>
						),
						type: 'element',
					}}
				/>
				<FieldKeyValMap
					defaults={defaults?.url_fields?.custom_headers}
					name={`${name}.url_fields.custom_headers`}
					tooltip={{
						content: 'Additional HTTP headers',
						type: 'string',
					}}
				/>
				{selectedTemplate && (
					<FieldKeyValMap
						defaults={defaults?.url_fields?.json_payload_vars}
						label="JSON Payload vars"
						name={`${name}.url_fields.json_payload_vars`}
						placeholders={{ key: 'e.g. key', value: 'e.g. value' }}
						tooltip={{
							content:
								"Override 'title' and 'message' with 'titleKey' and 'messageKey' respectively",
							type: 'string',
						}}
					/>
				)}
				<FieldKeyValMap
					defaults={defaults?.url_fields?.query_vars}
					label="Query vars"
					name={`${name}.url_fields.query_vars`}
					placeholders={{ key: 'e.g. foo', value: 'e.g. bar' }}
					tooltip={{
						content:
							'If you need to pass a query variable that is reserved, you can prefix it with an underscore',
						type: 'string',
					}}
				/>
			</FieldSet>
			<FieldSet className="col-span-full grid grid-cols-subgrid">
				<Heading title="Params" />
				<FieldSelect
					colSize={{ sm: 4 }}
					label="Request Method"
					name={`${name}.params.requestmethod`}
					options={genericRequestMethodOptionsNormalised}
					tooltip={{
						content: 'The HTTP request method',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 4 }}
					defaultVal={defaults?.params?.contenttype}
					label="Content Type"
					name={`${name}.params.contenttype`}
					tooltip={{
						content: 'The value of the Content-Type header',
						type: 'string',
					}}
				/>
				<FieldText
					colSize={{ sm: 4 }}
					defaultVal={defaults?.params?.template}
					label="Template"
					name={`${name}.params.template`}
					tooltip={{
						content: 'The template used for creating the request payload',
						type: 'string',
					}}
					type="text"
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.params?.messagekey}
					label="Message Key"
					name={`${name}.params.messagekey`}
					tooltip={{
						content: 'The key that will be used for the message value',
						type: 'string',
					}}
					type="text"
				/>
				<FieldText
					colSize={{ sm: 6 }}
					defaultVal={defaults?.params?.titlekey}
					label="Title Key"
					name={`${name}.params.titlekey`}
					tooltip={{
						content: 'The key that will be used for the title value',
						type: 'string',
					}}
					type="text"
				/>
				<FieldText
					colSize={{ sm: 12 }}
					defaultVal={defaults?.params?.title}
					label="Title"
					name={`${name}.params.title`}
					tooltip={{
						content: 'Text prepended to the message',
						type: 'string',
					}}
					type="text"
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

export default GENERIC;
