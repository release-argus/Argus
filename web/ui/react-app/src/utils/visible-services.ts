import {
	type ApprovalsToolbarOptions,
	HideValue,
	type HideValueType,
} from '@/constants/toolbar';
import type { TagsTriType } from '@/types/util';
import type {
	ServiceSummary,
	ServiceUpdateState,
} from '@/utils/api/types/config/summary';

/* The 'hide' toggle that removes each update state from view. */
const HIDE_VALUE_FOR_STATE: Record<
	NonNullable<ServiceUpdateState>,
	HideValueType
> = {
	AVAILABLE: HideValue.Updatable,
	SKIPPED: HideValue.Skipped,
	UP_TO_DATE: HideValue.UpToDate,
};

/**
 * Whether a service's tags match the toolbar filters.
 *
 * @param svc - The service summary to check.
 * @param tags - The toolbar tag filters.
 *
 * @returns True if the service matches the tag filters, false otherwise.
 */
const matchesTags = (svc: ServiceSummary, tags: TagsTriType): boolean =>
	!tags.exclude.some((tag) => svc.tags?.includes(tag)) &&
	(tags.include.length === 0 ||
		tags.include.some((tag) => svc.tags?.includes(tag)));

/**
 * Whether a service's name or ID matches the toolbar search.
 *
 * @param svc - The service summary to check.
 * @param search - The lower-cased search string.
 *
 * @returns True if the service matches the search, false otherwise.
 */
const matchesSearch = (svc: ServiceSummary, search: string): boolean =>
	(svc.name ?? svc.id).toLowerCase().includes(search);

/**
 * Whether a service's status matches the toolbar hide filters.
 *
 * @param svc  - The service summary to check.
 * @param hide - The toolbar hide filters.
 *
 * @returns True if the service matches the hide filters, false otherwise.
 */
const matchesHide = (svc: ServiceSummary, hide: HideValueType[]): boolean => {
	if (svc.active === false && hide.includes(HideValue.Inactive)) return false;

	const state = svc.status?.state;
	return !state || !hide.includes(HIDE_VALUE_FOR_STATE[state]);
};

export type ToolbarFilters = Pick<
	ApprovalsToolbarOptions,
	'hide' | 'search' | 'tags'
>;

/**
 * The services the dashboard renders under the given toolbar filters, in the
 * order given. Still-loading services always render - their filterable fields
 * haven't arrived yet.
 *
 * @param services - Service summaries, in dashboard order.
 * @param filters - The current toolbar filter values.
 *
 * @returns The services to render.
 */
export const visibleServices = (
	services: ServiceSummary[],
	filters: ToolbarFilters,
): ServiceSummary[] => {
	const search = filters.search.toLowerCase();

	return services.filter(
		(svc) =>
			svc.loading ||
			(matchesTags(svc, filters.tags) &&
				matchesSearch(svc, search) &&
				matchesHide(svc, filters.hide)),
	);
};
