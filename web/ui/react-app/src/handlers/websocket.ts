import type { Dispatch } from 'react';
import type { WebSocketResponse } from '@/types/websocket';

/**
 * Handle service change events from the WebSocket.
 *
 * @param action - The action to handle.
 * @param reducer - The reducer to dispatch the action.
 */
export const handleMessage = (
	action: WebSocketResponse,
	reducer: Dispatch<WebSocketResponse>,
) => {
	if (
		action.page === 'APPROVALS' &&
		['SERVICE', 'VERSION', 'EDIT', 'DELETE'].includes(action.type)
	) {
		reducer(action);
	}
};
