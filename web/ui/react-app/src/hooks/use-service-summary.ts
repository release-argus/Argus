import { useQuery } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/query-keys';
import { mapRequest } from '@/utils/api/types/api-request-handler';

/**
 * Fetch service summary for a given service.
 *
 * @param serviceID - The service to fetch.
 * @returns A React Query response object from `useQuery`.
 */
export const useServiceSummary = (serviceID?: string | null) =>
	useQuery({
		enabled: !!serviceID,
		queryFn: () =>
			mapRequest('SERVICE_SUMMARY', { serviceID: serviceID ?? '' }),
		queryKey: QUERY_KEYS.SERVICE.SUMMARY_ITEM(serviceID ?? ''),
	});
