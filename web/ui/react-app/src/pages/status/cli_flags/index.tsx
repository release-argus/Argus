import { Placeholder, Table } from "react-bootstrap";
import { ReactElement, useEffect, useState } from "react";
import {
  addMessageHandler,
  removeMessageHandler,
  sendMessage,
} from "contexts/websocket";

import { Dictionary } from "types/util";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { WebSocketResponse } from "types/websocket";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons";
import { useDelayedRender } from "hooks/delayed-render";
import { useTheme } from "contexts/theme";

export const Flags = (): ReactElement => {
  const delayedRender = useDelayedRender(750);
  const [flags, setFlags] = useState<Dictionary<string>>();
  const themeCtx = useTheme();

  useEffect(() => {
    sendMessage(
      JSON.stringify({
        version: 1,
        page: "FLAGS",
        type: "INIT",
      })
    );

    // Handler to listen to WebSocket messages
    const handler = (event: WebSocketResponse) => {
      if (event.page === "FLAGS" && event.flags_data) {
        setFlags(event.flags_data);
        removeMessageHandler("flags");
      }
    };
    addMessageHandler("flags", handler);
  }, []);

  return (
    <>
      <h2
        style={{
          display: "inline-block",
        }}
      >
        Command-Line Flags
        {flags === undefined &&
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
      <Table
        striped
        bordered
        variant={themeCtx.theme === "theme-dark" ? "dark" : undefined}
      >
        <thead>
          <tr>
            <th>Flag</th>
            <th>Value</th>
          </tr>
        </thead>
        <tbody>
          {flags === undefined
            ? [...Array.from(Array(9).keys())].map((num) => (
                <tr key={num}>
                  <th style={{ width: "35%" }}>
                    {delayedRender(() => (
                      <Placeholder xs={4} />
                    ))}
                    &nbsp;
                  </th>
                  <td>
                    {delayedRender(() => (
                      <Placeholder xs={3} />
                    ))}
                  </td>
                </tr>
              ))
            : Object.entries(flags).map(([k, v]) => {
                return (
                  <tr key={k}>
                    <th style={{ width: "35%" }}>{`-${k}`}</th>
                    <td>{v === null ? "" : `${v}`}</td>
                  </tr>
                );
              })}
        </tbody>
      </Table>
    </>
  );
};
