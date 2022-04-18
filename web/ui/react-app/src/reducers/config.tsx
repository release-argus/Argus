import { ConfigState } from "types/config";
import { websocketResponse } from "types/websocket";

export default function reducerConfig(
  state: ConfigState,
  action: websocketResponse
): ConfigState {
  state = JSON.parse(JSON.stringify(state));
  if (action.config_data === undefined) {
    return state;
  }
  switch (action.type) {
    // INIT
    case "SETTINGS":
      switch (action.sub_type) {
        case "INIT":
          state.data.settings = action.config_data.settings;
          // Blank out settings.log if no log settings
          if (
            state.data.settings?.log &&
            Object.keys(state.data.settings.log).length === 0
          ) {
            state.data.settings.log = undefined;
          }
          // Blank out settings.web if no web settings
          if (
            state.data.settings?.web &&
            Object.keys(state.data.settings.web).length === 0
          ) {
            state.data.settings.web = undefined;
            if (state.data.settings.log === undefined) {
              state.data.settings = undefined;
            }
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
          // Gotify
          if (
            state.data.defaults?.gotify &&
            (state.data.defaults?.gotify?.extras === undefined ||
              Object.keys(state.data.defaults.gotify.extras).length === 0)
          ) {
            state.data.defaults.gotify.extras = undefined;
            // Blank out extras if there aren't any defined
            if (Object.keys(state.data.defaults.gotify).length === 1) {
              state.data.defaults.gotify = undefined;
            }
          }
          // Slack
          if (
            state.data.defaults?.slack &&
            Object.keys(state.data.defaults.slack).length === 0
          ) {
            state.data.defaults.slack = undefined;
          }
          // WebHook
          if (
            state.data.defaults?.webhook &&
            Object.keys(state.data.defaults.webhook).length === 0
          ) {
            state.data.defaults.webhook = undefined;
          }
          // Service
          if (
            state.data.defaults?.service &&
            Object.keys(state.data.defaults.service).length === 0
          ) {
            state.data.defaults.service = undefined;
            // Defaults
            if (
              Object.keys(
                state.data.defaults.gotify === undefined &&
                  state.data.defaults.slack === undefined &&
                  state.data.defaults.webhook === undefined
              )
            ) {
              state.data.defaults = undefined;
            }
          }
          break;

        default:
          console.log(action);
          throw new Error();
      }
      break;

    // INIT
    case "GOTIFY":
      switch (action.sub_type) {
        case "INIT":
          state.data.gotify = action.config_data.gotify;
          break;

        default:
          console.log(action);
          throw new Error();
      }
      break;

    // INIT
    case "SLACK":
      switch (action.sub_type) {
        case "INIT":
          state.data.slack = action.config_data.slack;
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
  return state;
}
