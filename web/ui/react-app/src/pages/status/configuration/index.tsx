import { ReactElement, useEffect, useState } from "react";
import {
  containsEndsWith,
  containsStartsWith,
  fetchJSON,
  isEmptyObject,
} from "utils";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons";
import { stringify } from "yaml";
import { useDelayedRender } from "hooks/delayed-render";
import { useQuery } from "@tanstack/react-query";

/**
 * @returns The configuration page, which includes a preformatted YAML object of the config.yml.
 */
export const Config = (): ReactElement => {
  const delayedRender = useDelayedRender(750);
  const [mutatedData, setMutatedData] = useState<
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    undefined | Record<string, any>
  >(undefined);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data, isFetching } = useQuery<Record<string, any>>({
    queryKey: ["config"],
    queryFn: () => fetchJSON({ url: `api/v1/config` }),
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

/**
 * Recursively trims the object, removing empty objects.
 *
 * @param obj - The object to trim
 * @param path - The path of the object
 * @returns The trimmed object
 */
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
        isEmptyObject(obj[key]) &&
        !(
          // notify/webhook objects may be empty to reference mains.
          // .service.*.notify | .service.*.webhook
          // .defaults.service.*.notify | .defaults.service.*.webhook
          (
            containsEndsWith(path, [".notify", ".webhook"]) &&
            containsStartsWith(path, [".service", ".defaults.service"])
          )
        )
      )
        delete obj[key];
    }
  }
  return obj;
};

/**
 * Orders the services in the object according to the order array.
 *
 * @param object - The object to order
 * @param order - The ordering to apply
 * @returns The ordered object
 */
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

/**
 * Updates the configuration object, ordering the services and removing the order key.
 *
 * @param config - The configuration object
 * @returns The configuration object with the services ordered and the order key removed
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const updateConfig = (config: Record<string, any>) => {
  trimConfig(config);
  config.service = orderServices(config.service, config.order);
  delete config.order;

  return config;
};
