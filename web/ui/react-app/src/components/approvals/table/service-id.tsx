import type { Row } from '@tanstack/react-table';
import type { FC } from 'react';
import { Link } from 'react-router';
import { Button } from '@/components/ui/button';
import type { ServiceSummary } from '@/utils/api/types/config/summary';

type ServiceIDProps = {
	/* The row in the table */
	row: Row<ServiceSummary>;
};

/**
 * A functional component displaying the ID of a service, linking to the service's
 * web URL when it has one (as the title does in the grid layout).
 *
 * @param row - The ServiceSummary data of the service.
 *
 * @returns The service ID, as a link when the service has a web URL.
 */
export const ServiceID: FC<ServiceIDProps> = ({ row }) => {
	const { id, url } = row.original;

	if (!url) return id;

	return (
		<Button asChild className="p-0 text-foreground" size="fit" variant="link">
			<Link rel="noreferrer noopener" target="_blank" to={url}>
				{id}
			</Link>
		</Button>
	);
};
