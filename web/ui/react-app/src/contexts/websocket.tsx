import { useQueryClient } from '@tanstack/react-query';
import { createContext, type ReactNode, use, useMemo, useState } from 'react';
import ReconnectingWebSocket from 'reconnecting-websocket';
import { WebSocketStatus } from '@/components/websocket/status';
import { WS_ADDRESS } from '@/config';
import { handleMessage } from '@/handlers/websocket';
import { QUERY_KEYS } from '@/lib/query-keys';
import type { WebSocketResponse } from '@/types/websocket';
import { getBasename } from '@/utils';
import approvalsQueryCacheUpdater from '@/utils/api/query-cache';

type WebSocketContextProps = {
	/* The WebSocket connection. */
	ws?: ReconnectingWebSocket;
	/* Whether the WebSocket connection is established. */
	connected?: boolean;
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
	const [connected, setConnected] = useState<boolean | undefined>(undefined);

	const contextValue = useMemo(
		() => ({
			connected: connected,
			ws: ws,
		}),
		[connected],
	);

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
			handleMessage(msg, (msg) =>
				approvalsQueryCacheUpdater({ msg, queryClient }),
			);

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
