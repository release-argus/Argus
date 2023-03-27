import {
  NotifyOpsGenieDetailType,
  ServiceEditAPIType,
  ServiceEditType,
} from "types/service-edit";

import { NotifyOpsGenieTarget } from "types/config";

export const convertAPIServiceDataEditToUI = (
  name: string,
  serviceData?: ServiceEditAPIType
): ServiceEditType => {
  if (serviceData && name) {
    // Edit service defaults
    return {
      ...serviceData,
      options: {
        ...serviceData?.options,
        active: serviceData?.options?.active !== false,
      },
      latest_version: {
        ...serviceData?.latest_version,
        require: {
          ...serviceData?.latest_version?.require,
          command: serviceData?.latest_version?.require?.command?.map(
            (arg) => ({
              arg: arg as string,
            })
          ),
          docker: {
            type: serviceData?.latest_version?.require?.docker?.type || "hub",
            image: serviceData?.latest_version?.require?.docker?.image,
            tag: serviceData?.latest_version?.require?.docker?.tag,
            username: serviceData?.latest_version?.require?.docker?.username,
            token: serviceData?.latest_version?.require?.docker?.token,
          },
        },
      },
      name: name,
      deployed_version: {
        url: serviceData?.deployed_version?.url,
        allow_invalid_certs: serviceData?.deployed_version?.allow_invalid_certs,
        basic_auth: {
          username: serviceData?.deployed_version?.basic_auth?.username || "",
          password: serviceData?.deployed_version?.basic_auth?.password || "",
        },
        headers:
          serviceData?.deployed_version?.headers?.map((header, key) => ({
            ...header,
            oldIndex: key,
          })) || [],
        json: serviceData?.deployed_version?.json,
        regex: serviceData?.deployed_version?.regex,
      },
      command: serviceData?.command?.map((args) => ({
        args: args.map((arg) => ({ arg })),
      })),
      webhook: serviceData?.webhook?.map((item) => ({
        ...item,
        custom_headers: item.custom_headers?.map((header, index) => ({
          ...header,
          oldIndex: index,
        })),
        oldIndex: item.name,
      })),
      notify: serviceData?.notify?.map((item) => ({
        ...item,
        oldIndex: item.name,
        params: {
          avatar: "", // controlled param
          color: "", // ^
          icon: "", // ^
          ...(item.params &&
            Object.entries(item.params).reduce(
              (acc, [key, val]) =>
                Object.assign(acc, {
                  [key]:
                    item.type !== "opsgenie"
                      ? val
                      : ["responders", "visibleto"].includes(key)
                      ? (JSON.parse(val).map(
                          (obj: {
                            id: string;
                            type: string;
                            name: string;
                            username: string;
                          }) => {
                            if (obj.id) {
                              return {
                                type: obj.type,
                                sub_type: "id",
                                value: obj.id,
                              };
                            } else {
                              return {
                                type: obj.type,
                                sub_type:
                                  obj.type === "user" ? "username" : "name",
                                value: obj.name || obj.username,
                              };
                            }
                          }
                        ) as NotifyOpsGenieTarget[])
                      : ["details"].includes(key)
                      ? (Object.entries(JSON.parse(val)).map(
                          ([key, value]) => ({
                            key: key,
                            value: value,
                          })
                        ) as NotifyOpsGenieDetailType[])
                      : val,
                }),
              {}
            )),
        },
      })),
      dashboard: {
        auto_approve: undefined,
        icon: "",
        ...serviceData?.dashboard,
      },
    };
  }

  // New service defaults
  return {
    name: "",
    options: { active: true },
    latest_version: {
      type: "github",
      require: { docker: { type: "hub" } },
    },
    dashboard: {
      auto_approve: undefined,
      icon: "",
      icon_link_to: "",
      web_url: "",
    },
  };
};
