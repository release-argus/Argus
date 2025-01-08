import { Dispatch } from 'react';
import { WebSocketResponse } from 'types/websocket';

/**
 * Handles service change events from the WebSocket.
 *
 * @param action - The action to handle.
 * @param reducer - The reducer to dispatch the action.
 */
export function handleMessage(
	action: WebSocketResponse,
	reducer: Dispatch<WebSocketResponse>,
) {
	if (
		action.page === 'APPROVALS' &&
		['SERVICE', 'VERSION', 'EDIT', 'DELETE'].includes(action.type)
	) {
		reducer(action);
	}
}
