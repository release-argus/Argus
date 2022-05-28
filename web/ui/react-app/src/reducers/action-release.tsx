import {
  ActionModalData,
  CommandSummaryListType,
  WebHookSummaryListType,
} from "types/summary";

import { websocketResponse } from "types/websocket";

export default function reducerActionModal(
  state: ActionModalData,
  action: websocketResponse
): ActionModalData {
  // eslint-disable-next-line prefer-const
  let newState: ActionModalData = JSON.parse(JSON.stringify(state));

  switch (action.type) {
    // SUMMARY
    // EVENT
    // SENDING
    // INIT
    case "COMMAND":
    case "WEBHOOK":
      if (
        !action.service_data ||
        (!action.webhook_data && !action.command_data)
      ) {
        return state;
      }

      switch (action.sub_type) {
        case "SUMMARY":
          newState.commands = state.commands;

          action.webhook_data
            ? (newState.webhooks = action.webhook_data)
            : // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
              (newState.commands = action.command_data!);
          // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
          newState.service_id = action.service_data!.id;
          break;
        case "EVENT":
          if (action.webhook_data) {
            for (const webhook_id in action.webhook_data) {
              // Remove them from the sending list
              newState.sentWH.splice(
                newState.sentWH.indexOf(
                  // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                  `${action.service_data!.id} ${webhook_id}`
                ),
                1
              );

              // Record the success/fail
              // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
              if (newState.service_id === action.service_data!.id) {
                newState.webhooks[webhook_id] = {
                  failed: action.webhook_data[webhook_id].failed,
                };
              }
            }
          } else {
            for (const command in action?.command_data) {
              // Remove them from the sending list
              newState.sentC.splice(
                // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                newState.sentC.indexOf(`${action.service_data!.id} ${command}`),
                1
              );

              // Record the success/fail
              // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
              if (newState.service_id === action.service_data!.id) {
                newState.commands[command] = {
                  failed: action.command_data[command].failed,
                };
              }
            }
          }
          break;
        case "SENDING":
          let sending = action.webhook_data
            ? (action.webhook_data as WebHookSummaryListType)
            : // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
              (action.command_data! as CommandSummaryListType);

          // Send all button
          if (Object.keys(sending).length === 0) {
            sending = action.webhook_data ? state.webhooks : state.commands;
            // Don't re-send successfullu sent WebHooks
            Object.keys(sending).filter(
              (id: string) => sending[id].failed !== true
            );
          }

          if (action.webhook_data) {
            for (const webhook_id in sending) {
              // reset the failed state
              if (newState.webhooks[webhook_id] !== undefined) {
                newState.webhooks[webhook_id].failed = undefined;
              }
              // set it as sending
              newState.sentWH.push(`${action.service_data?.id} ${webhook_id}`);
            }
          } else {
            for (const command in sending) {
              // reset the failed state
              if (newState.commands[command] !== undefined) {
                newState.commands[command].failed = undefined;
              }
              // set it as sending
              newState.sentC.push(`${action.service_data?.id} ${command}`);
            }
          }
          break;
        case "INIT":
          break;
        default:
          console.log(action);
          throw new Error();
      }
      break;

    // RESET
    // SENDING
    case "ACTION":
      switch (action.sub_type) {
        case "RESET":
          newState.commands = {};
          newState.webhooks = {};
          break;
        case "SENDING":
          for (const webhook_id in state.webhooks) {
            if (newState.webhooks[webhook_id] !== undefined) {
              newState.webhooks[webhook_id].failed = undefined;
            }

            newState.sentWH.push(`${action.service_data?.id} ${webhook_id}`);
          }
          for (const command in state.commands) {
            if (newState.commands[command] !== undefined) {
              newState.commands[command].failed = undefined;
            }

            newState.sentC.push(`${action.service_data?.id} ${command}`);
          }
          console.log(state, newState);
          break;
        default:
          console.log(action);
          throw new Error();
      }
      break;

    default:
      console.log(action);
      throw new Error();
  }

  // Gotta update the state more for the reload
  state = newState;
  return state;
}
