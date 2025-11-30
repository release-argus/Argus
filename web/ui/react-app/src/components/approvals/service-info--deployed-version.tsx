import { AtSign } from 'lucide-react';
import type { FC } from 'react';
import ServiceInfoVersionItem, {
	type ServiceInfoVersionProps,
} from '@/components/approvals/service-info--version-item';
import { cn } from '@/lib/utils';

/**
 * The service's 'deployed' version information.
 *
 * @param service - The service.
 * @param updateAvailable - Update available for this service?
 * @param hasDeployedVersion - The service has a version deployed?
 */
const ServiceInfoDeployedVersion: FC<ServiceInfoVersionProps> = ({
	status,
	updateAvailable,
	hasDeployedVersion,
}) => {
	const {
		latest_version: latestVersion,
		deployed_version: deployedVersion,
		deployed_version_timestamp: deployedVersionTimestamp,
	} = status ?? {};

	// Omit if no 'deployed' version tracker for this service and 'latest' version not deployed.
	if (!hasDeployedVersion && latestVersion === deployedVersion) return null;

	return (
		<ServiceInfoVersionItem
			Icon={hasDeployedVersion ? AtSign : undefined}
			liKey="deployed_v"
			timestamp={deployedVersionTimestamp ?? undefined}
			tipClassName={cn(
				updateAvailable ? 'text-muted-foreground' : 'text-foreground',
			)}
			tooltipLabel="Deployed"
			value={deployedVersion ?? undefined}
		/>
	);
};

export default ServiceInfoDeployedVersion;
