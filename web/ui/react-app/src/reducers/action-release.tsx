import { ActionModalData } from "types/summary";
import { WebSocketResponse } from "types/websocket";

/**
 * Returns a reducer that handles actions on the action modal.
 *
 * @param state - The current state of the action modal
 * @param action - The action to perform on the action modal
 * @returns The new state of the action modal
 */
export default function reducerActionModal(
  state: ActionModalData,
  action: WebSocketResponse
): ActionModalData {
  // eslint-disable-next-line prefer-const
  let newState: ActionModalData = JSON.parse(JSON.stringify(state));

  switch (action.type) {
    // EVENT
    // INIT
    case "COMMAND":
    case "WEBHOOK":
      if (
        !action.service_data ||
        (!action.webhook_data && !action.command_data)
      )
        return state;

      if (action.sub_type == "EVENT") {
        if (action.webhook_data)
          for (const webhook_id in action.webhook_data) {
            // Remove them from the sending list.
            newState.sentWH.splice(
              newState.sentWH.indexOf(
                `${action.service_data.id} ${webhook_id}`
              ),
              1
            );

            // Record the success/fail (if it's the current modal service).
            if (
              action.service_data.id === state.service_id &&
              newState.webhooks[webhook_id] !== undefined
            )
              newState.webhooks[webhook_id] = {
                failed: action.webhook_data[webhook_id].failed,
                next_runnable: action.webhook_data[webhook_id].next_runnable,
              };
          }

        if (action.command_data)
          for (const command in action.command_data) {
            // Remove them from the sending list.
            newState.sentC.splice(
              newState.sentC.indexOf(`${action.service_data.id} ${command}`),
              1
            );

            // Record the success/fail (if it's the current modal service).
            if (
              action.service_data.id === state.service_id &&
              newState.commands[command] !== undefined
            )
              newState.commands[command] = {
                failed: action.command_data[command].failed,
                next_runnable: action.command_data[command].next_runnable,
              };
          }
        break;
      } else {
        console.error(action);
        throw new Error();
      }

    // REFRESH
    // RESET
    // SENDING
    case "ACTION":
      switch (action.sub_type) {
        case "REFRESH":
          newState.service_id = action.service_data?.id ?? newState.service_id;
          newState.commands = action.command_data ?? {};
          newState.webhooks = action.webhook_data ?? {};
          break;
        case "RESET":
          newState.commands = {};
          newState.webhooks = {};
          break;
        case "SENDING":
          // Command(s).
          if (action.command_data)
            for (const command in action.command_data) {
              // reset the failed states.
              if (newState.commands[command])
                newState.commands[command].failed = undefined;
              // set as sending.
              newState.sentC.push(`${action.service_data?.id} ${command}`);
            }

          // WebHook(s).
          if (action.webhook_data)
            for (const webhook_id in action.webhook_data) {
              // reset the failed states.
              if (newState.webhooks[webhook_id])
                newState.webhooks[webhook_id].failed = undefined;
              // set as sending.
              newState.sentWH.push(`${action.service_data?.id} ${webhook_id}`);
            }
          break;
        default:
          console.error(action);
          throw new Error();
      }
      break;

    default:
      console.error(action);
      throw new Error();
  }

  // Got to update the state more for the reload.
  state = newState;
  return state;
}
