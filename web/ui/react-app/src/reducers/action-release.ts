// noinspection JSUnfilteredForInLoop

import type { WebSocketResponse } from '@/types/websocket';
import type { ActionModalData } from '@/utils/api/types/config/summary';

/**
 * A reducer that handles actions on the action modal.
 *
 * @param state - The current state of the action modal.
 * @param action - The action to perform on the action modal.
 * @returns The new state of the action modal.
 */
export default function reducerActionModal(
	state: ActionModalData,
	action: WebSocketResponse,
): ActionModalData {
	const newState: ActionModalData = structuredClone(state);

	switch (action.type) {
		// EVENT
		// INIT
		case 'COMMAND':
		case 'WEBHOOK':
			if (
				!action.service_data ||
				(!action.webhook_data && !action.command_data)
			)
				return state;

			if (action.sub_type == 'EVENT') {
				if (action.webhook_data)
					for (const webhookID in action.webhook_data) {
						// Remove them from the sending list.
						newState.sentWH.splice(
							newState.sentWH.indexOf(`${action.service_data.id} ${webhookID}`),
							1,
						);

						// Record the success/fail (if current service in modal).
						if (
							action.service_data.id === state.service_id &&
							newState.webhooks[webhookID] !== undefined
						)
							newState.webhooks[webhookID] = {
								failed: action.webhook_data[webhookID].failed,
								next_runnable: action.webhook_data[webhookID].next_runnable,
							};
					}

				if (action.command_data)
					for (const command in action.command_data) {
						// Remove them from the sending list.
						newState.sentC.splice(
							newState.sentC.indexOf(`${action.service_data.id} ${command}`),
							1,
						);

						// Record the success/fail (if current service in modal).
						if (
							action.service_data.id === state.service_id &&
							newState.commands[command] !== undefined
						)
							newState.commands[command] = {
								failed: action.command_data[command].failed,
								next_runnable: action.command_data[command].next_runnable,
							};
					}
				break;
			} else {
				console.error(action);
				return state;
			}

		// REFRESH
		// RESET
		// SENDING
		case 'ACTION':
			switch (action.sub_type) {
				case 'REFRESH':
					newState.service_id = action.service_data?.id ?? newState.service_id;
					newState.commands = action.command_data ?? {};
					newState.webhooks = action.webhook_data ?? {};
					break;
				case 'RESET':
					newState.commands = {};
					newState.webhooks = {};
					break;
				case 'SENDING':
					// Commands.
					if (action.command_data)
						for (const command in action.command_data) {
							// reset the failed states.
							if (newState.commands[command])
								newState.commands[command].failed = undefined;
							// set as sending.
							newState.sentC.push(`${action.service_data?.id} ${command}`);
						}

					// WebHooks.
					if (action.webhook_data)
						for (const webhookID in action.webhook_data) {
							// reset the failed states.
							if (newState.webhooks[webhookID])
								newState.webhooks[webhookID].failed = undefined;
							// set as sending.
							newState.sentWH.push(`${action.service_data?.id} ${webhookID}`);
						}
					break;
				default:
					console.error(action);
					return state;
			}
			break;

		default:
			console.error(action);
			return state;
	}

	// Got to update the state more for the reload.
	state = newState;
	return state;
}
