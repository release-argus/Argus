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
						state = addService(state, action.service_data, undefined, false);
					}
					break;
				}
				case 'ORDER': {
					if (action.order === undefined) break;

					// Remove any services not in the new ordering.
					const newOrder = new Set<string>(state.order);
					state.order.forEach((id) => {
						if (!newOrder.has(id)) state = removeService(state, id);
					});

					state = {
						...state,
						order: action.order,
						service: action.order.reduce(
							(acc: { [key: string]: ServiceSummaryType }, id: string) => {
								acc[id] = state.service[id] ?? { id: id, loading: true };
								return acc;
							},
							{},
						),
					};

					// Fetch data for services that are new in this ordering.
					action.order.forEach((id) => {
						if (state.service[id]?.loading)
							fetchJSON<ServiceSummaryType | undefined>({
								url: `api/v1/service/summary/${encodeURIComponent(id)}`,
							}).then((data) => {
								if (data) {
									data.id = id; // Ensure the ID is set.
									state = addService(state, data, undefined, false);
								}
							});
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

		// NEW
		// INIT
		// QUERY
		// UPDATED
		// ACTION
		case 'VERSION': {
			const id = action.service_data.id;
			if (state.service[id] === undefined) return state;

			// Update service[id].status object.
			const service = {
				...state.service[id],
				id: id,
				status: {
					...(state.service[id].status ?? {}),
					...action.service_data.status,
				},
			};

			if (action.sub_type === 'INIT') {
				// Default deployed_version to latest_version.
				service.status.deployed_version =
					service.status.deployed_version ??
					action.service_data?.status?.latest_version;
				service.status.deployed_version_timestamp =
					service.status.deployed_version_timestamp ??
					action.service_data?.status?.latest_version_timestamp;
			}

			state = addService(state, service, undefined, false);

			return state;
		}

		case 'EDIT': {
			let newServiceData = action.service_data;
			if (newServiceData === undefined) {
				console.error('No service data');
				return state;
			}
			let oldService: ServiceSummaryType | undefined;
			let editedService: ServiceSummaryType;

			// Editing an existing service.
			if (action.sub_type !== undefined) {
				oldService = state.service[action.sub_type];
				// Check this service exists.
				if (oldService === undefined) {
					console.error(`Service ${action.sub_type} does not exist`);
					return state;
				}

				// Update the vars of this service.
				editedService = {
					...oldService,
					...action.service_data,
					name: newServiceData.name || undefined,
					status: {
						...newServiceData.status,
						...oldService.status,
					},
				};
				// Create, conflict with another service.
			} else if (state.service[newServiceData.id] !== undefined) {
				console.error(`Service ${newServiceData.id} already exists`);
				return state;
				// Create a new service.
			} else {
				editedService = {
					...newServiceData,
					loading: false,
				};
			}

			// Ensure the ID is set.
			editedService.id = newServiceData.id ?? oldService?.id;

			state = addService(state, editedService, oldService);

			return state;
		}

		case 'DELETE': {
			if (action.sub_type == undefined) {
				console.error('No sub_type for DELETE');
				return state;
			}

			// Remove the service from the state.
			const service = state.service[action.sub_type];
			if (service !== undefined) state = removeService(state, service.id);

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
 * @param state - The current state of the monitor.
 * @param service_data - The data to give the service.
 * @param old_service_data - The data of the service being modified.
 * @param add_to_order - Optional. Whether to add the service ID to the order array. Defaults to true.
 * @returns The updated monitor state with the new service added.
 */
const addService = (
	state: MonitorSummaryType,
	service_data: ServiceSummaryType,
	old_service_data?: ServiceSummaryType,
	add_to_order = true,
) => {
	// Add the ID to the order array?
	let newOrder: string[] | undefined;
	if (add_to_order) {
		// New service.
		if (old_service_data === undefined) {
			newOrder = [...state.order, service_data.id];
			// Edit service, but ID has changed.
		} else if (service_data.id !== old_service_data?.id) {
			// Replace the old ID with the new ID.
			newOrder = [...state.order];
			newOrder[state.order.indexOf(old_service_data.id)] = service_data.id;

			// Remove the old service.
			state = removeService(
				state,
				old_service_data.id,
				service_data.tags ?? [],
			);
		}
	}

	// Create the `service` dict, with the new service.
	const newService = {
		...state.service,
		[service_data.id]: { ...service_data, loading: false },
	};

	// Add the name to the `names` set.
	let newNames: Set<string> | undefined;
	if (service_data.name !== old_service_data?.name) {
		newNames = new Set(state.names);
		old_service_data?.name && newNames.delete(old_service_data?.name);
		service_data.name && newNames.add(service_data.name);
	}

	// Add the tags to the `tags` set.
	const newState = updateTags(
		state,
		service_data.tags,
		old_service_data?.tags,
		service_data.id,
	);

	return {
		...newState,
		names: newNames ?? state.names,
		order: newOrder ?? state.order,
		service: newService,
	};
};

/**
 * Removes a service from the state.
 *
 * @param id - The ID of the service to remove.
 * @param state - The current state of the monitor.
 * @param tags_of_replacement_service - The tags of the service that will replace the removed service.
 * 	(empty array if none. undefined if no replacement)
 * @returns The new state with the service removed.
 */
const removeService = (
	state: MonitorSummaryType,
	id: string,
	tags_of_replacement_service?: string[],
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

	// Remove any tags that are now unused.
	state = updateTags(
		state,
		state.service[id].tags,
		tags_of_replacement_service,
		id,
	);

	// Remove this service from the order array if it's not being replaced.
	const newOrder =
		tags_of_replacement_service === undefined
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
 * @param serviceID - The ID of the service whose tags are being updated.
 * @returns The new state with the updated tags.
 */
const updateTags = (
	state: MonitorSummaryType,
	newTags: string[] | undefined,
	oldTags: string[] | undefined,
	serviceID: string,
) => {
	const oldServiceTags = new Set(oldTags ?? []);
	const newServiceTags = newTags ?? [];
	if (oldServiceTags.size === 0 && newServiceTags.length === 0) return state;

	// Precompute the tags in use by other services.
	const usedTags = new Set<string>();
	Object.values(state.service).forEach((service) => {
		if (service.id === serviceID) return;
		return service.tags?.forEach((tag) => usedTags.add(tag));
	});

	// Add tags that aren't already in use.
	let tagsAdded = false;
	newServiceTags.forEach((tag) => {
		// New tag.
		if (!state.tags?.has(tag)) {
			tagsAdded = true;
		}
		usedTags.add(tag);
	});

	// If a new tag was added, or the number of tags changed, update the state.
	if (tagsAdded || (state.tags && usedTags.size !== state.tags.size)) {
		return {
			...state,
			tags: usedTags,
		};
	}
	return state;
};
