import {
	FormKeyValMap,
	FormLabel,
	FormSelect,
	FormText,
} from 'components/generic/form';
import { NotifyGenericRequestMethods, NotifyGenericType } from 'types/config';
import {
	convertHeadersFromString,
	normaliseForSelect,
} from 'components/modals/service-edit/util';
import { firstNonDefault, strToBool } from 'utils';

import { BooleanWithDefault } from 'components/generic';
import NotifyOptions from 'components/modals/service-edit/notify-types/shared';
import { useMemo } from 'react';
import { useWatch } from 'react-hook-form';

const GenericRequestMethodOptions: {
	label: NotifyGenericRequestMethods;
	value: NotifyGenericRequestMethods;
}[] = (
	[
		'OPTIONS',
		'GET',
		'HEAD',
		'POST',
		'PUT',
		'DELETE',
		'TRACE',
		'CONNECT',
	] as const
).map((method) => ({ label: method, value: method }));

/**
 * The generic form fields shared by many notifiers.
 *
 * @param name - The path to this `Generic` in the form.
 * @param main - The main values.
 * @param defaults - The default values.
 * @param hard_defaults - The hard default values.
 * @returns The form fields for this `Generic` notifier.
 */
const GENERIC = ({
	name,

	main,
	defaults,
	hard_defaults,
}: {
	name: string;

	main?: NotifyGenericType;
	defaults?: NotifyGenericType;
	hard_defaults?: NotifyGenericType;
}) => {
	const selectedTemplate: string | undefined = useWatch({
		name: `${name}.params.template`,
	});

	const convertedDefaults = useMemo(
		() => ({
			// URL Fields
			url_fields: {
				custom_headers: convertHeadersFromString(
					firstNonDefault(
						main?.url_fields?.custom_headers,
						defaults?.url_fields?.custom_headers,
						hard_defaults?.url_fields?.custom_headers,
					),
				),
				host: firstNonDefault(
					main?.url_fields?.host,
					defaults?.url_fields?.host,
					hard_defaults?.url_fields?.host,
				),
				json_payload_vars: convertHeadersFromString(
					firstNonDefault(
						main?.url_fields?.json_payload_vars,
						defaults?.url_fields?.json_payload_vars,
						hard_defaults?.url_fields?.json_payload_vars,
					),
				),
				path: firstNonDefault(
					main?.url_fields?.path,
					defaults?.url_fields?.path,
					hard_defaults?.url_fields?.path,
				),
				port: firstNonDefault(
					main?.url_fields?.port,
					defaults?.url_fields?.port,
					hard_defaults?.url_fields?.port,
				),
				query_vars: convertHeadersFromString(
					firstNonDefault(
						main?.url_fields?.query_vars,
						defaults?.url_fields?.query_vars,
						hard_defaults?.url_fields?.query_vars,
					),
				),
			},
			// Params
			params: {
				contenttype: firstNonDefault(
					main?.params?.contenttype,
					defaults?.params?.contenttype,
					hard_defaults?.params?.contenttype,
				),
				disabletls:
					strToBool(
						firstNonDefault(
							main?.params?.disabletls,
							defaults?.params?.disabletls,
							hard_defaults?.params?.disabletls,
						),
					) ?? true,
				messagekey: firstNonDefault(
					main?.params?.messagekey,
					defaults?.params?.messagekey,
					hard_defaults?.params?.messagekey,
				),
				requestmethod: firstNonDefault(
					main?.params?.requestmethod,
					defaults?.params?.requestmethod,
					hard_defaults?.params?.requestmethod,
				).toLowerCase(),
				template: firstNonDefault(
					main?.params?.template,
					defaults?.params?.template,
					hard_defaults?.params?.template,
				),
				title: firstNonDefault(
					main?.params?.title,
					defaults?.params?.title,
					hard_defaults?.params?.title,
				),
				titlekey: firstNonDefault(
					main?.params?.titlekey,
					defaults?.params?.titlekey,
					hard_defaults?.params?.titlekey,
				),
			},
		}),
		[main, defaults, hard_defaults],
	);

	const genericRequestMethodOptions: { label: string; value: string }[] =
		useMemo(() => {
			const defaultRequestMethod = normaliseForSelect(
				GenericRequestMethodOptions,
				convertedDefaults.params.requestmethod,
			);

			if (defaultRequestMethod)
				return [
					{ value: '', label: `${defaultRequestMethod.label} (default)` },
					...GenericRequestMethodOptions,
				];

			return GenericRequestMethodOptions;
		}, [convertedDefaults.params.requestmethod]);

	return (
		<>
			<NotifyOptions
				name={name}
				main={main?.options}
				defaults={defaults?.options}
				hard_defaults={hard_defaults?.options}
			/>
			<>
				<FormLabel text="URL Fields" heading />
				<>
					<FormText
						name={`${name}.url_fields.host`}
						required
						col_sm={9}
						label="Host"
						tooltip="e.g. gotify.example.com"
						defaultVal={convertedDefaults.url_fields.host}
					/>
					<FormText
						name={`${name}.url_fields.port`}
						col_sm={3}
						label="Port"
						tooltip="e.g. 443"
						isNumber
						defaultVal={convertedDefaults.url_fields.port}
						position="right"
					/>
					<FormText
						name={`${name}.url_fields.path`}
						col_sm={6}
						label="Path"
						tooltip={
							<>
								{'e.g. mattermost.example.io/'}
								<span className="bold-underline">path</span>
							</>
						}
						tooltipAriaLabel="Format: mattermost.example.io/PATH"
						defaultVal={convertedDefaults.url_fields.path}
					/>
					<FormKeyValMap
						name={`${name}.url_fields.custom_headers`}
						tooltip="Additional HTTP headers"
						defaults={convertedDefaults.url_fields.custom_headers}
					/>
					{selectedTemplate && (
						<FormKeyValMap
							name={`${name}.url_fields.json_payload_vars`}
							label="JSON Payload vars"
							tooltip="Override 'title' and 'message' with 'titleKey' and 'messageKey' respectively"
							keyPlaceholder="e.g. key"
							valuePlaceholder="e.g. value"
							defaults={convertedDefaults.url_fields.json_payload_vars}
						/>
					)}
					<FormKeyValMap
						name={`${name}.url_fields.query_vars`}
						label="Query vars"
						tooltip="If you need to pass a query variable that is reserved, you can prefix it with an underscore"
						keyPlaceholder="e.g. foo"
						valuePlaceholder="e.g. bar"
						defaults={convertedDefaults.url_fields.query_vars}
					/>
				</>
				<FormLabel text="Params" heading />
				<>
					<FormSelect
						name={`${name}.params.requestmethod`}
						col_sm={4}
						label="Request Method"
						tooltip="The HTTP request method"
						options={genericRequestMethodOptions}
					/>
					<FormText
						name={`${name}.params.contenttype`}
						col_sm={4}
						label="Content Type"
						tooltip="The value of the Content-Type header"
						defaultVal={convertedDefaults.params.contenttype}
						position="right"
					/>
					<FormText
						name={`${name}.params.template`}
						col_sm={4}
						type="text"
						label="Template"
						tooltip="The template used for creating the request payload"
						defaultVal={convertedDefaults.params.template}
						position="right"
					/>
					<FormText
						name={`${name}.params.messagekey`}
						col_sm={6}
						type="text"
						label="Message Key"
						tooltip="The key that will be used for the message value"
						defaultVal={convertedDefaults.params.messagekey}
					/>
					<FormText
						name={`${name}.params.titlekey`}
						col_sm={6}
						type="text"
						label="Title Key"
						tooltip="The key that will be used for the title value"
						defaultVal={convertedDefaults.params.titlekey}
						position="right"
					/>
					<FormText
						name={`${name}.params.title`}
						col_sm={12}
						type="text"
						label="Title"
						tooltip="Text prepended to the message"
						defaultVal={convertedDefaults.params.title}
					/>
					<BooleanWithDefault
						name={`${name}.params.disabletls`}
						label="Disable TLS"
						defaultValue={convertedDefaults.params.disabletls}
					/>
				</>
			</>
		</>
	);
};

export default GENERIC;
