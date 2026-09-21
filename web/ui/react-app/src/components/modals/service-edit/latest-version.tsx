import { useEffect, useMemo, useRef } from 'react';
import { useFormContext, useFormState, useWatch } from 'react-hook-form';
import { BooleanWithDefault } from '@/components/generic';
import {
	FieldKeyValMap,
	FieldSelect,
	FieldText,
} from '@/components/generic/field';
import EditServiceLatestVersionRequire from '@/components/modals/service-edit/latest-version-require';
import FormURLCommands from '@/components/modals/service-edit/latest-version-urlcommands';
import { withDefaultOption } from '@/components/modals/service-edit/util';
import { canonicalForgeHost } from '@/components/modals/service-edit/util/service-url';
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
	type LatestVersionLookup,
	type LatestVersionLookupForgejo,
	type LatestVersionLookupType,
	latestVersionLookupTypeOptions,
} from '@/utils/api/types/config/service/latest-version';

/**
 * The `latest_version` form fields.
 */
const EditServiceLatestVersion = () => {
	const name = 'latest_version';
	const hostFieldName = `${name}.host`;
	const urlFieldName = `${name}.url`;
	const { getFieldState, getValues, setValue, trigger } = useFormContext();
	const { schemaData, schemaDataDefaults, typeDataDefaults } =
		useSchemaContext();

	const latestVersionType = useWatch({
		name: `${name}.type`,
	}) as LatestVersionLookupType;
	const forgejoHost = useWatch({
		name: hostFieldName,
	}) as string | undefined;

	// Validate 'name' when the type changes if we have a 'name' value.
	// biome-ignore lint/correctness/useExhaustiveDependencies: getValues stable.
	useEffect(() => {
		if (getValues(urlFieldName)) void trigger(urlFieldName);
	}, [latestVersionType]);

	const typeDefaults = typeDataDefaults?.latest_version?.[latestVersionType];
	const defaultType = schemaDataDefaults?.latest_version?.type;

	// The stored lookup, whose values return when the Forgejo-host is put back.
	const saved = schemaData?.latest_version as LatestVersionLookup | undefined;
	const savedForgejo =
		saved?.type === LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value
			? (saved as LatestVersionLookupForgejo)
			: undefined;

	// The instances the defaults configure, labelled '<name> (<url>)' - or by
	// name alone when it is the URL of the instance.
	const hostOptions = useMemo(() => {
		const hosts = typeDefaults?.hosts ?? {};
		return Object.keys(hosts)
			.sort((a, b) => a.localeCompare(b))
			.map((name) => {
				const url = hosts[name]?.url;
				return {
					label:
						url && canonicalForgeHost(url) !== canonicalForgeHost(name)
							? `${name} (${url})`
							: name,
					value: name,
				};
			});
	}, [typeDefaults]);

	// Forgejo, default the host to previous || first default.
	// biome-ignore lint/correctness/useExhaustiveDependencies: setValue stable.
	useEffect(() => {
		if (latestVersionType !== LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value) return;
		if (forgejoHost || hostOptions.length === 0) return;

		setValue(
			hostFieldName,
			(schemaData?.latest_version.type ===
				LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value &&
				schemaData?.latest_version.host) ||
				hostOptions[0].value,
		);
	}, [latestVersionType, forgejoHost, hostOptions]);

	const hostDefaults = useMemo(() => {
		const hosts = typeDefaults?.hosts;
		if (!hosts || !forgejoHost) return undefined;

		return Object.entries(hosts).find(
			([instance]) => instance.toLowerCase() === forgejoHost.toLowerCase(),
		)?.[1];
	}, [forgejoHost, typeDefaults]);

	const tokenField = `${name}.access_token`;
	const certsField = `${name}.allow_invalid_certs`;
	// Only untouched values follow the instance - once typed in, a field is the user's.
	const formState = useFormState({ name: [tokenField, certsField] });
	const tokenUntouched = !getFieldState(tokenField, formState).isDirty;
	const certsUntouched = !getFieldState(certsField, formState).isDirty;

	const forgejoInstance = `${latestVersionType}|${canonicalForgeHost(
		hostDefaults?.url ?? forgejoHost,
	)}`;
	const previousInstance = useRef<string | null>(null);
	// biome-ignore lint/correctness/useExhaustiveDependencies: setValue stable, saved values read on change.
	useEffect(() => {
		if (previousInstance.current === null) {
			previousInstance.current = forgejoInstance;
			return;
		}
		if (previousInstance.current === forgejoInstance) return;

		const leftForgejo = previousInstance.current.startsWith(
			`${LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value}|`,
		);
		previousInstance.current = forgejoInstance;

		// Undirtied writes throughout, so a cleared field stays ours to restore.
		// Clear a stored access_token when the Forgejo instance changes.
		if (latestVersionType !== LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value) {
			if (leftForgejo && tokenUntouched) setValue(tokenField, '');
			return;
		}

		// Restore `access_token` and `allow_invalid_certs` when instance is reverted.
		const backToStored =
			savedForgejo !== undefined &&
			canonicalForgeHost(savedForgejo.host) === canonicalForgeHost(forgejoHost);
		if (tokenUntouched)
			setValue(
				tokenField,
				backToStored ? (savedForgejo.access_token ?? '') : '',
			);
		if (certsUntouched)
			setValue(
				certsField,
				backToStored ? (savedForgejo.allow_invalid_certs ?? null) : null,
			);
	}, [forgejoInstance]);

	// Add default to type options.
	const typeOptions = useMemo(
		() => withDefaultOption(latestVersionLookupTypeOptions, defaultType),
		[defaultType],
	);

	const urlTooltipText =
		{
			[LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value]:
				'Repository to query for the latest release version, e.g. Release-Argus/Argus',
			[LATEST_VERSION_LOOKUP_TYPE.GITHUB.value]:
				'Repository to query for the latest release version, e.g. release-argus/Argus',
			[LATEST_VERSION_LOOKUP_TYPE.URL.value]:
				'URL to query for the latest version',
		}[latestVersionType] ?? 'URL to query for the latest version';

	return (
		<AccordionItem value={name}>
			<AccordionTrigger>Latest Version:</AccordionTrigger>
			<AccordionContent className="grid grid-cols-12 gap-2">
				<FieldSelect
					colSize={{ sm: 4, xs: 4 }}
					label="Type"
					name={`${name}.type`}
					options={typeOptions}
				/>
				<VersionWithLink
					colSize={{ sm: 8, xs: 8 }}
					host={hostDefaults?.url ?? forgejoHost}
					name={urlFieldName}
					required
					tooltip={{
						content: urlTooltipText,
						type: 'string',
					}}
					type={latestVersionType}
				/>
				{latestVersionType === LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value ? (
					<>
						<FieldSelect
							colSize={{ sm: 12 }}
							creatable
							key="host"
							label="Host"
							name={hostFieldName}
							options={hostOptions}
							required
							tooltip={{
								content: 'Forgejo instance to query',
								type: 'string',
							}}
						/>
						<FieldText
							colSize={{ sm: 12 }}
							defaultVal={hostDefaults?.access_token}
							key="access_token"
							label="Access Token"
							name={`${name}.access_token`}
							tooltip={{
								content:
									'Access Token for this instance, to handle possible rate limits and/or private repos',
								type: 'string',
							}}
						/>
						<BooleanWithDefault
							defaultValue={hostDefaults?.allow_invalid_certs}
							label="Allow Invalid Certs"
							name={`${name}.allow_invalid_certs`}
						/>
						<BooleanWithDefault
							defaultValue={typeDefaults?.use_prerelease}
							label="Use pre-releases"
							name={`${name}.use_prerelease`}
							tooltip={{
								content:
									"Include releases marked 'Pre-release' in the latest version check",
								type: 'string',
							}}
						/>
					</>
				) : latestVersionType === LATEST_VERSION_LOOKUP_TYPE.GITHUB.value ? (
					<>
						<FieldText
							colSize={{ sm: 12 }}
							defaultVal={typeDefaults?.access_token}
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
							defaultValue={typeDefaults?.use_prerelease}
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
					<>
						<BooleanWithDefault
							defaultValue={typeDefaults?.allow_invalid_certs}
							label="Allow Invalid Certs"
							name={`${name}.allow_invalid_certs`}
						/>
						<FieldKeyValMap name={`${name}.headers`} />
					</>
				)}
				<FormURLCommands />
				<EditServiceLatestVersionRequire />

				<VersionWithRefresh vType="latest_version" />
			</AccordionContent>
		</AccordionItem>
	);
};

export default EditServiceLatestVersion;
