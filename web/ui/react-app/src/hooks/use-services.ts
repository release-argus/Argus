import { useQueries } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/query-keys.ts';
import { mapRequest } from '@/utils/api/types/api-request-handler.ts';
import type { OrderAPIResponse } from '@/utils/api/types/config/summary.ts';

/**
 * Fetch service summaries for a list of service IDs.
 *
 * @param order - The list of services to fetch.
 */
export const useServices = (order?: OrderAPIResponse['order']) =>
	useQueries({
		queries: (order ?? []).map((id) => ({
			placeholderData: { id: id, loading: true },
			queryFn: () => mapRequest('SERVICE_SUMMARY', { serviceID: id }),
			queryKey: QUERY_KEYS.SERVICE.SUMMARY_ITEM(id),
		})),
	});
