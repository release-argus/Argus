import { MonitorSummaryType, ServiceSummaryType } from 'types/summary';

import { WebSocketResponse } from 'types/websocket';
import { fetchJSON } from 'utils';

/**
 * A reducer that handles actions on the monitor.
 *
 * @param state - The current state of the monitor.
 * @param action - The action to perform on the monitor.
 * @returns The new state of the monitor.
 */
export default function reducerMonitor(
	state: MonitorSummaryType,
	action: WebSocketResponse,
): MonitorSummaryType {
	switch (action.type) {
		// INIT
		// ORDER
		case 'SERVICE': {
			switch (action.sub_type) {
				case 'INIT': {
					state = {
						...JSON.parse(JSON.stringify(state)),
						names: state.names, // Keep the names set.
						tags: state.tags, // Keep the tags set.
					};
					if (action.service_data) {
						addService(
							action.service_data.id,
							state,
							action.service_data,
							false,
						);
					}
					break;
				}
				case 'ORDER': {
					if (action.order === undefined) break;
					const newState: MonitorSummaryType = {
						order: action.order,
						names: state.names ?? new Set<string>(),
						tags: state.tags ?? new Set<string>(),
						service: state.service ?? {},
					};
					for (const key of newState.order) {
						newState.service[key] = state.service[key]
							? state.service[key]
							: { id: key, loading: true };
					}
					state = newState;
					break;
				}
				default: {
					console.error(action);
					throw new Error();
				}
			}
			return state;
		}

		// QUERY
		// NEW
		// UPDATED
		// INIT
		case 'VERSION': {
			const id = action.service_data.id;
			if (state.service[id] === undefined) return state;
			switch (action.sub_type) {
				case 'QUERY': {
					if (state.service[id]?.status === undefined) return state;

					// last_queried
					state.service[id].status.last_queried =
						action.service_data?.status?.last_queried;
					break;
				}
				case 'NEW': {
					// url
					state.service[id].url =
						action.service_data?.url ?? state.service[id].url;

					// status
					state.service[id].status = state.service[id].status ?? {};
					state.service[id].status = {
						...state.service[id].status,
						...action.service_data?.status,
					};
					break;
				}
				case 'UPDATED': {
					// status
					state.service[id].status = state.service[id].status ?? {};
					state.service[id].status = {
						...state.service[id].status,
						...action.service_data?.status,
					};
					break;
				}
				case 'INIT': {
					// Check we have the service.
					if (
						state.service[id] === undefined ||
						action.service_data?.status === undefined
					)
						return state;

					// status
					state.service[id].status = state.service[id].status ?? {};
					state.service[id].status = {
						...state.service[id].status,
						...action.service_data?.status,
						// Default deployed_version to latest_version.
						deployed_version:
							state.service[id].status.deployed_version ??
							action.service_data?.status?.latest_version,
						deployed_version_timestamp:
							state.service[id].status.deployed_version_timestamp ??
							action.service_data?.status?.latest_version_timestamp,
					};

					// url
					state.service[id].url =
						action.service_data?.url ?? state.service[id].url;

					break;
				}
				case 'ACTION': {
					if (state.service[id]?.status === undefined) return state;

					// approved_version
					state.service[id].status.approved_version =
						action.service_data?.status?.approved_version;

					break;
				}
				default: {
					return state;
				}
			}

			// Got to update the state more for the reload.
			state = {
				...JSON.parse(JSON.stringify(state)),
				names: state.names, // Keep the names set.
				tags: state.tags, // Keep the tags set.
			};

			return state;
		}

		case 'EDIT': {
			let service = action.service_data;
			if (service === undefined) {
				console.error('No service data');
				return state;
			}

			// Editing an existing service.
			if (action.sub_type !== undefined) {
				service = state.service[action.sub_type];
				// Check this service exists.
				if (service === undefined) {
					console.error(`Service ${action.sub_type} does not exist`);
					return state;
				}

				// Update the vars of this service.
				service = {
					...service,
					...action.service_data,
					name: action.service_data?.name ?? undefined,
					icon:
						action.service_data?.icon === '~'
							? undefined
							: action.service_data?.icon ?? service.icon,
					icon_link_to:
						action.service_data?.icon_link_to === '~'
							? undefined
							: action.service_data?.icon_link_to ?? service.icon_link_to,
					url:
						action.service_data?.url === '~'
							? undefined
							: action.service_data?.url ?? service.url,
					status: {
						...action.service_data?.status,
						...service.status,
					},
				};
				// Create, conflict with another service.
			} else if (state.service[service.id] !== undefined) {
				console.error(`Service ${service.id} already exists`);
				return state;
			}

			const oldName = state.service[action.sub_type]?.name;
			// New service, add it to the order.
			if (action.sub_type === undefined) {
				addService(service.id, state, service, true);
			}
			// Edited service, update the data.
			else {
				// Leave the state unchanged if no service changes.
				if (Object.keys(action.service_data ?? {}).length === 0) return state;

				// Renamed service, update the order.
				if (service.id !== action.sub_type || oldName !== service.name) {
					removeService(action.sub_type, state, false);
					state.order[state.order.indexOf(action.sub_type)] = service.id;
				}
				addService(service.id, state, service, false);
			}

			// Got to update the state more for the reload.
			state = {
				...JSON.parse(JSON.stringify(state)),
				names: state.names, // Keep the names set.
				tags: state.tags, // Keep the tags set.
			};

			return state;
		}

		case 'DELETE': {
			if (action.sub_type == undefined) {
				console.error('No sub_type for DELETE');
				return state;
			}
			if (action.order === undefined) {
				console.error('No order for DELETE');
				return state;
			}

			// Remove the service from the state.
			const service = state.service[action.sub_type];
			if (service !== undefined) removeService(action.sub_type, state, true);

			// Check whether we"ve missed any other removals.
			for (const id in state.service) {
				if (!action.order.includes(id)) {
					if (state.service[id] !== undefined) removeService(id, state, false);
				}
			}

			// Check whether we"ve missed any additions.
			for (const id of action.order) {
				if (state.service[id] === undefined)
					fetchJSON<ServiceSummaryType | undefined>({
						url: `api/v1/service/summary/${encodeURIComponent(id)}`,
					}).then((data) => {
						if (data) addService(id, state, data);
					});
			}

			// Got to update the state more for the reload.
			state = {
				...JSON.parse(JSON.stringify(state)),
				names: state.names, // Keep the names set.
				tags: state.tags, // Keep the tags set.
			};

			return state;
		}
		default: {
			console.error(action);
			throw new Error();
		}
	}
}

const addService = (
	id: string,
	state: MonitorSummaryType,
	service_data: ServiceSummaryType,
	add_to_order = true,
) => {
	// Set the service data.
	service_data.loading = false;
	state.service[id] = service_data;
	// Add the name to the names set.
	if (service_data.name) state.names.add(service_data.name);
	// Add the tags to the tags set.
	service_data.tags?.forEach((item) => state.tags.add(item));
	// Add the ID to the order array.
	if (add_to_order) state.order.push(id);
};

const removeService = (
	id: string,
	state: MonitorSummaryType,
	remove_from_order = true,
) => {
	// Remove the name from the names array.
	if (state.service[id].name) state.names.delete(state.service[id].name);
	// Remove the ID from the order array.
	if (remove_from_order) state.order.splice(state.order.indexOf(id), 1);
	// Remove the service from the service object.
	delete state.service[id];
};
