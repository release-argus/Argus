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
					if (action.service_data) {
						state = addService(
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
					state = {
						order: action.order,
						names: state.names ?? new Set<string>(),
						tags: state.tags ?? new Set<string>(),
						service: state.service ?? {},
					};
					// Default each undefined service.
					state.order.forEach((id) => {
						if (state.service[id] === undefined)
							state.service[id] = { id: id, loading: true };
					});

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
		// ACTION
		case 'VERSION': {
			const id = action.service_data.id;
			if (state.service[id] === undefined) return state;

			// Update service[id].status object.
			state = {
				...state,
				service: {
					...state.service,
					[id]: {
						...state.service[id],
						status: {
							...(state.service[id].status ?? {}),
							...action.service_data?.status,
						},
					},
				},
			};

			switch (action.sub_type) {
				// case 'QUERY':
				// case 'ACTION':
				// case 'UPDATED':
				case 'NEW': {
					// url
					state.service[id].url =
						action.service_data?.url ?? state.service[id].url;

					break;
				}
				case 'INIT': {
					// Default deployed_version to latest_version.
					if (state.service[id].status) {
						state.service[id].status.deployed_version =
							state.service[id].status.deployed_version ??
							action.service_data?.status?.latest_version;
						state.service[id].status.deployed_version_timestamp =
							state.service[id].status.deployed_version_timestamp ??
							action.service_data?.status?.latest_version_timestamp;
					}

					// url
					state.service[id].url =
						action.service_data?.url ?? state.service[id].url;

					break;
				}
				default: {
					return state;
				}
			}

			addService(id, state, state.service[id], false);

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
					name: action.service_data?.name || undefined,
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
				state = addService(service.id, state, service, true);
			}
			// Edited service, update the data.
			else {
				// Leave the state unchanged if no service changes.
				if (Object.keys(action.service_data ?? {}).length === 0) return state;

				// Renamed service, update the order.
				if (service.id !== action.sub_type || oldName !== service.name) {
					state = removeService(action.sub_type, state, false);
					state.order[state.order.indexOf(action.sub_type)] = service.id;
				}
				state = addService(service.id, state, service, false);
			}

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
			if (service !== undefined)
				state = removeService(action.sub_type, state, true);

			// Check whether we"ve missed any other removals.
			for (const id in state.service) {
				if (!action.order.includes(id)) {
					if (state.service[id] !== undefined)
						state = removeService(id, state, false);
				}
			}

			// Check whether we"ve missed any additions.
			for (const id of action.order) {
				if (state.service[id] === undefined)
					fetchJSON<ServiceSummaryType | undefined>({
						url: `api/v1/service/summary/${encodeURIComponent(id)}`,
					}).then((data) => {
						if (data) state = addService(id, state, data);
					});
			}

			return state;
		}
		default: {
			console.error(action);
			throw new Error();
		}
	}
}

/**
 * Adds a service to the monitor state.
 *
 * @param id - The unique identifier for the service.
 * @param state - The current state of the monitor.
 * @param service_data - The data of the service to be added.
 * @param add_to_order - Optional. Whether to add the service ID to the order array. Defaults to true.
 * @returns The updated monitor state with the new service added.
 */
const addService = (
	id: string,
	state: MonitorSummaryType,
	service_data: ServiceSummaryType,
	add_to_order = true,
) => {
	// Add the name to the names set.
	let newNames: Set<string> | undefined;
	if (service_data.name && !state.names.has(service_data.name)) {
		newNames = new Set(state.names);
		newNames.add(service_data.name);
	}

	// Add the tags to the tags set.
	state = updateTags(state, state.service[id].tags, service_data.tags, id);

	// Add the ID to the order array.
	const newOrder = add_to_order ? [...state.order, id] : state.order;

	// Set the service data.
	const newService = {
		...state.service,
		[id]: { ...service_data, loading: false },
	};

	return {
		...state,
		names: newNames ?? state.names,
		order: newOrder,
		service: newService,
	};
};

/**
 * Removes a service from the state.
 *
 * @param id - The ID of the service to remove.
 * @param state - The current state of the monitor.
 * @param remove_from_order - Whether to remove the service ID from the order array (and remove now-unused tags). Defaults to true.
 * @returns The new state with the service removed.
 */
const removeService = (
	id: string,
	state: MonitorSummaryType,
	remove_from_order = true,
) => {
	// Invalid/unknown service.
	if (!state.service[id]) return state;

	// Remove the name from the names Set.
	let newNames: Set<string> | undefined;
	if (state.service[id]?.name) {
		newNames = new Set(state.names);
		newNames.delete(state.service[id].name);
	}

	// Remove the service from the service object.
	const { [id]: _, ...newService } = state.service;

	// Remove any tags that are now unused (if we're updating ordering).
	state = updateTags(state, state.service[id].tags, [], id);

	// Remove the ID from the order array.
	const newOrder = remove_from_order
		? state.order.filter((item) => item !== id)
		: state.order;

	return {
		...state,
		names: newNames ?? state.names,
		order: newOrder,
		service: newService,
	};
};

/**
 * Updates the global tags set with the changes to a service's tags.
 *
 * @param state - The current state of the monitor summary.
 * @param oldTags - The previous tags associated with the service.
 * @param newTags - The new tags to be associated with the service.
 * @param targetServiceId - The ID of the service whose tags are being updated.
 * @returns The new state with the updated tags.
 */
const updateTags = (
	state: MonitorSummaryType,
	oldTags: string[] | undefined,
	newTags: string[] | undefined,
	targetServiceId: string,
) => {
	const oldServiceTags = new Set(oldTags ?? []);
	const newServiceTags = newTags ?? [];
	if (oldServiceTags.size === 0 && newServiceTags.length === 0) return state;

	// Precompute the tags in use by other services.
	const usedTags = new Set<string>();
	Object.values(state.service).forEach((service) => {
		if (service.id === targetServiceId) return;
		return service.tags?.forEach((tag) => usedTags.add(tag));
	});

	// Add tags that aren't already in use.
	let tagsAdded = false;
	newServiceTags.forEach((tag) => {
		// New tag.
		if (!state.tags.has(tag)) {
			tagsAdded = true;
		}
		usedTags.add(tag);
	});

	// If a new tag was added, or the number of tags changed, update the state.
	if (tagsAdded || usedTags.size !== state.tags.size) {
		return {
			...state,
			tags: usedTags,
		};
	}
	return state;
};
