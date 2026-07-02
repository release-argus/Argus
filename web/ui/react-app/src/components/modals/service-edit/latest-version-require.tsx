import { useMemo } from 'react';
import { useWatch } from 'react-hook-form';
import { HelpTooltip } from '@/components/generic';
import { FieldLabel, FieldSelect, FieldText } from '@/components/generic/field';
import type { TooltipWithAriaProps } from '@/components/generic/tooltip';
import Command from '@/components/modals/service-edit/command';
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from '@/components/ui/accordion';
import { FieldLegend, FieldSet } from '@/components/ui/field';
import { Separator } from '@/components/ui/separator';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import type { NonNull } from '@/types/util';
import {
	type DockerFilterType,
	type DockerFilterUsernameToken,
	type DockerType,
	LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE,
	LATEST_VERSION_LOOKUP_TYPE,
	type LatestVersionLookupType,
	latestVersionRequireDockerTypeOptions,
	type RequireDockerFilterDefaults,
} from '@/utils/api/types/config/service/latest-version';
import type { LatestVersionRequire } from '@/utils/api/types/config-edit/service/types/latest-version';
import {
	type NullString,
	nullString,
} from '@/utils/api/types/config-edit/shared/null-string';

// Widest auth shape across the registry union (ECR has none).
type WithDockerAuth = { auth?: DockerFilterUsernameToken['auth'] | null };

/**
 * The `latest_version.require` form fields.
 */
const EditServiceLatestVersionRequire = () => {
	const name = 'latest_version.require';
	const { schemaDataDefaults } = useSchemaContext();

	const values = useWatch({ name: name }) as LatestVersionRequire;

	const defaults = schemaDataDefaults?.latest_version?.require;
	const defaultDockerRegistry: RequireDockerFilterDefaults['type'] =
		defaults?.docker?.type ?? nullString;

	// Add default to docker registry options.
	const dockerRegistryOptions = useMemo(() => {
		// No default.
		if (defaultDockerRegistry === nullString)
			return latestVersionRequireDockerTypeOptions;

		// Find default value.
		const defaultLower = defaultDockerRegistry.toLowerCase();
		const defaultDockerRegistryLabel =
			latestVersionRequireDockerTypeOptions.find(
				(option) => option.value.toLowerCase() === defaultLower,
			);

		// Known default registry.
		if (defaultDockerRegistryLabel)
			return [
				{
					label: `${defaultDockerRegistryLabel.label} (default)`,
					value: nullString,
				},
				...latestVersionRequireDockerTypeOptions,
			];

		// Unknown default registry, return without this default.
		return latestVersionRequireDockerTypeOptions;
	}, [defaultDockerRegistry]);

	// Show the 'username' field if 'Docker Hub' type.
	const dockerRegistry = useWatch({
		name: 'latest_version.require.docker.type',
	}) as DockerType | NullString;
	const selectedDockerRegistry =
		dockerRegistry === nullString
			? (defaultDockerRegistry as NonNull<DockerFilterType>)
			: dockerRegistry;
	// Only Docker Hub has a username field.
	const showUsernameField =
		selectedDockerRegistry ===
		LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.DOCKER_HUB.value;
	// Amazon ECR Public Gallery is anonymous with no auth fields.
	const showTokenField =
		selectedDockerRegistry !==
		LATEST_VERSION_LOOKUP__REQUIRE_DOCKER_TYPE.AMAZON_ECR.value;
	const dockerDefaults = defaults?.docker?.registry?.[selectedDockerRegistry];
	// `auth` varies by registry (token / token+username / none for ECR), so read
	// it through the widest shape — every field is optional.
	const dockerAuth = (values.docker as WithDockerAuth)?.auth;
	const dockerDefaultsAuth = (dockerDefaults as WithDockerAuth | undefined)?.auth;

	// Target release assets or webpages.
	const latestVersionType = useWatch({
		name: 'latest_version.type',
	}) as NonNullable<LatestVersionLookupType>;
	const tooltipRegexContent: TooltipWithAriaProps = {
		content:
			latestVersionType === LATEST_VERSION_LOOKUP_TYPE.GITHUB.value
				? 'Release assets must contain a match'
				: 'Webpage must contain a match',
		type: 'string',
	};
	const hasContainer =
		!!values.docker.image.trim() &&
		(!!values.docker.tag.trim() || !!dockerDefaults?.tag?.trim());

	return (
		<Accordion className="col-span-full" collapsible type="single">
			<AccordionItem value="require">
				<AccordionTrigger>Require:</AccordionTrigger>
				<AccordionContent className="grid grid-cols-12 gap-2">
					<FieldText
						colSize={{ xs: 6 }}
						label={'RegEx Content'}
						name={`${name}.regex_content`}
						tooltip={tooltipRegexContent}
					/>
					<FieldText
						colSize={{ xs: 6 }}
						label={'RegEx Version'}
						name={`${name}.regex_version`}
						tooltip={{
							content:
								"Version found must match, e.g. exclude '*-beta' versions with '^[0-9.]+$'",
							type: 'string',
						}}
					/>

					<div className="col-span-full grid grid-cols-subgrid gap-3">
						<FieldLabel
							text="Command"
							tooltip={{
								content:
									'Command to run before a release is considered usable and notified/shown in the UI',
								type: 'string',
							}}
						/>
						<Command name={`${name}.command`} />
					</div>

					<Separator className="col-span-full my-4" />
					<FieldSet className="col-span-full grid grid-cols-subgrid gap-2">
						<FieldLegend className="flex flex-row items-center">
							Docker
							<HelpTooltip
								content="Docker image requirements for the version to be considered usable"
								delayDuration={500}
								type="string"
							/>
						</FieldLegend>
						<FieldSelect
							colSize={{ sm: 12 }}
							label="Type"
							name={`${name}.docker.type`}
							options={dockerRegistryOptions}
						/>
						<FieldText
							colSize={{ xs: 6 }}
							label="Image"
							name={`${name}.docker.image`}
							required={values?.docker?.tag}
						/>
						<FieldText
							colSize={{ xs: 6 }}
							defaultVal={dockerDefaults?.tag}
							label="Tag"
							name={`${name}.docker.tag`}
							required={values?.docker?.image}
						/>
						{showUsernameField && (
							<FieldText
								colSize={{ sm: 4 }}
								defaultVal={dockerDefaultsAuth?.username}
								key="username"
								label="Username"
								name={`${name}.docker.auth.username`}
								required={
									hasContainer &&
									(dockerAuth?.token || dockerDefaultsAuth?.token)
								}
							/>
						)}
						{showTokenField && (
							<FieldText
								colSize={{ sm: showUsernameField ? 8 : 12 }}
								defaultVal={dockerDefaultsAuth?.token}
								key="token"
								label="Token"
								name={`${name}.docker.auth.token`}
								required={
									showUsernameField &&
									hasContainer &&
									(dockerAuth?.username || dockerDefaultsAuth?.username)
								}
							/>
						)}
					</FieldSet>
				</AccordionContent>
			</AccordionItem>
		</Accordion>
	);
};

export default EditServiceLatestVersionRequire;
