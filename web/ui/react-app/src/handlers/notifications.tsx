import { NotificationType } from "types/notification";
import { WebSocketResponse } from "types/websocket";

export interface Props {
  event: WebSocketResponse;
  addNotification: (notification: NotificationType) => void;
}

export const handleNotifications = (props: Props) => {
  if (props.event.page !== "APPROVALS") return;
  // APPROVALS

  // VERSION
  // COMMAND
  // WEBHOOK
  switch (props.event.type) {
    case "VERSION":
      // QUERY
      // NEW
      // UPDATED
      // INIT
      // ACTION
      switch (props.event.sub_type) {
        case "QUERY":
          break;
        case "NEW":
          props.addNotification({
            type: "info",
            title: props.event.service_data?.id || "Unknown",
            body: `New version: ${
              props.event.service_data?.status?.latest_version || "Unknown"
            }`,
            small:
              props.event.service_data?.status?.latest_version_timestamp ||
              new Date().toString(),
            delay: 0,
          });
          break;
        case "UPDATED":
          props.addNotification({
            type: "success",
            title: props.event.service_data?.id || "Unknown",
            body: `Updated to version '${
              props.event.service_data?.status?.deployed_version || "Unknown"
            }'`,
            small:
              props.event.service_data?.status?.deployed_version_timestamp ||
              new Date().toString(),
            delay: 30000,
          });
          break;
        case "INIT":
          props.addNotification({
            type: "info",
            title: props.event.service_data?.id || "Unknown",
            body: `Latest version: ${
              props.event.service_data?.status?.latest_version || "Unknown"
            }`,
            small:
              props.event.service_data?.status?.latest_version_timestamp || "",
            delay: 5000,
          });
          break;
        case "ACTION":
          // Notify on SKIP, ignore latest version being approved
          if (props.event.service_data?.status?.approved_version) {
            if (
              props.event.service_data.status.approved_version.startsWith(
                "SKIP_"
              )
            )
              props.addNotification({
                type: "info",
                title: props.event.service_data?.id || "Unknown",
                body: `Skipped version: ${props.event.service_data.status.approved_version.slice(
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

    case "COMMAND":
      if (props.event.sub_type !== "EVENT") return;
      // EVENT
      for (const key in props.event.command_data) {
        props.event.command_data[key].failed === false
          ? props.addNotification({
              type: "success",
              title: props.event.service_data?.id || "Unknown",
              body: `'${key}' Command ran successfully`,
              small: new Date().toString(),
              delay: 30000,
            })
          : props.addNotification({
              type: "danger",
              title: props.event.service_data?.id || "Unknown",
              body: `'${key}' Command failed`,
              small: new Date().toString(),
              delay: 30000,
            });
      }
      break;
    case "WEBHOOK":
      if (props.event.sub_type !== "EVENT") return;
      // EVENT
      for (const key in props.event.webhook_data) {
        props.event.webhook_data[key].failed === false
          ? props.addNotification({
              type: "success",
              title: props.event.service_data?.id || "Unknown",
              body: `'${key}' WebHook sent successfully`,
              small: new Date().toString(),
              delay: 30000,
            })
          : props.addNotification({
              type: "danger",
              title: props.event.service_data?.id || "Unknown",
              body: `'${key}' WebHook failed to send`,
              small: new Date().toString(),
              delay: 30000,
            });
      }
      break;
  }
};
