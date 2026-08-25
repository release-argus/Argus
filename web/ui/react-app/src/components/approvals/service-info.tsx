import type { FC } from 'react';
import { ServiceActionRelease } from '@/components/approvals';
import ServiceInfoDeployedVersion from '@/components/approvals/service-info--deployed-version';
import ServiceInfoLatestVersion from '@/components/approvals/service-info--latest-version';
import { useToolbar } from '@/components/approvals/toolbar/toolbar-context';
import { CardTimestamp } from '@/constants/toolbar';
import { relativeDate } from '@/utils';
import type { ServiceSummary } from '@/utils/api/types/config/summary';

type ServiceInfoProps = {
	service?: ServiceSummary;
	updateAvailable: boolean;
	updateSkipped: boolean;
};

/**
 * The service's information, including the 'latest' version, the 'deployed' version,
 * and the time the 'latest' version was last queried.
 *
 * @param service - The service.
 * @param updateAvailable - Update available for this service?
 * @param updateSkipped - Skipped the latest release for this service?
 */
const ServiceInfo: FC<ServiceInfoProps> = ({
	service,
	updateAvailable,
	updateSkipped,
}) => {
	const { cardTimestamps } = useToolbar();
	const status = service?.status;

	return (
		<div className="flex size-full min-h-22 flex-col gap-y-2">
			<ul className="wrap-anywhere mb-auto flex w-full flex-col gap-1">
				<ServiceInfoDeployedVersion
					hasDeployedVersion={!!service?.deployed_version_type}
					status={service?.status}
					updateAvailable={updateAvailable}
				/>
				<ServiceInfoLatestVersion
					hasDeployedVersion={!!service?.deployed_version_type}
					status={service?.status}
					updateAvailable={updateAvailable}
				/>
			</ul>

			{service && (
				<ServiceActionRelease
					service={service}
					updateAvailable={updateAvailable}
					updateSkipped={updateSkipped}
				/>
			)}
			<small className="flex w-full flex-col gap-1 font-medium text-muted-foreground text-xs leading-none">
				{cardTimestamps.includes(CardTimestamp.Deployed) &&
					status?.deployed_version_timestamp && (
						<span>
							deployed{' '}
							{relativeDate(new Date(status.deployed_version_timestamp))}
						</span>
					)}
				{cardTimestamps.includes(CardTimestamp.Released) &&
					status?.latest_version_timestamp && (
						<span>
							released {relativeDate(new Date(status.latest_version_timestamp))}
						</span>
					)}
				{cardTimestamps.includes(CardTimestamp.Queried) && (
					<span>
						{status?.last_queried ? (
							<>queried {relativeDate(new Date(status.last_queried))}</>
						) : service?.loading ? (
							'loading'
						) : (
							'no successful queries'
						)}
					</span>
				)}
			</small>
		</div>
	);
};

export default ServiceInfo;
