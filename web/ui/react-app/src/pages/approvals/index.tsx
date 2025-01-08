import { ApprovalsToolbar, Service } from 'components/approvals';
import { ReactElement, useEffect, useMemo, useState } from 'react';

import { ApprovalsToolbarOptions } from 'types/util';
import { Container } from 'react-bootstrap';
import { OrderAPIResponse } from 'types/summary';
import { fetchJSON } from 'utils';
import useLocalStorage from 'hooks/local-storage';
import { useQuery } from '@tanstack/react-query';
import { useWebSocket } from 'contexts/websocket';

/**
 * @returns The approvals page, which includes a toolbar, and a list of services.
 */
export const Approvals = (): ReactElement => {
	const { monitorData, setMonitorData } = useWebSocket();
	const toolbarDefaults: ApprovalsToolbarOptions = {
		search: '',
		editMode: false,
		hide: [3],
	};
	const [toolbarOptionsLS, setLSToolbarOptions] =
		useLocalStorage<ApprovalsToolbarOptions>('toolbarOptions', toolbarDefaults);
	const [toolbarOptions, setToolbarOptions] = useState(toolbarOptionsLS);
	const {
		data: orderData,
		isFetched: isFetchedOrder,
		isFetching: isFetchingOrder,
	} = useQuery({
		queryKey: ['service/order'],
		queryFn: () => fetchJSON<OrderAPIResponse>({ url: 'api/v1/service/order' }),
		gcTime: 1000 * 60 * 30, // 30 minutes.
		initialData: { order: monitorData.order },
	});
	useEffect(() => {
		if (isFetchedOrder && !isFetchingOrder)
			setMonitorData({
				page: 'APPROVALS',
				type: 'SERVICE',
				sub_type: 'ORDER',
				...orderData,
			});
	}, [orderData]);

	// Keep local storage and state in sync.
	useEffect(() => {
		setLSToolbarOptions({
			search: toolbarDefaults.search,
			editMode: toolbarOptions.editMode,
			hide: toolbarOptions.hide,
		});
	}, [toolbarOptions]);

	const filteredServices = useMemo(() => {
		const search = toolbarOptions.search.toLowerCase();
		return Object.values(monitorData.order)
			.filter((service_id) => {
				const name = monitorData.service[service_id]?.name ?? service_id;
				if (
					name.toLowerCase().includes(search) &&
					monitorData.service[service_id]
				) {
					const svc = monitorData.service[service_id];
					const skipped =
						`SKIP_${svc.status?.latest_version}` ===
						svc.status?.approved_version;
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
			})
			.map((service) => monitorData.service[service]);
	}, [toolbarOptions, monitorData.service, monitorData.order]);

	return (
		<>
			<ApprovalsToolbar values={toolbarOptions} setValues={setToolbarOptions} />
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
				{monitorData.order.length === Object.keys(monitorData.service).length &&
					filteredServices.map((service) => (
						<Service
							key={service.id}
							service={service}
							editable={toolbarOptions.editMode}
						/>
					))}
			</Container>
		</>
	);
};
