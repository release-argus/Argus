import type { Row } from '@tanstack/react-table';
import { CheckCircle, HelpCircle, MinusCircle, XCircle } from 'lucide-react';
import type { ComponentType, FC } from 'react';
import { cn } from '@/lib/utils';
import type {
	ServiceSummary,
	ServiceUpdateState,
} from '@/utils/api/types/config/summary';

type ServiceStatusProps = {
	/* The row in the table */
	row: Row<ServiceSummary>;
};

type StatusMapTypes = { className: string; icon: ComponentType; label: string };
type StatusMapType = Record<NonNullable<ServiceUpdateState>, StatusMapTypes>;

const STATUS_MAP: StatusMapType = {
	AVAILABLE: {
		className: 'bg-destructive dark:bg-destructive/60 text-white',
		icon: XCircle,
		label: 'Update Available',
	},
	SKIPPED: {
		className: 'bg-primary text-foreground',
		icon: MinusCircle,
		label: 'Update Skipped',
	},
	UP_TO_DATE: {
		className: 'bg-primary text-primary-foreground dark:text-white',
		icon: CheckCircle,
		label: 'Up To Date',
	},
} as const;

/**
 * A functional component displaying the current status of a service using a status message, an associated icon,
 * and a background colour, based on the service's state.
 *
 * The possible states include:
 * - `AVAILABLE`: Indicates that an update is available.
 * - `SKIPPED`: Indicates that an update has been skipped.
 * - `UP_TO_DATE`: Indicates that the service is up to date.
 *
 * @param row - The ServiceSummary data of the service.
 *
 * @returns A visual representation of the service's status.
 */
export const ServiceStatus: FC<ServiceStatusProps> = ({ row }) => {
	const state = row.original?.status?.state as
		| keyof typeof STATUS_MAP
		| undefined;
	const cfg = state ? STATUS_MAP[state] : undefined;
	const Icon = cfg?.icon ?? HelpCircle;

	return (
		<div
			className={cn(
				'inline-flex shrink-0 items-center justify-center gap-1 rounded-full px-2 [&>svg]:size-4',
				cfg?.className ?? 'bg-muted text-muted-foreground',
			)}
		>
			<Icon />
			{cfg?.label ?? 'Unknown'}
		</div>
	);
};
