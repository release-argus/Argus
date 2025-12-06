import { useMutation, useQueryClient } from '@tanstack/react-query';
import { AlertCircleIcon, Loader2Icon, RefreshCw } from 'lucide-react';
import { type FC, useState } from 'react';
import { useFormContext, useWatch } from 'react-hook-form';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useSchemaContext } from '@/contexts/service-edit-zod-type';
import { useErrors } from '@/hooks/use-error';
import useValuesRefetch from '@/hooks/values-refetch';
import { QUERY_KEYS } from '@/lib/query-keys.ts';
import { cn } from '@/lib/utils';
import { beautifyGoErrors, removeEmptyValues } from '@/utils';
import { mapRequest } from '@/utils/api/types/api-request-handler';
import type { ServiceSummary } from '@/utils/api/types/config/summary.ts';
import {
	type URLCommandsSchema,
	urlCommandsSchemaOutgoing,
} from '@/utils/api/types/config-edit/service/types/latest-version';

/* The throttle time for refreshing the version. */
const TEST_THROTTLE_MS = 1000;

type VersionWithRefreshProps = {
	/* The version to display/fetch. */
	vType: 'latest_version' | 'deployed_version';
	/* Optional additional CSS class names for styling the component. */
	className?: string;
};

/**
 * The version with a button to refetch.
 *
 * @param vType - latest_version/deployed_version.
 * @param className - Extra class names for the component.
 * @returns The current version with a button to trigger a refetch.
 */
const VersionWithRefresh: FC<VersionWithRefreshProps> = ({
	vType,
	className,
}) => {
	const queryClient = useQueryClient();
	const { clearErrors, setValue, trigger } = useFormContext();
	const {
		serviceID,
		schema: schemaFull,
		schemaData: originalData,
	} = useSchemaContext();

	const [lastFetched, setLastFetched] = useState(0);
	const [queryError, setQueryError] = useState<string | null>(null);

	const schema = schemaFull.shape[vType];

	const url = useWatch({ name: `${vType}.url` }) as string;

	const vTypeErrors = useErrors(vType, true);
	const { data, refetchData } = useValuesRefetch(vType);
	const { data: semanticVersioning, refetchData: refetchSemanticVersioning } =
		useValuesRefetch<boolean | null>('options.semantic_versioning');

	const {
		mutate: fetchVersion,
		data: versionData,
		isPending: isFetching,
	} = useMutation({
		mutationFn: async () => {
			if (originalData === null) return null;

			const dataConverted = schema.parse(data);
			// Trim URL Commands.
			if (vType === 'latest_version' && 'url_commands' in dataConverted) {
				const parsedURLCommands = urlCommandsSchemaOutgoing.parse(
					dataConverted.url_commands,
				);
				const trimmedURLCommands = parsedURLCommands
					? parsedURLCommands.map((urlC) => removeEmptyValues(urlC))
					: null;
				dataConverted.url_commands = trimmedURLCommands as URLCommandsSchema;
			}

			return await mapRequest('VERSION_REFRESH', {
				data: dataConverted,
				dataSemanticVersioning: semanticVersioning ?? null,
				dataTarget: vType,
				original: originalData[vType],
				originalSemanticVersioning:
					originalData?.options?.semantic_versioning ?? null,
				serviceID: serviceID,
			});
		},
		onError: (error) => {
			setValue(`${vType}.version`, '');
			setQueryError(
				beautifyGoErrors(
					error instanceof Error ? error.message : String(error),
				),
			);
		},
		onMutate: () => {
			// Reset errors and version field before fetching.
			setQueryError(null);
			setValue(`${vType}.version`, '');
		},
		onSuccess: (data) => {
			if (data?.version) {
				setValue(`${vType}.version`, data.version);
				clearErrors(`${vType}.version`);
			}
		},
	});

	const versionFallback = serviceID
		? queryClient.getQueryData<ServiceSummary>(
				QUERY_KEYS.SERVICE.SUMMARY_ITEM(serviceID),
			)?.status?.[vType]
		: null;
	const version =
		(useWatch({ name: `${vType}.version` }) as string | null | undefined) ??
		versionFallback;

	// Refetch the version.
	const refetch = async () => {
		// Prevent refetching too often.
		const currentTime = Date.now();
		if (currentTime - lastFetched < TEST_THROTTLE_MS) return undefined;

		// Ensure valid form.
		const result = await trigger(vType, { shouldFocus: true });
		if (!result) return undefined;

		refetchSemanticVersioning();
		refetchData();
		// setTimeout to allow time for refetch setStates ^
		const timeout = setTimeout(() => {
			if (url) {
				fetchVersion();
				setLastFetched(currentTime);
			}
		});
		return () => clearTimeout(timeout);
	};

	return (
		<span className={cn('col-span-full', className)}>
			<span className="flex pt-1 pb-2">
				{vType === 'latest_version' ? 'Latest' : 'Deployed'} version: {version}
				{isFetching && <Skeleton className="ml-1 h-4 w-16 max-w-full" />}
				<Button
					aria-label="Refresh the version"
					className="ml-auto"
					disabled={isFetching || !url}
					onClick={refetch}
					variant="secondary"
				>
					{isFetching ? (
						<Loader2Icon className="animate-spin" />
					) : (
						<RefreshCw />
					)}
					Refresh
				</Button>
			</span>
			{(queryError ?? versionData?.message) && (
				<Alert variant="destructive">
					<AlertCircleIcon />
					<AlertTitle>Failed to refresh:</AlertTitle>
					<AlertDescription>
						{queryError ?? beautifyGoErrors(versionData?.message ?? '')}
					</AlertDescription>
				</Alert>
			)}
			{vTypeErrors && (
				<Alert className="mb-0 pl-8" variant="destructive">
					<ul className="col-span-full list-inside list-disc space-y-1">
						{Object.entries(vTypeErrors).map(([key, error]) => (
							<li key={key}>
								{key}: {error}
							</li>
						))}
					</ul>
				</Alert>
			)}
		</span>
	);
};

export default VersionWithRefresh;
