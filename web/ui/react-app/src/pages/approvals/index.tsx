import { ApprovalsToolbar, Service } from 'components/approvals';
import { DndContext, closestCenter } from '@dnd-kit/core';
import { ReactElement, useEffect, useMemo } from 'react';

import { ApprovalsToolbarOptions } from 'types/util';
import { Container } from 'react-bootstrap';
import { DEFAULT_HIDE_VALUE } from 'components/approvals/toolbar/filter-dropdown';
import { OrderAPIResponse } from 'types/summary';
import { SortableContext } from '@dnd-kit/sortable';
import { URL_PARAMS } from 'constants/toolbar';
import { fetchJSON } from 'utils';
import { useQuery } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import { useSortableServices } from 'hooks/sortable-services';
import { useWebSocket } from 'contexts/websocket';

/**
 * @returns The approvals page, which includes a toolbar, and a list of services.
 */
export const Approvals = (): ReactElement => {
	const { monitorData, setMonitorData } = useWebSocket();
	const [searchParams] = useSearchParams();

	const toolbarDefaults: ApprovalsToolbarOptions = {
		search: '',
		tags: [],
		editMode: false,
		hide: DEFAULT_HIDE_VALUE,
	};

	const toolbarOptions: ApprovalsToolbarOptions = useMemo(() => {
		const search =
			searchParams.get(URL_PARAMS.SEARCH) ?? toolbarDefaults.search;

		const tagsQueryParam = searchParams.get(URL_PARAMS.TAGS);
		let tags: string[] = [];
		try {
			tags = tagsQueryParam ? JSON.parse(tagsQueryParam) : toolbarDefaults.tags;
		} catch {}

		const editMode = searchParams.has(URL_PARAMS.EDIT_MODE);

		const hideQueryParam = searchParams.get(URL_PARAMS.HIDE);
		let hide: number[] = [];
		if (hideQueryParam === null) {
			hide = toolbarDefaults.hide;
		} else if (hideQueryParam) {
			try {
				hide = JSON.parse(hideQueryParam)
					.map(Number)
					.filter((num: unknown) => Number.isFinite(num));
			} catch {}
		} else {
			hide = [];
		}

		return { search, tags, editMode, hide };
	}, [searchParams]);

	const {
		sensors,
		handleDragEnd,
		handleSaveOrder,
		hasOrderChanged,
		resetOrder,
	} = useSortableServices(monitorData, setMonitorData);

	const { data: orderData } = useQuery({
		queryKey: ['service/order'],
		queryFn: () => fetchJSON<OrderAPIResponse>({ url: 'api/v1/service/order' }),
		gcTime: 1000 * 60 * 30, // 30 minutes.
		initialData: { order: monitorData.order },
	});
	useEffect(() => {
		if (orderData)
			setMonitorData({
				page: 'APPROVALS',
				type: 'SERVICE',
				sub_type: 'ORDER',
				...orderData,
			});
	}, [orderData]);

	const filteredServices = useMemo(() => {
		const search = (toolbarOptions.search ?? '').toLowerCase();
		const tags = toolbarOptions.tags ?? [];
		return Object.values(monitorData.order).filter((service_id) => {
			const name = monitorData.service[service_id]?.name ?? service_id;
			const hasTags =
				tags.length === 0 ||
				tags.some((tag) =>
					monitorData.service[service_id]?.tags?.includes(tag),
				);

			if (
				monitorData.service[service_id] &&
				hasTags &&
				name.toLowerCase().includes(search)
			) {
				const svc = monitorData.service[service_id];
				const skipped =
					`SKIP_${svc.status?.latest_version}` === svc.status?.approved_version;
				const upToDate =
					svc.status?.deployed_version === svc.status?.latest_version;
				return (
					// hideUpToDate - deployed_version NOT latest_version.
					(!toolbarOptions.hide.includes(0) || !upToDate) &&
					// hideUpdatable - deployed_version IS latest_version AND approved_version IS NOT "SKIP_"+latest_version.
					(!toolbarOptions.hide.includes(1) || upToDate || skipped) &&
					// hideSkipped - approved_version NOT "SKIP_"+latest_version OR NO approved_version.
					(!toolbarOptions.hide.includes(2) || !skipped) &&
					// hideInactive - active NOT false.
					(!toolbarOptions.hide.includes(3) || svc.active !== false)
				);
			}
		});
	}, [toolbarOptions, monitorData.service, monitorData.order]);

	return (
		<>
			<ApprovalsToolbar
				values={toolbarOptions}
				onEditModeToggle={(value: boolean) => !value && resetOrder()}
				onSaveOrder={handleSaveOrder}
				hasOrderChanged={hasOrderChanged}
			/>
			<Container
				fluid
				className="services"
				style={{
					maxWidth:
						filteredServices.length < 5
							? `${filteredServices.length * 30}rem`
							: '',
				}}
			>
				<DndContext
					sensors={sensors}
					collisionDetection={closestCenter}
					onDragEnd={handleDragEnd}
					autoScroll={{
						enabled: true,
						acceleration: 100,
						threshold: {
							x: 0.2, // Start scrolling when within 20% of the edge
							y: 0.2,
						},
						interval: 5,
					}}
				>
					<SortableContext items={monitorData.order}>
						{monitorData.order.length ===
							Object.keys(monitorData.service).length &&
							filteredServices.map((service_id) => (
								<Service
									key={service_id}
									id={service_id}
									service={monitorData.service[service_id]}
									editable={toolbarOptions.editMode}
								/>
							))}
					</SortableContext>
				</DndContext>
			</Container>
		</>
	);
};
