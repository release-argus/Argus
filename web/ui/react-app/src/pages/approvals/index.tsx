import { closestCenter, DndContext } from '@dnd-kit/core';
import { SortableContext } from '@dnd-kit/sortable';
import { useQuery } from '@tanstack/react-query';
import { type ReactElement, useEffect, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import { ApprovalsToolbar, Service } from '@/components/approvals';
import {
	type ApprovalsToolbarOptions,
	DEFAULT_HIDE_VALUE,
	HideValue,
	type HideValueType,
	URL_PARAMS,
} from '@/constants/toolbar';
import { useWebSocket } from '@/contexts/websocket';
import { useSortableServices } from '@/hooks/use-sortable-services';
import { QUERY_KEYS } from '@/lib/query-keys';
import type { TagsTriType } from '@/types/util';
import { mapRequest } from '@/utils/api/types/api-request-handler';

const toolbarDefaults: ApprovalsToolbarOptions = {
	editMode: false,
	hide: DEFAULT_HIDE_VALUE,
	search: '',
	tags: { exclude: [], include: [] },
};

/**
 * @returns The 'approvals' page, including a toolbar, and a list of services.
 */
export const Approvals = (): ReactElement => {
	const { monitorData, setMonitorData } = useWebSocket();
	const [searchParams] = useSearchParams();

	const toolbarOptions: ApprovalsToolbarOptions = useMemo(() => {
		const search =
			searchParams.get(URL_PARAMS.SEARCH) ?? toolbarDefaults.search;

		// Extract tags from URL.
		const tagsIncludeQueryParam = searchParams.get(URL_PARAMS.TAGS_INCLUDE);
		const tagsExcludeQueryParam = searchParams.get(URL_PARAMS.TAGS_EXCLUDE);
		let tags: TagsTriType;
		try {
			tags =
				tagsIncludeQueryParam || tagsExcludeQueryParam
					? {
							exclude: JSON.parse(tagsExcludeQueryParam ?? '[]') as string[],
							include: JSON.parse(tagsIncludeQueryParam ?? '[]') as string[],
						}
					: toolbarDefaults.tags;
		} catch (e) {
			toast.error('Failed to parse tags from URL', {
				description: `Error: ${e instanceof Error ? e.message : String(e)}`,
			});
			tags = toolbarDefaults.tags;
		}

		const editMode = searchParams.has(URL_PARAMS.EDIT_MODE);

		// Extract hide options from URL.
		const hideQueryParam = searchParams.get(URL_PARAMS.HIDE);
		let hide: HideValueType[] = [];
		if (hideQueryParam === null) {
			hide = [...toolbarDefaults.hide];
		} else if (hideQueryParam) {
			try {
				const parsedHide: unknown = JSON.parse(hideQueryParam);
				if (Array.isArray(parsedHide)) {
					const validValues = Object.values(HideValue) as HideValueType[];
					hide = parsedHide
						.map(Number)
						.filter(
							(num: unknown) =>
								Number.isFinite(num) &&
								validValues.includes(num as HideValueType),
						) as HideValueType[];
				}
			} catch (e) {
				toast.error('Failed to parse hide options from URL', {
					description: `Error: ${e instanceof Error ? e.message : String(e)}`,
				});
				hide = [];
			}
		}

		return { editMode, hide, search, tags };
	}, [searchParams]);

	const {
		sensors,
		handleDragEnd,
		handleSaveOrder,
		hasOrderChanged,
		resetOrder,
	} = useSortableServices();

	// Fetch the service ordering from the API.
	const { data: orderData } = useQuery({
		gcTime: 1000 * 60 * 30, // 30 minutes.
		queryFn: () => mapRequest('SERVICE_ORDER_GET', null),
		queryKey: QUERY_KEYS.SERVICE.ORDER(),
	});
	// Push the ordering to the 'monitorData' state.
	// biome-ignore lint/correctness/useExhaustiveDependencies: setMonitorData stable.
	useEffect(() => {
		if (orderData && orderData.order.length > 0)
			setMonitorData({
				page: 'APPROVALS',
				sub_type: 'ORDER',
				type: 'SERVICE',
				...orderData,
			});
	}, [orderData]);

	// Filter the services based on the toolbar options.
	const filteredServices = useMemo(() => {
		const {
			search = '',
			tags = toolbarDefaults.tags,
			hide,
		} = {
			...toolbarOptions,
			search: toolbarOptions.search.toLowerCase(),
		};
		const filterOnTags = tags.include.length > 0 || tags.exclude.length > 0;
		const excludeOnly = filterOnTags && tags.include.length === 0;

		return Object.values(monitorData.order).filter((serviceID) => {
			const service = monitorData.service[serviceID];
			if (!service) return false;

			const hideInactiveServices = hide.includes(HideValue.Inactive);
			if (
				!monitorData.tagsLoaded &&
				(!hideInactiveServices || service.active !== false)
			)
				return true;

			// Filter on 'tags'.
			//     Have no tags to filter on,
			//   OR
			//     The service doesn't have any EXCLUDE tags
			//       AND
			//     We are only excluding tags, OR the service has all INCLUDE tags.
			const hasTags =
				!filterOnTags ||
				(!tags.exclude.some((tag) => service.tags?.includes(tag)) &&
					(excludeOnly ||
						tags.include.some((tag) => service.tags?.includes(tag))));
			if (!hasTags) return false;

			// Filter on 'name'.
			const name = (service.name ?? serviceID).toLowerCase();
			if (!name.includes(search)) return false;

			// Filter on 'hide' options.
			const skipped =
				service.status?.latest_version &&
				service.status.approved_version ===
					`SKIP_${service.status.latest_version}`;
			const upToDate =
				service.status?.deployed_version === service.status?.latest_version;
			return (
				// hideUpToDate: deployed_version NOT latest_version.
				(!hide.includes(HideValue.UpToDate) || !upToDate) &&
				// hideUpdatable: deployed_version IS latest_version AND approved_version NOT "SKIP_"+latest_version.
				(!hide.includes(HideValue.Updatable) || upToDate || skipped) &&
				// hideSkipped: approved_version NOT "SKIP_"+latest_version OR NO approved_version.
				(!hide.includes(HideValue.Skipped) || !skipped) &&
				// hideInactive: active NOT false.
				(!hideInactiveServices || service.active !== false)
			);
		});
	}, [
		toolbarOptions,
		monitorData.service,
		monitorData.order,
		monitorData.tagsLoaded,
	]);

	return (
		<>
			<ApprovalsToolbar
				hasOrderChanged={hasOrderChanged}
				onEditModeToggle={(value: boolean) => {
					if (!value) resetOrder();
				}}
				onSaveOrder={() => handleSaveOrder()}
				values={toolbarOptions}
			/>
			<div className="grid gap-4 [grid-template-columns:repeat(auto-fill,minmax(17.5rem,1fr))]">
				<DndContext
					autoScroll={{
						acceleration: 100,
						enabled: true,
						interval: 5,
						threshold: {
							x: 0.2, // Start scrolling when within 20% of the edge.
							y: 0.2,
						},
					}}
					collisionDetection={closestCenter}
					onDragEnd={handleDragEnd}
					sensors={sensors}
				>
					<SortableContext items={monitorData.order}>
						{monitorData.order.length ===
							Object.keys(monitorData.service).length &&
							filteredServices.map((id) => (
								<Service
									editable={toolbarOptions.editMode}
									id={id}
									key={id}
									service={monitorData.service[id]}
								/>
							))}
					</SortableContext>
				</DndContext>
			</div>
		</>
	);
};
