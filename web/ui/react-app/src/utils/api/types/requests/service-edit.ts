import type { QueryClient } from '@tanstack/react-query';
import { formatRFC3339 } from 'date-fns';
import { QUERY_KEYS } from '@/lib/query-keys';
import type {
	ServiceSummary,
	ServiceUpdateState,
	StatusSummaryType,
} from '@/utils/api/types/config/summary';

export type ServiceEditRequestBuilder = {
	/* The service ID. */
	serviceID: string | null;
	/* Service JSON */
	body: unknown;
};

export type ServiceEditResponse = {
	/* The result of the edit. */
	message: string;
};

export const serviceSummaryReducer = (
	service?: ServiceSummary,
	oldData?: ServiceSummary,
): ServiceSummary => {
	const status = {
		...oldData?.status,
		...service?.status,
	};
	status.deployed_version ??= status?.latest_version;
	status.deployed_version_timestamp ??= status?.latest_version_timestamp;
	// Treat all version updates as both latest/deployed if no deployed version for this service.
	if (oldData?.deployed_version_type === '' && service?.status) {
		service.status.deployed_version ??= service.status.latest_version;
		service.status.deployed_version_timestamp ??=
			service.status.latest_version_timestamp;
		service.status.latest_version ??= service.status.deployed_version;
		service.status.latest_version_timestamp ??=
			service.status.deployed_version_timestamp;
	}

	return {
		...oldData,
		...service,
		loading: false,
		status: {
			...status,
			state: getServiceUpdateState(status),
		},
	} as ServiceSummary;
};

/**
 * Applies a new `deployed_version` to a service's cached summary, re-deriving
 * `status.state` (and stamping `deployed_version_timestamp`) via `serviceSummaryReducer`
 * so 'update available'/'approve'/'skip' UI stays in sync with the new version.
 *
 * @param queryClient - The React Query client.
 * @param serviceID - The ID of the service to update.
 * @param version - The new `deployed_version`.
 */
export const applyDeployedVersionUpdate = (
	queryClient: QueryClient,
	serviceID: string,
	version: string,
) => {
	queryClient.setQueryData<ServiceSummary>(
		QUERY_KEYS.SERVICE.SUMMARY_ITEM(serviceID),
		(oldData) =>
			serviceSummaryReducer(
				{
					id: serviceID,
					status: {
						// Clear approved_version only when the current latest version is being deployed.
						approved_version:
							version === oldData?.status?.latest_version
								? ''
								: oldData?.status?.approved_version,
						deployed_version: version,
						deployed_version_timestamp: formatRFC3339(new Date()),
					},
				},
				oldData,
			),
	);
};

export const getServiceUpdateState = (
	status?: StatusSummaryType,
): ServiceUpdateState => {
	// Loading 'status' still.
	if (status === undefined) return null;

	// Latest version is deployed.
	if (status.latest_version === status.deployed_version) return 'UP_TO_DATE';

	// Latest version is skipped.
	if (
		status.approved_version &&
		status.approved_version === `SKIP_${status.latest_version}`
	)
		return 'SKIPPED';

	// Latest version must not be deployed/skipped.
	return 'AVAILABLE';
};
