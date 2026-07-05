import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { beautifyGoErrors } from '@/utils';
import { mapRequest } from '@/utils/api/types/api-request-handler';
import { DEPLOYED_VERSION_LOOKUP_TYPE } from '@/utils/api/types/config/service/deployed-version';
import { applyDeployedVersionUpdate } from '@/utils/api/types/requests/service-edit';

export type ApproveManualDeployedVersionVariables = {
	/* The ID of the service to approve the release for. */
	serviceID: string;
	/* The version to approve (the latest known version). */
	targetVersion: string;
};

/**
 * Approves the latest version for a service with a manual `deployed_version` and no
 * WebHooks/Commands, setting the deployed version to match the latest version.
 */
export const useApproveManualDeployedVersion = () => {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (variables: ApproveManualDeployedVersionVariables) =>
			mapRequest('VERSION_REFRESH', {
				data: {
					type: DEPLOYED_VERSION_LOOKUP_TYPE.MANUAL.value,
					version: variables.targetVersion,
				},
				dataSemanticVersioning: null,
				dataTarget: 'deployed_version',
				original: {
					type: DEPLOYED_VERSION_LOOKUP_TYPE.MANUAL.value,
				},
				originalSemanticVersioning: null,
				serviceID: variables.serviceID,
			}),
		onError: (error) => {
			toast.error('Failed to approve release', {
				description: beautifyGoErrors(error.message),
			});
		},
		onSuccess: (data, variables) => {
			applyDeployedVersionUpdate(
				queryClient,
				variables.serviceID,
				data.version || '',
			);
		},
	});
};
