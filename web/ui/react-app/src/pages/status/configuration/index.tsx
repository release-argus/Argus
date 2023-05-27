import { ReactElement, useEffect, useState } from "react";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons";
import { fetchJSON } from "utils";
import { stringify } from "yaml";
import { useDelayedRender } from "hooks/delayed-render";
import { useQuery } from "@tanstack/react-query";

export const Config = (): ReactElement => {
  const delayedRender = useDelayedRender(750);
  const [mutated, setMutated] = useState(false);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data, isFetching } = useQuery<Record<string, any>>(
    ["config"],
    () => fetchJSON(`api/v1/config`),
    { staleTime: 0 }
  );

  useEffect(() => {
    if (!isFetching && data) {
      trimConfig(data);
      data.service = orderServices(data.service, data.order);
      delete data.order;
    }
    setMutated(!isFetching);
  }, [data]);

  return (
    <>
      <h2
        style={{
          display: "inline-block",
        }}
      >
        Configuration
        {isFetching &&
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
      {mutated && <pre className="config">{stringify(data)}</pre>}
    </>
  );
};

const trimConfig = (
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  obj: Record<string, any>,
  path = ""
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
): Record<string, any> => {
  for (const key in obj) {
    if (typeof obj[key] === "object" && obj[key] !== null) {
      obj[key] = trimConfig(obj[key], `${path}.${key}`);
      if (
        Object.keys(obj[key]).length === 0 &&
        !(
          path.startsWith(".service") &&
          (path.endsWith("notify") || path.endsWith("webhook"))
        )
      )
        delete obj[key];
    }
  }
  return obj;
};

const orderServices = <T extends Record<string, unknown>>(
  object: T,
  order?: Array<keyof T>
): T => {
  if (!order) return object;
  const orderedObject = {} as T;
  order.forEach((key) => {
    if (object.hasOwnProperty(key)) orderedObject[key] = object[key];
  });
  return orderedObject;
};
