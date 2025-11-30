import type { LucideIcon } from 'lucide-react';
import type { FC } from 'react';
import Tip from '@/components/ui/tip';
import { cn } from '@/lib/utils';
import { relativeDate } from '@/utils';
import type { StatusSummaryType } from '@/utils/api/types/config/summary';

export type ServiceInfoVersionProps = {
	/* The service's status. */
	status?: StatusSummaryType;
	/* Whether an update is available for this service. */
	updateAvailable: boolean;
	/* Whether the service has a deployed version. */
	hasDeployedVersion: boolean;
};

export type ServiceInfoVersionItemProps = {
	/* Unique key for the list item. */
	liKey?: string;

	/* Icon to display before the version value. */
	Icon?: LucideIcon;
	/* Version value to display. */
	value?: string | null;
	/* Timestamp of the version. */
	timestamp?: string | null;

	/* Label for the tooltip. */
	tooltipLabel: string;
	/* Class name for the tooltip. */
	tipClassName?: string;
	/* Props for the tooltip content. */
	contentProps?: Parameters<typeof Tip>[0]['contentProps'];
};

/**
 * Service version item with an associated icon, value, timestamp and tooltip label.
 *
 * @property liKey - Unique key required for the `<li>` element in the rendered list.
 * @property Icon - An optional icon to display alongside the item details.
 * @property value - The version value to be displayed.
 * @property timestamp - Timestamp information for the version.
 * @property tooltipLabel - Descriptive label text for the tooltip.
 * @property tipClassName - Optional class name for styling the tooltip.
 * @property contentProps - Additional properties to pass to the tooltip content configuration.
 */
const ServiceInfoVersionItem: FC<ServiceInfoVersionItemProps> = ({
	liKey,
	Icon,
	value,
	timestamp,
	tooltipLabel,
	tipClassName,
	contentProps,
}) => {
	const tooltipContent = `${tooltipLabel} ${timestamp ? relativeDate(new Date(timestamp)) : 'Unknown'}`;

	return (
		<li key={liKey}>
			<Tip
				className={cn('flex flex-wrap items-center gap-x-1 p-0', tipClassName)}
				content={tooltipContent}
				contentProps={contentProps}
				delayDuration={500}
			>
				{Icon && <Icon />}
				<p className="flex h-4 items-center justify-end text-balance font-mono font-semibold">
					{value ?? 'Unknown'}
				</p>
			</Tip>
		</li>
	);
};

export default ServiceInfoVersionItem;
