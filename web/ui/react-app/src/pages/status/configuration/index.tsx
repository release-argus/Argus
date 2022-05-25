import { ReactElement, useEffect, useReducer } from "react";
import { addMessageHandler, sendMessage } from "contexts/websocket";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import YAML from "yaml";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons";
import reducerConfig from "reducers/config";
import { useDelayedRender } from "hooks/delayed-render";
import { websocketResponse } from "types/websocket";

export const Config = (): ReactElement => {
  const delayedRender = useDelayedRender(750);
  const [config, setConfig] = useReducer(reducerConfig, {
    data: {},
    waiting_on: ["SETTINGS", "DEFAULTS", "NOTIFY", "WEBHOOK", "SERVICE"],
  });

  useEffect(() => {
    sendMessage(
      JSON.stringify({
        version: 1,
        page: "CONFIG",
        type: "INIT",
      })
    );

    // Handler to listen to WebSocket messages
    const handler = (event: websocketResponse) => {
      if (event.page === "CONFIG") {
        setConfig(event);
      }
    };
    addMessageHandler("config", handler);
  }, []);

  return (
    <>
      <h2
        style={{
          display: "inline-block",
        }}
      >
        Configuration
        {config.waiting_on.length !== 0 &&
          delayedRender(() => (
            <div
              style={{
                display: "inline-block",
                justifyContent: "center",
                alignItems: "center",
                height: "2rem",
                paddingLeft: "1rem",
              }}
            >
              <FontAwesomeIcon
                icon={faCircleNotch}
                className="fa-spin"
                style={{
                  height: "100%",
                }}
              />
            </div>
          ))}
      </h2>

      <pre className="config">{YAML.stringify(config.data)}</pre>
    </>
  );
};
