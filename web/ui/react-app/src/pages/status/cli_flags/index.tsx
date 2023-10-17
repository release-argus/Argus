import { Placeholder, Table } from "react-bootstrap";
import { ReactElement, useEffect, useState } from "react";

import { Dictionary } from "types/util";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons";
import { fetchJSON } from "utils";
import { useDelayedRender } from "hooks/delayed-render";
import { useQuery } from "@tanstack/react-query";
import { useTheme } from "contexts/theme";

export const Flags = (): ReactElement => {
  const delayedRender = useDelayedRender(750);
  const [flags, setFlags] =
    useState<Dictionary<string | boolean | undefined>>();
  const themeCtx = useTheme();

  const { data, isFetching } = useQuery<
    Dictionary<string | boolean | undefined>
  >({
    queryKey: ["flags"],
    queryFn: () => fetchJSON(`api/v1/flags`),
    staleTime: Infinity,
  });

  useEffect(() => {
    if (!isFetching && data) setFlags(data);
  }, [data]);

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
