import {
	ActionAPIType,
	MonitorSummaryType,
	OrderAPIResponse,
	ServiceSummaryType,
} from 'types/summary';
import {
	Dispatch,
	JSX,
	createContext,
	useContext,
	useEffect,
	useMemo,
	useReducer,
	useState,
} from 'react';
import { compareStringArrays, fetchJSON, getBasename } from 'utils';
import { useQuery, useQueryClient } from '@tanstack/react-query';

import { BooleanType } from 'types/boolean';
import ReconnectingWebSocket from 'reconnecting-websocket';
import { WS_ADDRESS } from 'config';
import { WebSocketResponse } from 'types/websocket';
import { WebSocketStatus } from 'components/websocket/status';
import { handleMessage } from 'handlers/websocket';
import reducerMonitor from 'reducers/monitor';

type Bool = boolean | undefined;
type Socket = ReconnectingWebSocket | undefined;
interface WebSocketCtx {
	ws: Socket;
	connected: BooleanType;
	monitorData: MonitorSummaryType;
	setMonitorData: Dispatch<WebSocketResponse>;
}

/**
 * Provides the WebSocket connection and monitor data.
 *
 * @param ws - The WebSocket connection.
 * @param connected - Whether the WebSocket connection is established.
 * @param monitorData - The monitor data.
 * @param setMonitorData - Function to set the monitor data.
 * @returns The WebSocket context.
 */
export const WebSocketContext = createContext<WebSocketCtx>({
	ws: undefined,
	connected: false,
	monitorData: {
		order: [],
		names: new Set<string>(),
		tags: new Set<string>(),
		service: {},
	},
	// eslint-disable-next-line @typescript-eslint/no-empty-function
	setMonitorData: () => {},
});

interface Props {
	children: JSX.Element[];
}

const ws = new ReconnectingWebSocket(`${WS_ADDRESS}${getBasename()}/ws`);
/**
 * @returns The WebSocket connection and monitor data.
 */
export const WebSocketProvider = (props: Props) => {
	const queryClient = useQueryClient();
	const [monitorData, setMonitorData] = useReducer(reducerMonitor, {
		order: ['monitorData_loading'],
		names: new Set<string>(),
		tags: new Set<string>(),
		service: {},
	});
	const [connected, setConnected] = useState<Bool>(undefined);

	const contextValue = useMemo(
		() => ({
			ws: ws,
			connected: connected,
			monitorData: monitorData,
			setMonitorData: setMonitorData,
		}),
		[connected, monitorData],
	);

	const { data: orderData, isFetching: orderIsFetching } = useQuery({
		queryKey: ['service/order'],
		queryFn: () => fetchJSON<OrderAPIResponse>({ url: 'api/v1/service/order' }),
		gcTime: 1000 * 60 * 30, // 30 minutes.
	});
	useEffect(() => {
		// Not a disconnect, still fetching, or no ordering.
		if (
			connected === false ||
			orderIsFetching ||
			orderData?.order === undefined
		)
			return;

		// Only if the order has changed.
		if (!compareStringArrays(orderData.order, monitorData.order)) {
			setMonitorData({
				page: 'APPROVALS',
				type: 'SERVICE',
				sub_type: 'ORDER',
				...orderData,
			});
		}

		orderData.order.forEach((service) => {
			// If service already cached, do not refetch.
			if (monitorData.service[service]?.status?.latest_version_timestamp)
				return;
			fetchJSON<ServiceSummaryType | undefined>({
				url: `api/v1/service/summary/${encodeURIComponent(service)}`,
			}).then((data) => {
				if (data)
					setMonitorData({
						page: 'APPROVALS',
						type: 'SERVICE',
						sub_type: 'INIT',
						service_data: data,
					});
			});
		});
	}, [orderData, connected]);

	ws.onopen = () => {
		// Invalidate the cache if not the first connect event.
		if (connected !== undefined)
			queryClient.invalidateQueries({
				queryKey: ['service/order'],
			});
		setConnected(true);
	};

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	ws.onmessage = (event: any) => {
		if (event.data === '') return;
		// Validate the JSON
		if (event.data.length > 1 && event.data[0] == '{') {
			const msg = JSON.parse(event.data.trim()) as WebSocketResponse;
			handleMessage(msg, setMonitorData);
			// update/invalidate caches.
			if (msg.page === 'APPROVALS') {
				if (msg.type === 'EDIT') {
					queryClient.invalidateQueries({
						queryKey: ['actions', { service: msg.sub_type }],
					});
					queryClient.invalidateQueries({
						queryKey: ['service/edit', { service: msg.sub_type }],
					});
				}

				if (
					(msg.type === 'COMMAND' || msg.type === 'WEBHOOK') &&
					msg.sub_type === 'EVENT'
				) {
					const queryKey = ['actions', { service: msg.service_data?.id }];
					const queryData = queryClient.getQueryData(queryKey);
					if (queryData !== undefined) {
						if (msg.command_data)
							for (const command in msg.command_data) {
								// store it in the cache.
								(queryData as ActionAPIType).command[command] = {
									failed: msg.command_data[command].failed,
									next_runnable: msg.command_data[command].next_runnable,
								};
							}

						if (msg.webhook_data)
							for (const webhook_id in msg.webhook_data) {
								// store it in the cache.
								(queryData as ActionAPIType).webhook[webhook_id] = {
									failed: msg.webhook_data[webhook_id].failed,
									next_runnable: msg.webhook_data[webhook_id].next_runnable,
								};
							}
						queryClient.setQueryData(queryKey, queryData);
					}
				}
			}

			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			messageHandlers.forEach((item: { handler: any; params?: any }) =>
				item.params
					? item.handler({
							event: msg,
							...item.params,
						})
					: item.handler(msg),
			);
		}
	};

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	ws.onerror = (event: any) => {
		connected && setConnected(false);
		console.error('ws err', event);
	};

	return (
		<WebSocketContext.Provider value={contextValue}>
			<WebSocketStatus connected={connected} />
			{props.children}
		</WebSocketContext.Provider>
	);
};

export const sendMessage = (data: string) => {
	ws.send(data);
};

const messageHandlers = new Map();

export const addMessageHandler = (
	id: string,
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	handler: any,
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	params?: any,
): void => {
	messageHandlers.set(id, { handler: handler, params: params });
};

export const removeMessageHandler = (id: string) => {
	messageHandlers.delete(id);
};

export const useWebSocket = () => {
	return useContext(WebSocketContext);
};
