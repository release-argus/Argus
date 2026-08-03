import { useMutation, useQueryClient } from '@tanstack/react-query';
import { LoaderCircle } from 'lucide-react';
import type { FC } from 'react';
import { toast } from 'sonner';
import { DropdownMenuItem } from '@/components/ui/dropdown-menu';
import { getServiceSummaries } from '@/hooks/use-services';
import { beautifyGoErrors } from '@/utils';
import { mapRequest } from '@/utils/api/types/api-request-handler';

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
 * Re-queries `deployed_version` for every configured service.
 */
const RefreshDeployedVersionsMenuItem: FC = () => {
	const queryClient = useQueryClient();

	const { mutate: refreshAll, isPending } = useMutation({
		mutationFn: async () => {
			const serviceIDs = getServiceSummaries(queryClient)
				.filter((svc) => !svc.loading && svc.deployed_version_type)
				.map((svc) => svc.id);

			const total = serviceIDs.length;
			if (total === 0) return { failed: 0, total };

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

			for (let i = 0; i < serviceIDs.length; i += REFRESH_CONCURRENCY) {
				const batch = serviceIDs.slice(i, i + REFRESH_CONCURRENCY);
				const results = await Promise.allSettled(
					batch.map((serviceID) =>
						mapRequest('VERSION_REFRESH', {
							dataSemanticVersioning: null,
							dataTarget: 'deployed_version',
							original: null,
							originalSemanticVersioning: null,
							serviceID,
						}).finally(reportProgress),
					),
				);
				failed += results.filter((r) => r.status === 'rejected').length;
			}

			return { failed, total };
		},
		onError: (error) => {
			toast.error('Failed to refresh deployed versions', {
				...REFRESH_TOAST_OPTS,
				description: beautifyGoErrors(error.message),
			});
		},
		onSuccess: ({ failed, total }) => {
			if (total === 0) {
				toast.info(
					'No services with a deployed version to refresh',
					REFRESH_TOAST_OPTS,
				);
				return;
			}
			if (failed > 0) {
				toast.error(
					`Refreshed deployed versions: ${failed}/${total} failed`,
					REFRESH_TOAST_OPTS,
				);
			} else {
				toast.success(
					`Refreshed deployed version for ${total} service${total === 1 ? '' : 's'}`,
					REFRESH_TOAST_OPTS,
				);
			}
		},
	});

	return (
		<DropdownMenuItem
			className="cursor-pointer"
			disabled={isPending}
			onClick={() => refreshAll()}
		>
			Refresh all deployed versions
			{isPending && <LoaderCircle className="ml-2 animate-spin" />}
		</DropdownMenuItem>
	);
};

export default RefreshDeployedVersionsMenuItem;
