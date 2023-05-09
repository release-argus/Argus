import { ActionModalData } from "types/summary";
import { WebSocketResponse } from "types/websocket";
import { isAfterDate } from "utils";

export default function reducerActionModal(
  state: ActionModalData,
  action: WebSocketResponse
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

          if (action.webhook_data !== undefined) {
            newState.webhooks = action.webhook_data;
          } else if (action.command_data !== undefined) {
            newState.commands = action.command_data;
          }
          newState.service_id = action.service_data.id;
          break;
        case "EVENT":
          if (action.webhook_data) {
            for (const webhook_id in action.webhook_data) {
              // Remove them from the sending list
              newState.sentWH.splice(
                newState.sentWH.indexOf(
                  `${action.service_data.id} ${webhook_id}`
                ),
                1
              );

              // Record the success/fail
              if (newState.service_id === action.service_data.id) {
                newState.webhooks[webhook_id] = {
                  failed: action.webhook_data[webhook_id].failed,
                  next_runnable: action.webhook_data[webhook_id].next_runnable,
                };
              }
            }
          } else {
            for (const command in action.command_data) {
              // Remove them from the sending list
              newState.sentC.splice(
                newState.sentC.indexOf(`${action.service_data.id} ${command}`),
                1
              );

              // Record the success/fail
              if (newState.service_id === action.service_data.id) {
                newState.commands[command] = {
                  failed: action.command_data[command].failed,
                  next_runnable: action.command_data[command].next_runnable,
                };
              }
            }
          }
          break;
        case "SENDING":
          if (action.webhook_data) {
            for (const webhook_id in action.webhook_data) {
              // reset the failed state
              if (newState.webhooks[webhook_id] !== undefined) {
                newState.webhooks[webhook_id].failed = undefined;
              }
              // set it as sending
              newState.sentWH.push(`${action.service_data?.id} ${webhook_id}`);
            }
          } else {
            for (const command in action.command_data) {
              // reset the failed state
              if (newState.commands[command] !== undefined) {
                newState.commands[command].failed = undefined;
              }
              // set it as sending
              newState.sentC.push(`${action.service_data.id} ${command}`);
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
          // Send all button
          // WebHooks
          for (const webhook_id in state.webhooks) {
            // skip webhooks that aren't after next_runnable
            if (
              state.webhooks[webhook_id].next_runnable !== undefined &&
              // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
              isAfterDate(state.webhooks[webhook_id].next_runnable!)
            ) {
              continue;
            }
            // reset the failed states
            if (newState.webhooks[webhook_id] !== undefined) {
              newState.webhooks[webhook_id].failed = undefined;
            }
            // set as sending
            newState.sentWH.push(`${action.service_data?.id} ${webhook_id}`);
          }

          // Commands
          for (const command in state.commands) {
            // skip commands that aren't after next_runnable
            if (
              state.commands[command].next_runnable !== undefined &&
              // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
              isAfterDate(state.commands[command].next_runnable!)
            ) {
              continue;
            }
            // reset the failed states
            if (newState.commands[command] !== undefined) {
              newState.commands[command].failed = undefined;
            }
            // set as sending
            newState.sentC.push(`${action.service_data?.id} ${command}`);
          }
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
