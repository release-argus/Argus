import { Dispatch } from "react";
import { websocketResponse } from "types/websocket";

export function handleMessage(
  action: websocketResponse,
  reducer: Dispatch<websocketResponse>
) {
  switch (action.page) {
    case "APPROVALS":
      switch (action.type) {
        // SERVICE || VERSION
        case "SERVICE":
        case "VERSION":
          reducer(action);
      }
  }
}
