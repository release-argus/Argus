import {
	DeployedVersionLookupType,
	DeployedVersionLookupURLType,
	ServiceOptionsType,
} from 'types/config';
import { FC, memo, useEffect, useMemo } from 'react';
import {
	FormCheck,
	FormKeyValMap,
	FormLabel,
	FormSelect,
	FormText,
	FormTextArea,
} from 'components/generic/form';
import { FormGroup, Row } from 'react-bootstrap';
import { useFormContext, useWatch } from 'react-hook-form';

import { BooleanWithDefault } from 'components/generic';
import { DeployedVersionLookupEditType } from 'types/service-edit';
import VersionWithLink from './version-with-link';
import VersionWithRefresh from './version-with-refresh';

const DeployedVersionMethodOptions = [
	{ label: 'GET', value: 'GET' },
	{ label: 'POST', value: 'POST' },
] as const;

interface Props {
	serviceID: string;
	original?: DeployedVersionLookupEditType;
	original_options?: ServiceOptionsType;
	defaults?: DeployedVersionLookupType;
	hard_defaults?: DeployedVersionLookupType;
}

const DeployedVersionURL: FC<Props> = ({
	serviceID,
	original,
	original_options,
	defaults,
	hard_defaults,
}) => {
	const { setValue } = useFormContext();
	const selectedMethod = useWatch({ name: 'deployed_version.method' });
	const templateToggle = useWatch({ name: 'deployed_version.template_toggle' });

	useEffect(() => {
		if (!templateToggle) {
			setValue('deployed_version.regex_template', '');
			setValue('deployed_version.template_toggle', false);
		}
	}, [templateToggle, setValue]);

	const convertedDefaults = useMemo(
		() => ({
			allow_invalid_certs:
				(defaults as DeployedVersionLookupURLType)?.allow_invalid_certs ??
				(hard_defaults as DeployedVersionLookupURLType)?.allow_invalid_certs,
		}),
		[defaults, hard_defaults],
	);

	return (
		<>
			<FormSelect
				name="deployed_version.method"
				col_sm={6}
				col_lg={2}
				label="Method"
				options={DeployedVersionMethodOptions}
				positionSM="right"
				positionLG="middle"
			/>
			<VersionWithLink
				name="deployed_version.url"
				type="url"
				col_sm={12}
				col_lg={8}
				tooltip="URL to query for the version that's running"
				positionXS="right"
			/>
			<BooleanWithDefault
				name="deployed_version.allow_invalid_certs"
				label="Allow Invalid Certs"
				defaultValue={convertedDefaults.allow_invalid_certs}
			/>
			<FormGroup className="pt-1 mb-2 col-12">
				<FormLabel text="Basic auth credentials" />
				<Row>
					<FormText
						key="username"
						name="deployed_version.basic_auth.username"
						col_xs={6}
						label="Username"
					/>
					<FormText
						key="password"
						name="deployed_version.basic_auth.password"
						col_xs={6}
						label="Password"
						positionXS="right"
					/>
				</Row>
			</FormGroup>
			<FormKeyValMap name="deployed_version.headers" />
			{selectedMethod === 'POST' && (
				<FormTextArea
					name="deployed_version.body"
					col_sm={12}
					rows={3}
					label="Body"
					tooltip="Body to send with this request"
				/>
			)}
			<FormText
				name="deployed_version.target_header"
				col_sm={12}
				col_lg={6}
				label="Target header"
				tooltip="Ignore the body and retrieve the version from this header in the response?"
			/>
			<FormText
				name="deployed_version.json"
				col_xs={6}
				label="JSON"
				tooltip={
					<>
						If the URL gives JSON, take the var at this location. e.g.{' '}
						<span className="bold-underline">data.version</span>
					</>
				}
				tooltipAriaLabel="If the URL gives JSON, take the var at this location. e.g. data.version"
				positionXS="left"
				positionLG="right"
			/>
			<FormText
				name="deployed_version.regex"
				required={templateToggle ? 'Required for template' : undefined}
				col_xs={4}
				col_sm={5}
				col_lg={templateToggle ? 6 : 11}
				label="RegEx"
				tooltip={
					<>
						RegEx to extract the version from the URL, e.g.{' '}
						<span className="bold-underline">v([0-9.]+)</span>
					</>
				}
				tooltipAriaLabel="RegEx to extract the version from the URL, e.g. v([0-9.]+)"
				isRegex
				positionXS="middle"
				positionLG="left"
			/>
			{templateToggle && (
				<FormText
					name="deployed_version.regex_template"
					className="order-2 order-lg-1"
					col_sm={12}
					col_lg={5}
					label="RegEx Template"
					tooltip="e.g. RegEx of 'v(\d)-(\d)-(\d)' on 'v4-0-1' with template '$1.$2.$3' would give '4.0.1'"
					positionXS="middle"
				/>
			)}
			<FormCheck
				name={`deployed_version.template_toggle`}
				className="order-1 order-lg-2"
				col_sm={1}
				col_xs={2}
				size="lg"
				label="T"
				smallLabel
				tooltip="Use the RegEx to create a template"
				positionXS="right"
			/>
			<VersionWithRefresh
				className="order-3"
				vType={1}
				serviceID={serviceID}
				original={original}
				original_options={original_options}
			/>
		</>
	);
};

export default memo(DeployedVersionURL);
