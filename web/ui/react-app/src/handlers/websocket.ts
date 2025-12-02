import type { Dispatch } from 'react';
import type { WebSocketResponse } from '@/types/websocket';

/**
 * Handle service change events from the WebSocket.
 *
 * @param msg - The message to handle.
 * @param reducer - The reducer to dispatch the message.
 */
export const handleMessage = (
	msg: WebSocketResponse,
	reducer: Dispatch<WebSocketResponse>,
) => {
	if (
		msg.page === 'APPROVALS' &&
		['SERVICE', 'VERSION', 'EDIT', 'DELETE'].includes(msg.type)
	) {
		reducer(msg);
	}
};
