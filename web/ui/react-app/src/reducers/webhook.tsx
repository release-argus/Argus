import { WebHookModalData, WebHookSummaryListType } from "types/summary";

import { websocketResponse } from "types/websocket";

export default function reducerWebHookModal(
  state: WebHookModalData,
  action: websocketResponse
): WebHookModalData {
  switch (action.type) {
    // SUMMARY
    // EVENT
    // SENDING
    // RESENDING
    // RESET
    // INIT
    case "WEBHOOK":
      let newState: WebHookModalData = {
        service_id: state.service_id,
        sent: state.sent,
        webhooks: {},
      };
      switch (action.sub_type) {
        case "SUMMARY":
          if (!action.webhook_data || !action.service_data) {
            return state;
          }
          newState.webhooks = action.webhook_data;
          newState.service_id = action.service_data.id;
          break;
        case "EVENT":
          newState = JSON.parse(JSON.stringify(state));

          for (const webhook_id in action.webhook_data) {
            if (
              newState.webhooks[webhook_id] !== undefined &&
              action.service_data?.id === state.service_id
            ) {
              newState.webhooks[webhook_id].failed =
                action.webhook_data[webhook_id].failed;
            }
            newState.sent.splice(
              newState.sent.indexOf(`${state.service_id} ${webhook_id}`),
              1
            );
          }
          break;
        case "SENDING":
          if (!action.webhook_data) {
            return state;
          }
          newState = JSON.parse(JSON.stringify(state));
          let sending: WebHookSummaryListType = action.webhook_data;
          // Send all button
          if (Object.keys(sending).length === 0) {
            sending = state.webhooks;
            // Don't re-send successfullu sent WebHooks
            Object.keys(sending).filter(
              (webhook_id: string) => sending[webhook_id].failed !== true
            );
          }

          for (const webhook_id in sending) {
            if (newState.webhooks[webhook_id] !== undefined) {
              newState.webhooks[webhook_id].failed = undefined;
            }
            newState.sent.push(`${action.service_data?.id} ${webhook_id}`);
          }
          break;
        case "RESENDING":
          newState = JSON.parse(JSON.stringify(state));
          for (const webhook_id in state.webhooks) {
            if (newState.webhooks[webhook_id] !== undefined) {
              newState.webhooks[webhook_id].failed = undefined;
            }

            newState.sent.push(`${action.service_data?.id} ${webhook_id}`);
          }
          break;
        case "RESET":
          break;
        case "INIT":
          break;
        default:
          console.log(action);
          throw new Error();
      }

      // Gotta update the state more for the reload
      state = newState;
      return state;
    default:
      console.log(action);
      throw new Error();
  }
}
