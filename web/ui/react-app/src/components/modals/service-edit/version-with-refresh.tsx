import { Alert, Button } from 'react-bootstrap';
import {
	DeployedVersionLookupEditType,
	LatestVersionLookupEditType,
} from 'types/service-edit';
import { FC, useMemo, useState } from 'react';
import { beautifyGoErrors, fetchVersionJSON, } from 'utils';
import {
	convertUIDeployedVersionDataEditToAPI,
	convertUILatestVersionDataEditToAPI,
} from 'components/modals/service-edit/util';
import { faSpinner, faSync } from '@fortawesome/free-solid-svg-icons';
import { useFormContext, useWatch } from 'react-hook-form';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ServiceOptionsType } from 'types/config';
import cx from 'classnames';
import { useErrors } from 'hooks/errors';
import { useMutation } from '@tanstack/react-query';
import useValuesRefetch from 'hooks/values-refetch';
import { useWebSocket } from 'contexts/websocket';

interface VersionWithRefreshProps {
	/** The type of version to display/fetch (0 for "latest_version", otherwise "deployed_version"). */
	vType: 0 | 1;
	/** Optional additional CSS class names for styling the component. */
	className?: string;
	/** The unique identifier for the service. */
	serviceID: string;
	/** The original version data for the service. */
	original?: LatestVersionLookupEditType | DeployedVersionLookupEditType;
	/** The .options of the original version data. */
	original_options?: ServiceOptionsType;
}

/**
 * The version with a button to refetch.
 *
 * @param vType - 0: latest_version, 1: deployed_version.
 * @param className - Extra class name(s) for the component.
 * @param serviceID - The ID of the service.
 * @param original - The original values in the form.
 * @param original_options - The original service.options of the form.
 * @returns The current version with a button to trigger a refetch.
 */
const VersionWithRefresh: FC<VersionWithRefreshProps> = ({
	vType,
	className,
	serviceID,
	original,
	original_options,
}) => {
	const [lastFetched, setLastFetched] = useState(0);
	const { monitorData } = useWebSocket();
	const { clearErrors, setError, setValue, trigger } = useFormContext();
	const dataTarget = vType === 0 ? 'latest_version' : 'deployed_version';
	const convertedOriginal = useMemo(() => {
		if (original === null) return {};
		const preparedOriginal = serviceID
			? original
			: // Remove type from original if it's a new service.
			{ ...original, type: '' };
		// Latest version
		if (dataTarget === 'latest_version')
			return convertUILatestVersionDataEditToAPI(
				preparedOriginal as LatestVersionLookupEditType,
			);
		// Deployed version
		return convertUIDeployedVersionDataEditToAPI(
			preparedOriginal as DeployedVersionLookupEditType,
		);
	}, [original, serviceID, dataTarget]);
	const url: string | undefined = useWatch({ name: `${dataTarget}.url` });
	const dataTargetErrors = useErrors(dataTarget, true);
	const { data, refetchData } = useValuesRefetch<{ [x: string]: any }>(
		dataTarget,
	);
	const { data: semanticVersioning, refetchData: refetchSemanticVersioning } =
		useValuesRefetch<boolean>('options.semantic_versioning');

	const {
		mutate: fetchVersion,
		data: versionData,
		isPending: isFetching,
	} = useMutation({
		mutationFn: () =>
			fetchVersionJSON({
				serviceID,
				dataTarget,
				semanticVersioning,
				options: original_options,
				data,
				original: convertedOriginal,
			}),
		onSuccess: (data) => {
			if (data.version) {
				setValue(`${dataTarget}.version`, data.version);
				clearErrors(`${dataTarget}.version`);
			}
		},
		onError: (error) => {
			setValue(`${dataTarget}.version`, '');
			setError(`${dataTarget}.version`, {
				type: 'manual',
				message: beautifyGoErrors(error instanceof Error ? error.message : String(error)),
			});
		},
	});
	const version = useWatch({ name: `${dataTarget}.version` }) ?? monitorData.service[serviceID]?.status?.[dataTarget];

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
				fetchVersion();
				setLastFetched(currentTime);
			}
		});
		return () => clearTimeout(timeout);
	};
	const LoadingSpinner = (
		<FontAwesomeIcon icon={faSpinner} spin style={{ marginLeft: '0.5rem' }} />
	);

	return (
		<span className={cx('w-100', className)}>
			<span className="pt-1 pb-2" style={{ display: 'flex' }}>
				{vType === 0 ? 'Latest' : 'Deployed'} version: {version}
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
			{(versionData?.error || versionData?.message) && (
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
