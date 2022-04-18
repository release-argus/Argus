import { NotificationType } from "types/notification";
import { websocketResponse } from "types/websocket";

export interface notificationsParams {
  event: websocketResponse;
  addNotification: (notification: NotificationType) => void;
}

export const handleNotifications = (params: notificationsParams) => {
  switch (params.event.page) {
    case "APPROVALS":
      switch (params.event.type) {
        case "SERVICE":
          // INIT
          // ORDERING
          switch (params.event.sub_type) {
            case "INIT":
              break;
            case "ORDERING":
              break;
            case "CHANGE":
              break;
            default:
              break;
          }
          break;
        case "VERSION":
          // QUERY
          // NEW
          // UPDATED
          // INIT
          switch (params.event.sub_type) {
            case "QUERY":
              break;
            case "NEW":
              params.addNotification({
                type: "info",
                title: params.event.service_data?.id || "Unknown",
                body: `New version: ${
                  params.event.service_data?.status?.latest_version || "Unknown"
                }`,
                small:
                  params.event.service_data?.status?.latest_version_timestamp ||
                  new Date().toString(),
                delay: 0,
              });
              break;
            case "UPDATED":
              params.addNotification({
                type: "success",
                title: params.event.service_data?.id || "Unknown",
                body: `Updated to version '${
                  params.event.service_data?.status?.current_version ||
                  "Unknown"
                }'`,
                small:
                  params.event.service_data?.status
                    ?.current_version_timestamp || new Date().toString(),
                delay: 30000,
              });
              break;
            case "INIT":
              params.addNotification({
                type: "info",
                title: params.event.service_data?.id || "Unknown",
                body: `Latest version: ${
                  params.event.service_data?.status?.latest_version || "Unknown"
                }`,
                small:
                  params.event.service_data?.status?.latest_version_timestamp ||
                  "",
                delay: 5000,
              });
              break;
            case "SKIPPED":
              if (params.event.service_data?.status?.approved_version) {
                params.addNotification({
                  type: "info",
                  title: params.event.service_data?.id || "Unknown",
                  body: `Skipped version: ${params.event.service_data.status.approved_version.slice(
                    "SKIP_".length
                  )}`,
                  small: new Date().toString(),
                  delay: 30000,
                });
              }
              break;
            default:
              break;
          }
          break;
        case "RESET":
          break;

        case "WEBHOOK":
          // SUMMARY
          // EVENT
          switch (params.event.sub_type) {
            case "SUMMARY":
              break;
            case "EVENT":
              for (const key in params.event.webhook_data) {
                params.event.webhook_data[key].failed === false
                  ? params.addNotification({
                      type: "success",
                      title: params.event.service_data?.id || "Unknown",
                      body: `'${key}' WebHook sent successfully`,
                      small: new Date().toString(),
                      delay: 30000,
                    })
                  : params.addNotification({
                      type: "danger",
                      title: params.event.service_data?.id || "Unknown",
                      body: `'${key}' WebHook failed to send`,
                      small: new Date().toString(),
                      delay: 30000,
                    });
              }
              break;
            case "SENDING":
              break;
            case "RESET":
              break;
          }
          break;
        default:
          break;
      }
      break;
    default:
      break;
  }
};
