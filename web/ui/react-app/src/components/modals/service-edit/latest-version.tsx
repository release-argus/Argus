import { memo, useEffect } from 'react';
import { useFormContext, useWatch } from 'react-hook-form';
import { BooleanWithDefault } from '@/components/generic';
import { FieldSelect, FieldText } from '@/components/generic/field';
import EditServiceLatestVersionRequire from '@/components/modals/service-edit/latest-version-require';
import FormURLCommands from '@/components/modals/service-edit/latest-version-urlcommands';
import VersionWithLink from '@/components/modals/service-edit/version-with-link';
import VersionWithRefresh from '@/components/modals/service-edit/version-with-refresh';
import {
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from '@/components/ui/accordion';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import {
	LATEST_VERSION_LOOKUP_TYPE,
	type LatestVersionLookupType,
	latestVersionLookupTypeOptions,
} from '@/utils/api/types/config/service/latest-version';

/**
 * The `latest_version` form fields.
 */
const EditServiceLatestVersion = () => {
	const name = 'latest_version';
	const urlFieldName = `${name}.url`;
	const { getValues, trigger } = useFormContext();
	const { schemaDataDefaults } = useSchemaContext();

	const latestVersionType = useWatch({
		name: `${name}.type`,
	}) as LatestVersionLookupType;

	// Validate 'name' when the type changes if we have a 'name' value.
	// biome-ignore lint/correctness/useExhaustiveDependencies: getValues stable.
	useEffect(() => {
		if (getValues(urlFieldName)) void trigger(urlFieldName);
	}, [latestVersionType]);

	const defaults = schemaDataDefaults?.latest_version;

	const urlTooltipText =
		latestVersionType === LATEST_VERSION_LOOKUP_TYPE.GITHUB.value
			? 'GitHub repository to query for the latest release version'
			: 'URL to query for the latest version';

	return (
		<AccordionItem value={name}>
			<AccordionTrigger>Latest Version:</AccordionTrigger>
			<AccordionContent className="grid grid-cols-12 gap-2">
				<FieldSelect
					colSize={{ sm: 4, xs: 4 }}
					label="Type"
					name={`${name}.type`}
					options={latestVersionLookupTypeOptions}
				/>
				<VersionWithLink
					colSize={{ sm: 8, xs: 8 }}
					name={urlFieldName}
					required
					tooltip={{
						content: urlTooltipText,
						type: 'string',
					}}
					type={latestVersionType}
				/>
				{latestVersionType === LATEST_VERSION_LOOKUP_TYPE.GITHUB.value ? (
					<>
						<FieldText
							colSize={{ sm: 12 }}
							defaultVal={defaults?.access_token}
							key="access_token"
							label="Access Token"
							name={`${name}.access_token`}
							tooltip={{
								content:
									'GitHub Personal Access Token to handle possible rate limits and/or private repos',
								type: 'string',
							}}
						/>
						<BooleanWithDefault
							defaultValue={defaults?.use_prerelease}
							label="Use pre-releases"
							name={`${name}.use_prerelease`}
							tooltip={{
								content:
									"Include releases marked 'Pre-release' in the latest version check",
								type: 'string',
							}}
						/>
					</>
				) : (
					<BooleanWithDefault
						defaultValue={defaults?.allow_invalid_certs}
						label="Allow Invalid Certs"
						name={`${name}.allow_invalid_certs`}
					/>
				)}
				<FormURLCommands />
				<EditServiceLatestVersionRequire />

				<VersionWithRefresh vType="latest_version" />
			</AccordionContent>
		</AccordionItem>
	);
};

export default memo(EditServiceLatestVersion);
