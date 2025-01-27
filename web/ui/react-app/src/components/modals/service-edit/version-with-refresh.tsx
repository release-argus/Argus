import { Alert, Button } from 'react-bootstrap';
import {
	DeployedVersionLookupEditType,
	LatestVersionLookupEditType,
	ServiceRefreshType,
} from 'types/service-edit';
import { FC, useMemo, useState } from 'react';
import {
	beautifyGoErrors,
	convertToQueryParams,
	fetchJSON,
	getChanges,
	removeEmptyValues,
} from 'utils';
import {
	convertUIDeployedVersionDataEditToAPI,
	convertUILatestVersionDataEditToAPI,
} from 'components/modals/service-edit/util';
import { faSpinner, faSync } from '@fortawesome/free-solid-svg-icons';
import { useFormContext, useWatch } from 'react-hook-form';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ServiceOptionsType } from 'types/config';
import { useErrors } from 'hooks/errors';
import { useQuery } from '@tanstack/react-query';
import useValuesRefetch from 'hooks/values-refetch';
import { useWebSocket } from 'contexts/websocket';

interface Props {
	vType: 0 | 1; // 0: Latest, 1: Deployed
	serviceID: string;
	original?: LatestVersionLookupEditType | DeployedVersionLookupEditType;
	original_options?: ServiceOptionsType;
}

/**
 * The version with a button to refetch.
 *
 * @param vType - 0: Latest, 1: Deployed.
 * @param serviceID - The ID of the service.
 * @param original - The original values in the form.
 * @param original_options - The original service.options of the form.
 * @returns The current version with a button to trigger a refetch.
 */
const VersionWithRefresh: FC<Props> = ({
	vType,
	serviceID,
	original,
	original_options,
}) => {
	const [lastFetched, setLastFetched] = useState(0);
	const { monitorData } = useWebSocket();
	const { trigger } = useFormContext();
	const dataTarget = vType === 0 ? 'latest_version' : 'deployed_version';
	const convertedOriginal = useMemo(() => {
		if (original === null) return {};
		// Latest version
		if (dataTarget === 'latest_version')
			return convertUILatestVersionDataEditToAPI(
				original as LatestVersionLookupEditType,
			);
		// Deployed version
		return convertUIDeployedVersionDataEditToAPI(
			original as DeployedVersionLookupEditType,
		);
	}, [original, serviceID, dataTarget]);
	const url: string | undefined = useWatch({ name: `${dataTarget}.url` });
	const dataTargetErrors = useErrors(dataTarget, true);
	const { data, refetchData } = useValuesRefetch(dataTarget);
	const { data: semanticVersioning, refetchData: refetchSemanticVersioning } =
		useValuesRefetch('options.semantic_versioning');

	const fetchVersionJSON = () => {
		let semantic_versioning;
		if (
			(semanticVersioning ?? '') !==
			(original_options?.semantic_versioning ?? '')
		) {
			if (semanticVersioning === null) {
				semantic_versioning = 'null';
			} else {
				semantic_versioning = `${semanticVersioning}`;
			}
		}
		const overrides = data
			? getChanges({
					params: data,
					defaults: convertedOriginal,
					target: dataTarget,
				})
			: '';
		return fetchJSON<ServiceRefreshType>({
			url: `api/v1/${vType === 0 ? 'latest' : 'deployed'}_version/refresh${
				serviceID ? `/${encodeURIComponent(serviceID)}` : ''
			}${convertToQueryParams({
				overrides,
				semantic_versioning,
			})}`,
		});
	};

	const {
		data: versionData,
		isFetching,
		refetch: refetchVersion,
	} = useQuery({
		queryKey: [
			'version/refresh',
			dataTarget,
			{ id: serviceID },
			{
				params: removeEmptyValues(data),
				semantic_versioning: semanticVersioning,
				original_data: removeEmptyValues(original ?? []),
			},
		],
		queryFn: () => fetchVersionJSON(),
		enabled: false,
		initialData: {
			version: monitorData.service[serviceID]?.status?.[dataTarget] ?? '',
			error: '',
			timestamp: '',
		},
		notifyOnChangeProps: 'all',
		staleTime: 0,
	});

	const refetch = async () => {
		// Prevent refetching too often.
		const currentTime = Date.now();
		if (currentTime - lastFetched < 1000) return;

		// Ensure valid form.
		const result = await trigger(dataTarget, { shouldFocus: true });
		if (!result) return;

		refetchSemanticVersioning();
		refetchData();
		// setTimeout to allow time for refetch setStates ^
		const timeout = setTimeout(() => {
			if (url) {
				refetchVersion();
				setLastFetched(currentTime);
			}
		});
		return () => clearTimeout(timeout);
	};

	const LoadingSpinner = (
		<FontAwesomeIcon icon={faSpinner} spin style={{ marginLeft: '0.5rem' }} />
	);

	return (
		<span style={{ alignItems: 'center' }}>
			<span className="pt-1 pb-2" style={{ display: 'flex' }}>
				{vType === 0 ? 'Latest' : 'Deployed'} version: {versionData.version}
				{data?.url !== '' && isFetching && LoadingSpinner}
				<Button
					aria-label="Refresh the version"
					variant="secondary"
					style={{ marginLeft: 'auto', padding: '0 1rem' }}
					onClick={refetch}
					disabled={isFetching || !url}
				>
					<FontAwesomeIcon icon={faSync} style={{ paddingRight: '0.25rem' }} />
					Refresh
				</Button>
			</span>
			{(versionData.error || versionData.message) && (
				<span
					className="mb-2"
					style={{ width: '100%', wordBreak: 'break-all' }}
				>
					<Alert variant="danger">
						Failed to refresh:
						<br />
						{beautifyGoErrors(
							(versionData.error || versionData.message) as string,
						)}
					</Alert>
				</span>
			)}
			{dataTargetErrors && (
				<Alert
					variant="danger"
					style={{ paddingLeft: '2rem', marginBottom: 'unset' }}
				>
					{Object.entries(dataTargetErrors).map(([key, error]) => (
						<li key={key}>
							{key}: {error}
						</li>
					))}
				</Alert>
			)}
		</span>
	);
};

export default VersionWithRefresh;
