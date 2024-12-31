import { BuildInfo, RuntimeInfo } from "types/info";
import { Placeholder, Table } from "react-bootstrap";

import { Dictionary } from "types/util";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { ReactElement } from "react";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons";
import { fetchJSON } from "utils";
import { useDelayedRender } from "hooks/delayed-render";
import { useQuery } from "@tanstack/react-query";

const titleMappings: Dictionary<string> = {
  cwd: "Working directory",
};
const ignoreCapitalize = ["GOMAXPROCS", "GOGC", "GODEBUG"];

/**
 * @returns The status page, which includes tables of runtime info and of build info.
 */
export const Status = (): ReactElement => {
  const delayedRender = useDelayedRender(750);

  const { data: runtimeData } = useQuery<RuntimeInfo>({
    queryKey: ["status/runtime"],
    queryFn: () => fetchJSON({ url: `api/v1/status/runtime` }),
    staleTime: Infinity, // won't change unless Argus is restarted.
  });
  const { data: buildData } = useQuery<BuildInfo>({
    queryKey: ["version"],
    queryFn: () => fetchJSON({ url: `api/v1/version` }),
    staleTime: Infinity, // won't change unless Argus is restarted.
  });

  return (
    <>
      <h2
        style={{
          display: "inline-block",
        }}
      >
        Runtime Information
        {runtimeData === undefined &&
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
      <Table striped bordered>
        <tbody>
          {runtimeData === undefined
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
            : Object.entries(runtimeData).map(([k, v]) => {
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
        {buildData === undefined &&
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
      <Table striped bordered>
        <tbody>
          {buildData === undefined
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
            : Object.entries(buildData).map(([k, v]) => {
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
