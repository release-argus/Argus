import { useQuery } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/query-keys.ts';
import { mapRequest } from '@/utils/api/types/api-request-handler.ts';

/**
 * Fetch service summary for a given service.
 *
 * @param serviceID - The service to fetch.
 */
export const useServiceSummary = (serviceID?: string | null) =>
	useQuery({
		enabled: !!serviceID,
		queryFn: () =>
			mapRequest('SERVICE_SUMMARY', { serviceID: serviceID ?? '' }),
		queryKey: QUERY_KEYS.SERVICE.SUMMARY_ITEM(serviceID ?? ''),
	});
