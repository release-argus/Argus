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
