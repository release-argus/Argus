import { createContext, useContext, useReducer, useState } from "react";

import { BooleanType } from "types/boolean";
import { MonitorSummaryType } from "types/summary";
import ReconnectingWebSocket from "reconnecting-websocket";
import { WS_ADDRESS } from "config";
import { WebSocketStatus } from "components/websocket/status";
import { getBasename } from "utils/get_basename";
import { handleMessage } from "handlers/websocket";
import reducerMonitor from "reducers/monitor";

type Bool = boolean | undefined;
type Socket = ReconnectingWebSocket | undefined;
interface WebSocketCtx {
  ws: Socket;
  connected: BooleanType;
  monitorData: MonitorSummaryType;
}

export const WebSocketContext = createContext<WebSocketCtx>({
  ws: undefined,
  connected: false,
  monitorData: { order: [], service: {} },
});

interface Props {
  children: JSX.Element[];
}

const ws = new ReconnectingWebSocket(`${WS_ADDRESS}${getBasename()}/ws`);
export const WebSocketProvider = (props: Props) => {
  const [monitorData, setMonitorData] = useReducer(reducerMonitor, {
    order: ["monitorData_loading"],
    service: {},
  });
  const [connected, setConnected] = useState<Bool>(undefined);

  ws.onopen = () => {
    setConnected(true);
    // INIT for approvals can be requested here (called on every starting page load),
    // but the message fails to parse correctly
    //
    // sendMessage(
    //   JSON.stringify({
    //     version: 1,
    //     page: "APPROVALS",
    //     type: "INIT",
    //   })
    // );
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ws.onmessage = (event: any) => {
    if (event.data === "") {
      return;
    }
    // If message is valid JSON
    if (event.data.length > 1 && event.data[0] == "{") {
      handleMessage(JSON.parse(event.data.trim()), setMonitorData);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      messageHandlers.forEach((item: { handler: any; params?: any }) =>
        item.params
          ? item.handler({
              event: JSON.parse(event.data.trim()),
              ...item.params,
            })
          : item.handler(JSON.parse(event.data.trim()))
      );
      // If message isn't valid JSON, request an update of all services
    } else {
      sendMessage(
        JSON.stringify({
          version: 1,
          page: "APPROVALS",
          type: "INIT",
        })
      );
    }
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ws.onerror = (event: any) => {
    connected && setConnected(false);
    console.log("ws.onerror");
    console.log(event);
  };

  return (
    <WebSocketContext.Provider
      value={{
        ws: ws,
        connected: connected,
        monitorData: monitorData,
      }}
    >
      <WebSocketStatus connected={connected} />
      {props.children}
    </WebSocketContext.Provider>
  );
};

export const sendMessage = (data: string) => {
  ws.send(data);
};

const messageHandlers = new Map();

export const addMessageHandler = (
  id: string,
  // handler: any,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  handler: any,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  params?: any
): void => {
  messageHandlers.set(id, { handler: handler, params: params });
};

export const removeMessageHandler = (id: string) => {
  messageHandlers.delete(id);
};

export const useWebSocket = () => {
  return useContext(WebSocketContext);
};
