import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
	createContext,
	type Dispatch,
	type ReactNode,
	use,
	useEffect,
	useMemo,
	useReducer,
	useState,
} from 'react';
import ReconnectingWebSocket from 'reconnecting-websocket';
import { WebSocketStatus } from '@/components/websocket/status';
import { WS_ADDRESS } from '@/config';
import { handleMessage } from '@/handlers/websocket';
import { QUERY_KEYS } from '@/lib/query-keys';
import reducerMonitor from '@/reducers/monitor';
import type { WebSocketResponse } from '@/types/websocket';
import { compareStringArrays, getBasename } from '@/utils';
import { mapRequest } from '@/utils/api/types/api-request-handler';
import type {
	ActionAPIType,
	MonitorSummaryType,
} from '@/utils/api/types/config/summary';

type WebSocketContextProps = {
	/* The WebSocket connection. */
	ws?: ReconnectingWebSocket;
	/* Whether the WebSocket connection is established. */
	connected?: boolean;
	/* The service monitor data. */
	monitorData: MonitorSummaryType;
	/* Function to set monitor data. */
	setMonitorData: Dispatch<WebSocketResponse>;
};

/**
 * Provides the WebSocket connection and monitor data.
 *
 * @param ws - The WebSocket connection.
 * @param connected - Whether the WebSocket connection is established.
 * @param monitorData - The monitor data.
 * @param setMonitorData - Function to set the monitor data.
 * @returns The WebSocket context.
 */
export const WebSocketContext = createContext<WebSocketContextProps>({
	connected: false,
	monitorData: {
		names: new Set<string>(),
		order: [],
		service: {},
		tagsLoaded: false,
	},
	setMonitorData: () => {
		/* noop */
	},
	ws: undefined,
});

type WebSocketProviderProps = {
	/* The content to wrap. */
	children: ReactNode;
};

const ws = new ReconnectingWebSocket(`${WS_ADDRESS}${getBasename()}/ws`);
/**
 * @returns The WebSocket connection and monitor data.
 */
export const WebSocketProvider = (props: WebSocketProviderProps) => {
	const queryClient = useQueryClient();
	const [monitorData, setMonitorData] = useReducer(reducerMonitor, {
		names: new Set<string>(),
		order: [],
		service: {},
		tagsLoaded: false,
	});
	const [connected, setConnected] = useState<boolean | undefined>(undefined);

	const contextValue = useMemo(
		() => ({
			connected: connected,
			monitorData: monitorData,
			setMonitorData: setMonitorData,
			ws: ws,
		}),
		[connected, monitorData],
	);

	const { data: orderData, isFetching: orderIsFetching } = useQuery({
		gcTime: 1000 * 60 * 30, // 30 minutes.
		queryFn: () => mapRequest('SERVICE_ORDER_GET', null),
		queryKey: QUERY_KEYS.SERVICE.ORDER(),
	});
	// biome-ignore lint/correctness/useExhaustiveDependencies: orderData covers monitorData.order.
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
				sub_type: 'ORDER',
				type: 'SERVICE',
				...orderData,
			});
		}
	}, [orderData, connected]);

	ws.onopen = () => {
		// Invalidate the cache if not the first 'connect' event.
		if (connected !== undefined) {
			void queryClient.invalidateQueries({
				queryKey: QUERY_KEYS.SERVICE.ORDER(),
			});
		}
		setConnected(true);
	};

	ws.onmessage = (event: MessageEvent) => {
		if (typeof event.data !== 'string' || event.data === '') return;
		// Validate the JSON
		if (event.data.length > 1 && event.data.startsWith('{')) {
			const msg = JSON.parse(event.data.trim()) as WebSocketResponse;
			handleMessage(msg, setMonitorData);

			// update/invalidate caches.
			if (msg.page === 'APPROVALS') {
				if (msg.type === 'EDIT' && msg.sub_type) {
					void queryClient.invalidateQueries({
						queryKey: QUERY_KEYS.SERVICE.ACTIONS(msg.sub_type),
					});
					void queryClient.invalidateQueries({
						queryKey: QUERY_KEYS.SERVICE.EDIT_ITEM(msg.sub_type),
					});
				}

				if (msg.type === 'COMMAND' || msg.type === 'WEBHOOK') {
					const queryKey = QUERY_KEYS.SERVICE.ACTIONS(msg.service_data?.id);
					const queryData = queryClient.getQueryData(queryKey);
					if (queryData !== undefined) {
						if (msg.command_data) {
							for (const [commandID, commandData] of Object.entries(
								msg.command_data,
							)) {
								// Store it in the cache.
								(queryData as ActionAPIType).command[commandID] = {
									failed: commandData.failed,
									next_runnable: commandData.next_runnable,
								};
							}
						}

						if (msg.webhook_data) {
							for (const [webhookID, webhookData] of Object.entries(
								msg.webhook_data,
							)) {
								// {
								//  "page":"APPROVALS",
								//  "type":"WEBHOOK",
								//  "sub_type":"EVENT",
								//  "service_data":{"id":"autobrr/autobrr"},
								//  "webhook_data":{"awx":{"failed":false,"next_runnable":"2025-11-07T02:48:14.157942826Z"}}
								// }
								// Store it in the cache.
								(queryData as ActionAPIType).webhook[webhookID] = {
									failed: webhookData.failed,
									next_runnable: webhookData.next_runnable,
								};
							}
						}
						queryClient.setQueryData(queryKey, queryData);
					}
				}
			}

			for (const { handler, params } of messageHandlers.values()) {
				handler(msg, params);
			}
		}
	};

	ws.onerror = (event: unknown) => {
		if (connected) setConnected(false);
		console.error('ws err', event);
	};

	return (
		<WebSocketContext value={contextValue}>
			<WebSocketStatus connected={connected} />
			{props.children}
		</WebSocketContext>
	);
};

export const sendMessage = (data: string) => {
	ws.send(data);
};

export type MessageHandler<P = undefined> = {
	handler: (event: WebSocketResponse, params?: P) => void;
	params?: P;
};
// biome-ignore lint/suspicious/noExplicitAny: any message.
const messageHandlers = new Map<string, MessageHandler<any>>();

export const addMessageHandler = <P = undefined>(
	id: string,
	handlerObj: MessageHandler<P>,
): void => {
	messageHandlers.set(id, handlerObj);
};

export const removeMessageHandler = (id: string) => {
	messageHandlers.delete(id);
};

/**
 * useWebSocket retrieves the WebSocket context value from the WebSocketContext.
 */
export const useWebSocket = () => {
	return use(WebSocketContext);
};
