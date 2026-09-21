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
import {
	canonicalForgeHost,
	forgeHostEntry,
	resolveForgeHost,
} from '@/components/modals/service-edit/util/service-url';
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

	// The stored lookup, whose values return when its instance is addressed again.
	const saved = schemaData?.latest_version as LatestVersionLookup | undefined;
	const savedShared = saved as
		| { access_token?: string; allow_invalid_certs?: boolean | null }
		| undefined;
	const savedForgejoHost =
		saved?.type === LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value
			? (saved as LatestVersionLookupForgejo).host
			: undefined;

	// The instances the defaults configure, labelled '<name> (<url>)' - or by
	// name alone when it is the URL of the instance.
	const hostOptions = useMemo(() => {
		const hosts = typeDefaults?.hosts ?? {};
		const canonicalHostCounts = new Map<string, number>();

		for (const { url } of Object.values(hosts)) {
			if (!url) continue;

			const canonicalHost = canonicalForgeHost(url);
			canonicalHostCounts.set(
				canonicalHost,
				(canonicalHostCounts.get(canonicalHost) ?? 0) + 1,
			);
		}

		// `name` when canonical url is unique AND canonical(name) === canonical(url)
		// `name (canonical(url))` otherwise
		return Object.entries(hosts)
			.sort(([a], [b]) => a.localeCompare(b))
			.map(([name, host]) => {
				const url = host?.url;
				const hostCanonical = url && canonicalForgeHost(url);
				const isUnique =
					hostCanonical !== undefined &&
					canonicalHostCounts.get(hostCanonical) === 1;

				return {
					label:
						hostCanonical === canonicalForgeHost(name) && isUnique
							? name
							: `${name} (${canonicalForgeHost(url)})`,
					value: name,
				};
			});
	}, [typeDefaults]);

	// Forgejo, default the host to previous || first default.
	// biome-ignore lint/correctness/useExhaustiveDependencies: setValue stable.
	useEffect(() => {
		if (latestVersionType !== LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value) return;
		if (forgejoHost || hostOptions.length === 0) return;

		setValue(hostFieldName, savedForgejoHost || hostOptions[0].value);
	}, [latestVersionType, forgejoHost, hostOptions]);

	const hostDefaults = useMemo(
		() => forgeHostEntry(forgejoHost, typeDefaults?.hosts),
		[forgejoHost, typeDefaults],
	);

	const tokenField = `${name}.access_token`;
	const certsField = `${name}.allow_invalid_certs`;
	// Only untouched values follow the instance - once typed in, a field is the user's.
	const formState = useFormState({ name: [tokenField, certsField] });
	const tokenUntouched = !getFieldState(tokenField, formState).isDirty;
	const certsUntouched = !getFieldState(certsField, formState).isDirty;

	const forgejoHosts =
		typeDataDefaults?.latest_version?.[LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value]
			?.hosts;

	// The instance a lookup addresses.
	const instanceKey = (
		type: LatestVersionLookupType | null | undefined,
		host: string | undefined,
	) =>
		type === LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value
			? `${type}|${resolveForgeHost(host ?? '', forgejoHosts)}`
			: String(type);

	const currentInstance = instanceKey(latestVersionType, forgejoHost);
	const storedInstance = instanceKey(saved?.type, savedForgejoHost);
	const previousInstance = useRef<string | null>(null);
	// biome-ignore lint/correctness/useExhaustiveDependencies: setValue stable, stored values read on change.
	useEffect(() => {
		if (previousInstance.current === null) {
			previousInstance.current = currentInstance;
			return;
		}
		if (previousInstance.current === currentInstance) return;
		previousInstance.current = currentInstance;

		// Undirtied writes throughout, so a cleared field stays ours to restore.
		const backToStored = currentInstance === storedInstance;
		if (tokenUntouched)
			setValue(
				tokenField,
				backToStored ? (savedShared?.access_token ?? '') : '',
			);
		if (certsUntouched)
			setValue(
				certsField,
				backToStored ? (savedShared?.allow_invalid_certs ?? null) : null,
			);
	}, [currentInstance]);

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
		}[latestVersionType] ?? '';

	const typeFields = () => {
		switch (latestVersionType) {
			case LATEST_VERSION_LOOKUP_TYPE.FORGEJO.value:
				return (
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
				);
			case LATEST_VERSION_LOOKUP_TYPE.GITHUB.value:
				return (
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
				);
			default:
				return (
					<>
						<BooleanWithDefault
							defaultValue={typeDefaults?.allow_invalid_certs}
							label="Allow Invalid Certs"
							name={`${name}.allow_invalid_certs`}
						/>
						<FieldKeyValMap name={`${name}.headers`} />
					</>
				);
		}
	};

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
				{typeFields()}
				<FormURLCommands />
				<EditServiceLatestVersionRequire />

				<VersionWithRefresh vType="latest_version" />
			</AccordionContent>
		</AccordionItem>
	);
};

export default EditServiceLatestVersion;
