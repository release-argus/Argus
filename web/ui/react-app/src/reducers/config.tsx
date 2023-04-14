import { ConfigState } from "types/config";
import { WebSocketResponse } from "types/websocket";
import { cleanEmpty } from "utils";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const deleteUndefinedProperties = (obj: any) => {
  for (const key in obj) {
    if (obj[key] && typeof obj[key] === "object") {
      deleteUndefinedProperties(obj[key]);
      if (Object.keys(obj[key]).length === 0) {
        delete obj[key];
      }
    } else if (obj[key] === undefined) {
      delete obj[key];
    }
  }
};

export default function reducerConfig(
  state: ConfigState,
  action: WebSocketResponse
): ConfigState {
  state = JSON.parse(JSON.stringify(state));
  if (action.config_data === undefined) {
    return state;
  }
  deleteUndefinedProperties(action);
  switch (action.type) {
    // INIT
    case "SETTINGS":
      switch (action.sub_type) {
        case "INIT":
          if (action.config_data?.settings !== undefined) {
            state.data.settings = action.config_data.settings;
          }
          break;

        default:
          console.log(action);
          throw new Error();
      }
      break;

    // INIT
    case "DEFAULTS":
      switch (action.sub_type) {
        case "INIT":
          state.data.defaults = action.config_data.defaults;
          break;

        default:
          console.log(action);
          throw new Error();
      }
      break;

    // INIT
    case "NOTIFY":
      switch (action.sub_type) {
        case "INIT":
          state.data.notify = action.config_data.notify;
          break;

        default:
          console.log(action);
          throw new Error();
      }
      break;

    // INIT
    case "WEBHOOK":
      switch (action.sub_type) {
        case "INIT":
          state.data.webhook = action.config_data.webhook;
          break;

        default:
          console.log(action);
          throw new Error();
      }
      break;

    // INIT
    case "SERVICE":
      switch (action.sub_type) {
        case "INIT":
          state.data.service = {};
          if (action.config_data?.order && action.config_data.service) {
            for (const service_id of action.config_data.order) {
              state.data.service[service_id] =
                action.config_data.service[service_id];
            }
          }
          break;

        default:
          throw new Error();
      }
      break;

    default:
      console.log(action);
      throw new Error();
  }

  if (action.sub_type === "INIT" && state.waiting_on.includes(action.type)) {
    state.waiting_on = state.waiting_on.filter((item) => item !== action.type);
  }

  state.data = cleanEmpty(state.data);
  return state;
}
