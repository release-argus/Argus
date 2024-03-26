import { ReactElement, useEffect, useState } from "react";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons";
import { fetchJSON } from "utils";
import { stringify } from "yaml";
import { useDelayedRender } from "hooks/delayed-render";
import { useQuery } from "@tanstack/react-query";

export const Config = (): ReactElement => {
  const delayedRender = useDelayedRender(750);
  const [mutatedData, setMutatedData] = useState<
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    undefined | Record<string, any>
  >(undefined);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data, isFetching } = useQuery<Record<string, any>>({
    queryKey: ["config"],
    queryFn: () => fetchJSON({url:"api/v1/config"}),
    staleTime: 0,
  });

  useEffect(() => {
    if (!isFetching && data) {
      setMutatedData(updateConfig(data));
    }
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
      {mutatedData && <pre className="config">{stringify(mutatedData)}</pre>}
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
          (path.startsWith(".service") &&
            (path.endsWith("notify") || path.endsWith("webhook"))) ||
          (path.startsWith(".defaults.service") && path.endsWith("notify")) ||
          path.endsWith("webhook")
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

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const updateConfig = (config: Record<string, any>) => {
  trimConfig(config);
  config.service = orderServices(config.service, config.order);
  delete config.order;

  return config;
};
