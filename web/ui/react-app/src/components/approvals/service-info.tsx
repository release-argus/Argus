import type { FC } from 'react';
import { ServiceActionRelease } from '@/components/approvals';
import ServiceInfoDeployedVersion from '@/components/approvals/service-info--deployed-version';
import ServiceInfoLatestVersion from '@/components/approvals/service-info--latest-version';
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
	return (
		<div className="flex size-full min-h-22 flex-col gap-y-2">
			<ul className="wrap-anywhere mb-auto flex w-full flex-col gap-1">
				<ServiceInfoDeployedVersion
					hasDeployedVersion={!!service?.has_deployed_version}
					status={service?.status}
					updateAvailable={updateAvailable}
				/>
				<ServiceInfoLatestVersion
					hasDeployedVersion={!!service?.has_deployed_version}
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
			<small className="w-full items-center font-medium text-muted-foreground text-xs leading-none">
				{service?.status?.last_queried ? (
					<>queried {relativeDate(new Date(service?.status.last_queried))}</>
				) : service?.loading ? (
					'loading'
				) : (
					'no successful queries'
				)}
			</small>
		</div>
	);
};

export default ServiceInfo;
