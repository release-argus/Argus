import { useEffect, useMemo } from 'react';
import { useFormContext, useWatch } from 'react-hook-form';
import { BooleanWithDefault } from '@/components/generic';
import {
	FieldCheck,
	FieldKeyValMap,
	FieldSelect,
	FieldText,
	FieldTextArea,
} from '@/components/generic/field';
import { normaliseForSelect } from '@/components/modals/service-edit/util';
import VersionWithLink from '@/components/modals/service-edit/version-with-link';
import VersionWithRefresh from '@/components/modals/service-edit/version-with-refresh';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import {
	type DeployedVersionLookupURLMethod,
	deployedVersionLookupURLMethodOptions,
} from '@/utils/api/types/config/service/deployed-version';
import type {
	DeployedVersionURLMethod,
	DeployedVersionURLSchema,
} from '@/utils/api/types/config-edit/service/types/deployed-version';
import { nullString } from '@/utils/api/types/config-edit/shared/null-string';
import { ensureValue } from '@/utils/form-utils';

/**
 * The `deployed_version` form fields for 'url' version.
 */
const DeployedVersionURL = () => {
	const name = 'deployed_version';
	const { schemaDataDefaults } = useSchemaContext();
	const { getValues, setValue, trigger } = useFormContext();
	// biome-ignore lint/correctness/useExhaustiveDependencies: deployed_version static.
	const defaults = useMemo(
		() =>
			schemaDataDefaults?.deployed_version as Partial<DeployedVersionURLSchema>,
		[],
	);

	const regexTemplateFormName = `${name}.regex_template`;
	const methodFormName = `${name}.method`;
	const templateToggleFormName = `${name}.template_toggle`;
	// Ensure selects have a valid value.
	// biome-ignore lint/correctness/useExhaustiveDependencies: defaults stable.
	useEffect(() => {
		ensureValue<DeployedVersionLookupURLMethod>({
			defaultValue: defaults?.method,
			fallback: Object.values(deployedVersionLookupURLMethodOptions)[0].value,
			getValues,
			path: methodFormName,
			setValue,
		});
	}, []);

	const selectedMethod = (useWatch({
		name: methodFormName,
	}) || defaults?.method) as DeployedVersionURLMethod;
	const templateToggle = (useWatch({
		name: templateToggleFormName,
	}) ?? false) as boolean;

	const deployedVersionLookupURLMethodNormalised = useMemo(() => {
		const defaultMethod = normaliseForSelect(
			deployedVersionLookupURLMethodOptions,
			defaults?.method,
		);

		if (defaultMethod)
			return [
				{ label: `${defaultMethod.label} (default)`, value: nullString },
				...deployedVersionLookupURLMethodOptions,
			];

		return deployedVersionLookupURLMethodOptions;
	}, [defaults?.method]);

	// biome-ignore lint/correctness/useExhaustiveDependencies: setValue stable
	useEffect(() => {
		if (!templateToggle) {
			setValue(regexTemplateFormName, '');
			setValue(templateToggleFormName, false);
			void trigger(`${name}.regex`);
		}
	}, [templateToggle]);

	return (
		<>
			<FieldSelect
				colSize={{ lg: 2, xs: 6 }}
				label="Method"
				name={methodFormName}
				options={deployedVersionLookupURLMethodNormalised}
			/>
			<VersionWithLink
				colSize={{ lg: 8, sm: 12 }}
				name={`${name}.url`}
				tooltip={{
					content: "URL to query for the version that's running",
					type: 'string',
				}}
				type="url"
			/>
			<BooleanWithDefault
				defaultValue={defaults?.allow_invalid_certs}
				label="Allow Invalid Certs"
				name={`${name}.allow_invalid_certs`}
			/>
			<div className="col-span-full mb-2 grid grid-cols-subgrid pt-1">
				<p className="col-span-full">Basic auth credentials</p>
				<FieldText
					colSize={{ xs: 6 }}
					key="username"
					label="Username"
					name={`${name}.basic_auth.username`}
				/>
				<FieldText
					colSize={{ xs: 6 }}
					key="password"
					label="Password"
					name={`${name}.basic_auth.password`}
				/>
			</div>
			<FieldKeyValMap name={`${name}.headers`} />
			{selectedMethod === 'POST' && (
				<FieldTextArea
					colSize={{ sm: 12 }}
					label="Body"
					name={`${name}.body`}
					rows={3}
					tooltip={{
						content: 'Body to send with this request',
						type: 'string',
					}}
				/>
			)}
			<FieldText
				colSize={{ lg: 6, sm: 12 }}
				label="Target header"
				name={`${name}.target_header`}
				tooltip={{
					content:
						'Ignore the body and retrieve the version from this header in the response?',
					type: 'string',
				}}
			/>
			<FieldText
				colSize={{ xs: 6 }}
				label="JSON"
				name={`${name}.json`}
				tooltip={{
					ariaLabel:
						'If the URL gives JSON, take the var at this location. e.g. data.version',
					content: (
						<>
							If the URL gives JSON, take the var at this location. e.g.{' '}
							<span className="bold underline">data.version</span>
						</>
					),
					type: 'element',
				}}
			/>
			<FieldText
				colSize={{ lg: templateToggle ? 6 : 11, sm: 5, xs: 4 }}
				label="RegEx"
				name={`${name}.regex`}
				required={templateToggle ? 'Required for template' : undefined}
				tooltip={{
					ariaLabel:
						'RegEx to extract the version from the URL, e.g. v([0-9.]+)',
					content: (
						<>
							RegEx to extract the version from the URL, e.g.{' '}
							<span className="bold underline">v([0-9.]+)</span>
						</>
					),
					type: 'element',
				}}
			/>
			{templateToggle && (
				<FieldText
					className="order-2 lg:order-1"
					colSize={{ lg: 5, sm: 12 }}
					label="RegEx Template"
					name={regexTemplateFormName}
					tooltip={{
						content: String.raw`e.g. RegEx of 'v(\d)-(\d)-(\d)' on 'v4-0-1' with template '$1.$2.$3' would give '4.0.1'`,
						type: 'string',
					}}
				/>
			)}
			<FieldCheck
				className="order-1 lg:order-2"
				colSize={{ sm: 1, xs: 2 }}
				label="T"
				name={templateToggleFormName}
				tooltip={{
					content: 'Use the RegEx to create a template',
					type: 'string',
				}}
			/>
			<VersionWithRefresh className="order-3" vType="deployed_version" />
		</>
	);
};

export default DeployedVersionURL;
