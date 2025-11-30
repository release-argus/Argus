import { CornerDownRight } from 'lucide-react';
import type { FC } from 'react';
import ServiceInfoVersionItem, {
	type ServiceInfoVersionProps,
} from '@/components/approvals/service-info--version-item';

/**
 * The service's 'latest' version information.
 *
 * @param service - The service.
 * @param updateAvailable - Update available for this service?
 * @param hasDeployedVersion - The service has a version deployed?
 */
const ServiceInfoLatestVersion: FC<ServiceInfoVersionProps> = ({
	status,
	updateAvailable,
	hasDeployedVersion,
}) => {
	const {
		latest_version: latestVersion,
		latest_version_timestamp: latestVersionTimestamp,
		deployed_version: deployedVersion,
	} = status ?? {};

	// Omit if 'latest' version is deployed.
	if (latestVersion === deployedVersion && hasDeployedVersion) return null;

	return (
		<ServiceInfoVersionItem
			contentProps={{ side: 'bottom' }}
			Icon={updateAvailable ? CornerDownRight : undefined}
			liKey="latest_v"
			timestamp={latestVersionTimestamp ?? undefined}
			tipClassName="text-foreground"
			tooltipLabel="Latest version found"
			value={latestVersion ?? undefined}
		/>
	);
};

export default ServiceInfoLatestVersion;
