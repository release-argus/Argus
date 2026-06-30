import { useQueryClient } from '@tanstack/react-query';
import { WebSocket } from 'partysocket';
import { createContext, type ReactNode, use, useMemo, useState } from 'react';
import { WebSocketStatus } from '@/components/websocket/status';
import { WS_ADDRESS } from '@/config';
import { handleMessage } from '@/handlers/websocket';
import { QUERY_KEYS } from '@/lib/query-keys';
import type { WebSocketResponse } from '@/types/websocket';
import { getBasename } from '@/utils';
import approvalsQueryCacheUpdater from '@/utils/api/query-cache';
import { API_BASE } from '@/utils/api/types/api-request';

type WebSocketContextProps = {
	/* The WebSocket connection. */
	ws?: WebSocket;
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

/**
 * Resolves the URL for the "/ws" WebSocket endpoint.
 *
 * Safari/WebKit doesn't forward cached HTTP Basic Auth credentials on
 * WebSocket handshake requests, so a short-lived token is fetched from the
 * (Basic Auth protected) "/api/v1/ws-token" endpoint via a normal
 * authenticated request, and passed as a "token" query parameter instead.
 *
 * 204 from "/ws-token" means Basic Auth is not configured.
 *
 * Called on every (re)connection attempt, so a fresh token is used each time.
 */
const getWebSocketURL = async (): Promise<string> => {
	const wsURL = `${WS_ADDRESS}${getBasename()}/ws`;

	try {
		const resp = await fetch(`${API_BASE}/ws-token`);
		// 204 = Basic Auth not configured; connect without a token.
		if (resp.status === 204) return wsURL;
		if (!resp.ok) {
			throw new Error(`Failed to fetch WebSocket token: HTTP ${resp.status}`);
		}
		const { token } = (await resp.json()) as { token: string };
		return `${wsURL}?token=${encodeURIComponent(token)}`;
	} catch (error) {
		console.error('Failed to fetch WebSocket token:', error);
		return wsURL;
	}
};

const ws = new WebSocket(getWebSocketURL);
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
