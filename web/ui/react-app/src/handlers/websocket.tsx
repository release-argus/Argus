import { Dispatch } from "react";
import { WebSocketResponse } from "types/websocket";

export function handleMessage(
  action: WebSocketResponse,
  reducer: Dispatch<WebSocketResponse>
) {
  if (
    action.page === "APPROVALS" &&
    ["SERVICE", "VERSION", "EDIT", "DELETE", "RESET"].includes(action.type)
  ) {
    reducer(action);
  }
}
