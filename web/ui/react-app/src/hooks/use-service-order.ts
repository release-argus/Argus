import { useQuery } from '@tanstack/react-query';
import { QUERY_KEYS } from '@/lib/query-keys.ts';
import { mapRequest } from '@/utils/api/types/api-request-handler.ts';
import type { OrderAPIResponse } from '@/utils/api/types/config/summary.ts';

/*
 * Fetch the order of services.
 *
 * @returns A React Query response object from `useQuery`.
 * */
export const useServiceOrder = () =>
	useQuery<OrderAPIResponse>({
		gcTime: 1000 * 60 * 30, // 30 minutes.
		queryFn: () => mapRequest('SERVICE_ORDER_GET', null),
		queryKey: QUERY_KEYS.SERVICE.ORDER(),
	});
