/* eslint-disable @typescript-eslint/no-non-null-assertion */
import { MonitorSummaryType } from "types/summary";
import { websocketResponse } from "types/websocket";

export default function reducerMonitor(
  state: MonitorSummaryType,
  action: websocketResponse
): MonitorSummaryType {
  switch (action.type) {
    // INIT
    // ORDERING
    // CHANGE
    case "SERVICE":
      switch (action.sub_type) {
        case "INIT":
          state = JSON.parse(JSON.stringify(state));
          if (action.service_data) {
            action.service_data.loading = false;
            state.service[action.service_data.id] = action.service_data;
          }
          break;

        case "ORDERING":
          if (action.order === undefined) {
            break;
          }
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

        case "CHANGE":
          if (
            action?.service_data?.id &&
            action.service_data.id in state.service
          ) {
            return state;
          }
          break;

        default:
          console.log(action);
          throw new Error();
      }
      return state;

    // QUERY
    // NEW
    // UPDATED
    // INIT
    case "VERSION":
      const id = action.service_data?.id as string;
      if (state.service[id] === undefined) {
        return state;
      }
      switch (action.sub_type) {
        case "QUERY":
          // last_queried
          state.service[id].status!.last_queried =
            action?.service_data?.status?.last_queried;
          break;

        case "NEW":
          // url
          state.service[id].url =
            action.service_data?.url || state.service[id].url;

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
          // current_version
          state.service[id].status!.current_version =
            action.service_data?.status?.current_version;
          state.service[id].status!.latest_version =
            action.service_data?.status?.current_version;

          // current_version_timestamp
          state.service[id].status!.current_version_timestamp =
            action.service_data?.status?.current_version_timestamp;
          state.service[id].status!.latest_version_timestamp =
            action.service_data?.status?.current_version_timestamp;
          state.service[id].status!.last_queried =
            action.service_data?.status?.current_version_timestamp;

          // ?url
          if (action.service_data?.url) {
            state.service[id]!.url = action.service_data.url;
          }
          break;

        case "INIT":
          // latest_version
          state.service[id].status!.current_version =
            action.service_data?.status?.latest_version;
          state.service[id].status!.latest_version =
            action.service_data?.status?.latest_version;

          // latest_version_timestamp
          state.service[id].status!.current_version_timestamp =
            action.service_data?.status?.latest_version_timestamp;
          state.service[id].status!.latest_version_timestamp =
            action.service_data?.status?.latest_version_timestamp;
          state.service[id].status!.last_queried =
            action.service_data?.status?.latest_version_timestamp;

          // ?url
          if (action.service_data?.url) {
            state.service[id]!.url = action.service_data.url;
          }
          break;

        case "SKIPPED":
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
    case "RESET":
      state = {
        order: [],
        service: {},
      };
      return state;

    default:
      console.log(action);
      throw new Error();
  }
}
