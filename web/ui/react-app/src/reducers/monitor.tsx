/* eslint-disable @typescript-eslint/no-non-null-assertion */
import { MonitorSummaryType, ServiceSummaryType } from "types/summary";

import { WebSocketResponse } from "types/websocket";
import { fetchJSON } from "utils";

/**
 * Returns a reducer that handles actions on the monitor.
 *
 * @param state - The current state of the monitor
 * @param action - The action to perform on the monitor
 * @returns The new state of the monitor WebSocket
 */
export default function reducerMonitor(
  state: MonitorSummaryType,
  action: WebSocketResponse
): MonitorSummaryType {
  switch (action.type) {
    // INIT
    // ORDER
    case "SERVICE":
      switch (action.sub_type) {
        case "INIT":
          state = JSON.parse(JSON.stringify(state));
          if (action.service_data) {
            action.service_data.loading = false;
            state.service[action.service_data.id] = action.service_data;
          }
          break;

        case "ORDER":
          if (action.order === undefined) break;
          const newState: MonitorSummaryType = {
            order: action.order,
            service: {},
          };
          for (const key of newState.order) {
            newState.service[key] = state.service[key]
              ? state.service[key]
              : { id: key, loading: true };
          }
          state = newState;
          break;

        default:
          console.error(action);
          throw new Error();
      }
      return state;

    // QUERY
    // NEW
    // UPDATED
    // INIT
    case "VERSION":
      const id = action.service_data?.id as string;
      if (state.service[id] === undefined) return state;
      switch (action.sub_type) {
        case "QUERY":
          if (state.service[id]?.status === undefined) return state;

          // last_queried
          state.service[id].status!.last_queried =
            action.service_data?.status?.last_queried;
          break;

        case "NEW":
          // url
          state.service[id].url =
            action.service_data?.url || state.service[id].url;

          // status
          state.service[id].status = state.service[id].status ?? {};

          // latest_version
          state.service[id].status!.latest_version =
            action.service_data?.status?.latest_version;

          // latest_version_timestamp
          state.service[id].status!.latest_version_timestamp =
            action.service_data?.status?.latest_version_timestamp;
          state.service[id].status!.last_queried =
            action.service_data?.status?.latest_version_timestamp;
          break;

        case "UPDATED":
          // status
          state.service[id].status = state.service[id].status ?? {};

          // deployed_version
          state.service[id].status!.deployed_version =
            action.service_data?.status?.deployed_version;
          state.service[id].status!.approved_version =
            action.service_data?.status?.deployed_version;

          // deployed_version_timestamp
          state.service[id].status!.deployed_version_timestamp =
            action.service_data?.status?.deployed_version_timestamp;
          break;

        case "INIT":
          // Check we have the service
          if (
            state.service[id] === undefined ||
            action.service_data?.status === undefined
          )
            return state;

          state.service[id].status = state.service[id].status ?? {};

          // latest_version
          state.service[id].status!.latest_version =
            action.service_data?.status?.latest_version;
          state.service[id].status!.latest_version_timestamp =
            action.service_data?.status?.latest_version_timestamp;
          // last_queried
          state.service[id].status!.last_queried =
            action.service_data?.status?.latest_version_timestamp;

          // deployed_version
          state.service[id].status!.deployed_version =
            state.service[id].status!.deployed_version ||
            action.service_data?.status?.latest_version;
          state.service[id].status!.deployed_version_timestamp =
            state.service[id].status!.deployed_version_timestamp ||
            action.service_data?.status?.latest_version_timestamp;

          // url
          state.service[id].url =
            action.service_data?.url || state.service[id].url;

          break;

        case "ACTION":
          if (state.service[id]?.status === undefined) return state;

          // approved_version
          state.service[id].status!.approved_version =
            action.service_data?.status?.approved_version;

          break;

        default:
          return state;
      }

      // Gotta update the state more for the reload
      state = JSON.parse(JSON.stringify(state));

      return state;

    case "EDIT":
      let service = action.service_data;
      if (service === undefined) {
        console.error("No service data");
        return state;
      }

      // If we're editing an existing service
      if (action.sub_type !== undefined) {
        service = state.service[action.sub_type];
        // Check this service exists
        if (service === undefined) {
          console.error(`Service ${action.sub_type} does not exist`);
          return state;
        }

        // Update the vars of this service
        service.id = action.service_data?.id || service.id;
        service.active = action.service_data?.active ?? service.active;
        service.type = action.service_data?.type || service.type;
        service.url = action.service_data?.url || service.url;
        service.icon = action.service_data?.icon || service.icon;
        service.icon_link_to =
          action.service_data?.icon_link_to || service.icon_link_to;
        service.has_deployed_version =
          action.service_data?.has_deployed_version ??
          service.has_deployed_version;
        service.command = action.service_data?.command ?? service.command;
        service.webhook = action.service_data?.webhook ?? service.webhook;
        // status
        service.status!.approved_version =
          action.service_data?.status?.approved_version ||
          service.status!.approved_version;
        service.status!.deployed_version =
          action.service_data?.status?.deployed_version ||
          service.status!.deployed_version;
        service.status!.deployed_version_timestamp =
          action.service_data?.status?.deployed_version_timestamp ||
          service.status!.deployed_version_timestamp;
        service.status!.latest_version =
          action.service_data?.status?.latest_version ||
          service.status!.latest_version;
        service.status!.latest_version_timestamp =
          action.service_data?.status?.latest_version_timestamp ||
          service.status!.latest_version_timestamp;
        service.status!.last_queried =
          action.service_data?.status?.last_queried ||
          service.status!.last_queried;
        // create and the service already exists
      } else if (state.service[service.id] !== undefined) {
        console.error(`Service ${service.id} already exists`);
        return state;
      }

      service.loading = false;
      state.service[service.id] = service;

      // If the service has been renamed, we need to update the order
      if (service.id !== action.sub_type && action.sub_type !== undefined) {
        delete state.service[action.sub_type];
        state.order[state.order.indexOf(action.sub_type)] = service.id;

        // If the service is new, we need to add it to the order
      } else action.sub_type === undefined && state.order.push(service.id);

      // Gotta update the state more for the reload
      state = JSON.parse(JSON.stringify(state));

      return state;

    case "DELETE":
      if (action.sub_type == undefined) {
        console.error("No sub_type for DELETE");
        return state;
      }
      if (action.order === undefined) {
        console.error("No order for DELETE");
        return state;
      }

      // Remove the service from the state
      if (state.service[action.sub_type] !== undefined)
        delete state.service[action.sub_type];
      state.order = action.order;

      // Check whether we've missed any other removals
      for (const id in state.service) {
        if (!action.order.includes(id)) delete state.service[id];
      }

      // Check whether we've missed any additions
      for (const id of action.order) {
        if (state.service[id] === undefined)
          fetchJSON<ServiceSummaryType | undefined>({
            url: `api/v1/service/summary/${encodeURIComponent(id)}`,
          }).then((data) => {
            if (data) state.service[id] = data;
          });
      }

      // Gotta update the state more for the reload
      state = JSON.parse(JSON.stringify(state));

      return state;

    default:
      console.error(action);
      throw new Error();
  }
}
