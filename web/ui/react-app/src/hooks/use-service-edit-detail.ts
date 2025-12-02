import { useQuery } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/query-keys.ts';
import { mapRequest } from '@/utils/api/types/api-request-handler.ts';

/**
 * Fetch service detail for a given service.
 *
 * @param serviceID - The service to fetch.
 */
export const useServiceEditDetail = (serviceID?: string | null) =>
	useQuery({
		enabled: !!serviceID,
		queryFn: () =>
			mapRequest('SERVICE_EDIT_DETAIL', { serviceID: serviceID ?? '' }),
		queryKey: QUERY_KEYS.SERVICE.EDIT_ITEM(serviceID ?? ''),
	});
