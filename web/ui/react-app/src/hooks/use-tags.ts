import { useMemo } from 'react';
import { useServiceOrder } from '@/hooks/use-service-order';
import { useServices } from '@/hooks/use-services';

export type UseTagsResult = {
	/* Sorted list of all tags across services. */
	tags: string[];
	/* Usage count map of all tags across services. */
	counts: Record<string, number>;
	/* Whether any service is currently loading. */
	isLoading: boolean;
};

/**
 * Build a sorted list of all tags across services, and a usage count map.
 *
 * @param excludeServiceID - Optional service ID to exclude from the counts (but still include its tags in the list).
 *
 * @returns
 * - `tags`: Sorted list of all tags across services.
 * - `counts`: Usage count map of all tags across services.
 * - `isLoading`: Whether any data is currently loading.
 */
export const useTags = (excludeServiceID?: string | null): UseTagsResult => {
	const { data: orderData, isFetching: isFetchingOrderData } =
		useServiceOrder();
	const services = useServices(orderData?.order);

	return useMemo(() => {
		const tagCounts: Record<string, number> = {};
		const allTags = new Set<string>();

		// Retrieve all tags, and a count of usage for each tag.
		for (const svc of services) {
			const svcID = svc.data?.id;
			const svcTags = svc.data?.tags ?? [];

			for (const t of svcTags) allTags.add(t);

			// Count usage, optionally excluding a specific service.
			if (svcID && excludeServiceID && svcID === excludeServiceID) continue;
			for (const t of svcTags) tagCounts[t] = (tagCounts[t] ?? 0) + 1;
		}

		// Alphabetise tags
		const sortedTags = Array.from(allTags).toSorted((a, b) =>
			a.localeCompare(b, undefined, { sensitivity: 'base' }),
		);

		// Loading state if fetching order, or any service still loading.
		const anyLoading =
			isFetchingOrderData || services.some((s) => s.isFetching);

		return { counts: tagCounts, isLoading: anyLoading, tags: sortedTags };
	}, [excludeServiceID, isFetchingOrderData, services]);
};
