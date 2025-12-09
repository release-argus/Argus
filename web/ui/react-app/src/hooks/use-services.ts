import { type QueryClient, useQueries } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/query-keys';
import { mapRequest } from '@/utils/api/types/api-request-handler';
import type {
	OrderAPIResponse,
	ServiceSummary,
} from '@/utils/api/types/config/summary';

/**
 * Fetch service summaries for a list of service IDs.
 *
 * @param order - The list of services to fetch.
 * @returns A React Query response object from `useQueries`.
 */
export const useServices = (order?: OrderAPIResponse['order']) =>
	useQueries({
		queries: (order ?? []).map((id) => ({
			placeholderData: { id: id, loading: true },
			queryFn: () => mapRequest('SERVICE_SUMMARY', { serviceID: id }),
			queryKey: QUERY_KEYS.SERVICE.SUMMARY_ITEM(id),
		})),
	});

export const getServiceSummaries = (
	queryClient: QueryClient,
): ServiceSummary[] => {
	return queryClient
		.getQueriesData<ServiceSummary>({
			predicate: (query) =>
				Array.isArray(query.queryKey) &&
				query.queryKey
					.slice(0, QUERY_KEYS.SERVICE.SUMMARY_ITEM_BASE.length)
					.every((v, i) => v === QUERY_KEYS.SERVICE.SUMMARY_ITEM_BASE[i]),
		})
		.map(([_, data]) => data)
		.filter((svc): svc is ServiceSummary => svc !== undefined);
};
