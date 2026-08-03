import { useMutation } from '@tanstack/react-query';
import { LoaderCircle } from 'lucide-react';
import type { FC } from 'react';
import { toast } from 'sonner';
import { DropdownMenuItem } from '@/components/ui/dropdown-menu';
import { beautifyGoErrors } from '@/utils';
import { mapRequest } from '@/utils/api/types/api-request-handler';
import { DEPLOYED_VERSION_LOOKUP_TYPE } from '@/utils/api/types/config/service/deployed-version';
import type { ServiceSummary } from '@/utils/api/types/config/summary';

/* Number of `deployed_version` refreshes to run concurrently. */
const REFRESH_CONCURRENCY = 5;

/*
 * Shared id so the loading/success/error toasts update in place, pinned to
 * its own position so a flood of per-service update notifications (default
 * bottom-right) can't push the progress toast out of the visible stack.
 */
const REFRESH_TOAST_OPTS = {
	id: 'refresh-deployed-versions',
	position: 'top-center',
} as const;

/**
 * Whether a service's deployed_version has an external source to re-query.
 *
 * @param svc - The service summary to check.
 * @returns True if the service's deployed_version can be re-queried.
 */
export const canRefreshDeployedVersion = (svc: ServiceSummary): boolean =>
	Boolean(svc.deployed_version_type) &&
	svc.deployed_version_type !== DEPLOYED_VERSION_LOOKUP_TYPE.MANUAL.value;

type RefreshDeployedVersionsMenuItemProps = {
	/* IDs of the services to re-query. */
	serviceIDs: string[];
};

/**
 * Re-queries the `deployed_version` of each of `serviceIDs`.
 *
 * @param serviceIDs - IDs of the services to re-query.
 */
const RefreshDeployedVersionsMenuItem: FC<
	RefreshDeployedVersionsMenuItemProps
> = ({ serviceIDs }) => {
	const total = serviceIDs.length;

	const { mutate: refreshVisible, isPending } = useMutation({
		mutationFn: async () => {
			let completed = 0;
			let failed = 0;
			toast.loading(
				`Refreshing deployed versions: ${completed}/${total}`,
				REFRESH_TOAST_OPTS,
			);

			const reportProgress = () => {
				completed += 1;
				toast.loading(
					`Refreshing deployed versions: ${completed}/${total}`,
					REFRESH_TOAST_OPTS,
				);
			};

			let next = 0;
			const worker = async () => {
				while (next < serviceIDs.length) {
					const serviceID = serviceIDs[next++];
					try {
						await mapRequest('VERSION_REFRESH', {
							dataSemanticVersioning: null,
							dataTarget: 'deployed_version',
							original: null,
							originalSemanticVersioning: null,
							serviceID: serviceID,
						});
					} catch {
						failed += 1;
					} finally {
						reportProgress();
					}
				}
			};
			await Promise.all(
				Array.from(
					{ length: Math.min(REFRESH_CONCURRENCY, serviceIDs.length) },
					worker,
				),
			);

			return { failed, total };
		},
		onError: (error) => {
			toast.error('Failed to refresh deployed versions', {
				...REFRESH_TOAST_OPTS,
				description: beautifyGoErrors(error.message),
			});
		},
		onSuccess: ({ failed, total }) => {
			if (failed > 0) {
				toast.error(`Refreshed deployed versions`, {
					description: (
						<div className="flex flex-col gap-1">
							<span>Succeeded: {total - failed}</span>
							<span>Failed: {failed}</span>
						</div>
					),
					...REFRESH_TOAST_OPTS,
				});
			} else {
				toast.success(`Refreshed deployed versions`, {
					description: `Succeeded: ${total}`,
					...REFRESH_TOAST_OPTS,
				});
			}
		},
	});

	return (
		<DropdownMenuItem
			className="cursor-pointer"
			disabled={isPending}
			onClick={() => refreshVisible()}
		>
			Refresh visible deployed versions
			{isPending && <LoaderCircle className="ml-2 animate-spin" />}
		</DropdownMenuItem>
	);
};

export default RefreshDeployedVersionsMenuItem;
