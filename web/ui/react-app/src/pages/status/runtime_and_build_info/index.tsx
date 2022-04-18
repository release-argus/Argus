import { Placeholder, Table } from "react-bootstrap";
import { ReactElement, useEffect, useState } from "react";
import {
  addMessageHandler,
  removeMessageHandler,
  sendMessage,
} from "contexts/websocket";

import { Dictionary } from "types/util";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { Info } from "types/info";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons";
import { useDelayedRender } from "hooks/delayed-render";
import { useTheme } from "contexts/theme";
import { websocketResponse } from "types/websocket";

const titleMappings: Dictionary<string> = {
  cwd: "Working directory",
};
const ignoreCapitalize = ["GOMAXPROCS", "GOGC", "GODEBUG"];

export const Status = (): ReactElement => {
  const delayedRender = useDelayedRender(750);
  const [info, setInfo] = useState<Info>();
  const themeCtx = useTheme();

  useEffect(() => {
    sendMessage(
      JSON.stringify({
        version: 1,
        page: "RUNTIME_BUILD",
        type: "INIT",
      })
    );

    // Handler to listen to WebSocket messages
    const handler = (event: websocketResponse) => {
      if (event.page === "RUNTIME_BUILD" && event.info_data) {
        setInfo({
          build: event.info_data.build,
          runtime: event.info_data.runtime,
        });
        removeMessageHandler("status");
      }
    };
    addMessageHandler("status", handler);
  }, []);

  return (
    <>
      <h2
        style={{
          display: "inline-block",
        }}
      >
        Runtime Information
        {info?.runtime === undefined &&
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
        <tbody>
          {info?.runtime === undefined
            ? [...Array.from(Array(4).keys())].map((num) => (
                <tr key={num}>
                  <th style={{ width: "35%" }}>
                    {delayedRender(() => (
                      <Placeholder xs={4} />
                    ))}
                    &nbsp;
                  </th>
                  <td>
                    {delayedRender(() => (
                      <Placeholder xs={2} />
                    ))}
                  </td>
                </tr>
              ))
            : Object.entries(info.runtime).map(([k, v]) => {
                const title = (
                  k in titleMappings ? titleMappings[k] : k
                ).replaceAll("_", " ");
                const capitalize = ignoreCapitalize.includes(k)
                  ? ""
                  : "capitalize-title";

                return (
                  <tr key={k}>
                    <th className={capitalize} style={{ width: "35%" }}>
                      {title}
                    </th>
                    <td>{v}</td>
                  </tr>
                );
              })}
        </tbody>
      </Table>
      <h2
        style={{
          display: "inline-block",
        }}
      >
        Build Information
        {info?.build === undefined &&
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
        <tbody>
          {info?.build === undefined
            ? [...Array.from(Array(3).keys())].map((num) => (
                <tr key={num}>
                  <th style={{ width: "35%" }}>
                    {delayedRender(() => (
                      <Placeholder xs={2} />
                    ))}
                    &nbsp;
                  </th>
                  <td>
                    {delayedRender(() => (
                      <Placeholder xs={2} />
                    ))}
                  </td>
                </tr>
              ))
            : Object.entries(info.build).map(([k, v]) => {
                const title = (
                  k in titleMappings ? titleMappings[k] : k
                ).replaceAll("_", " ");
                const capitalize = ignoreCapitalize.includes(k)
                  ? ""
                  : "capitalize-title";

                return (
                  <tr key={k}>
                    <th className={capitalize} style={{ width: "35%" }}>
                      {title}
                    </th>
                    <td>{v}</td>
                  </tr>
                );
              })}
        </tbody>
      </Table>
    </>
  );
};
