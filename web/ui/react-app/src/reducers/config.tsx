import { ConfigState, NotifyType, ServiceDict } from "types/config";

import { cleanEmpty } from "utils/clean_empty";
import { websocketResponse } from "types/websocket";

const pruneNotify = (notify: NotifyType) => {
  if (
    notify?.options !== undefined &&
    Object.keys(
      // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
      notify!.options!
    ).length === 0
  ) {
    notify.options = undefined;
  }
  if (
    notify?.url_fields !== undefined &&
    Object.keys(
      // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
      notify!.url_fields!
    ).length === 0
  ) {
    notify.url_fields = undefined;
  }
  if (
    notify?.params !== undefined &&
    Object.keys(
      // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
      notify!.params!
    ).length === 0
  ) {
    notify.params = undefined;
  }
};

const pruneNotifies = (notifies: ServiceDict<NotifyType>) => {
  for (const notify_id in notifies) {
    pruneNotify(notifies[notify_id]);
  }
};

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
          // Notify
          if (state.data.defaults?.notify) {
            pruneNotifies(state.data.defaults.notify);
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
                state.data.defaults.notify === undefined &&
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
    case "NOTIFY":
      switch (action.sub_type) {
        case "INIT":
          if (action.config_data.notify !== undefined) {
            // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
            pruneNotifies(state.data.notify!);
          }
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
              if (action.config_data.service[service_id].notify !== undefined) {
                // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                pruneNotifies(action.config_data.service[service_id].notify!);
              }
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
