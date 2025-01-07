import { MonitorSummaryType, ServiceSummaryType } from "types/summary";

import { MutableRefObject } from "react";
import { NotificationType } from "types/notification";
import { WebSocketResponse } from "types/websocket";

export interface Props {
  event: WebSocketResponse;
  addNotification: (notification: NotificationType) => void;
  monitorData: MutableRefObject<MonitorSummaryType>;
}

/**
 * Adds a notification based on the event type and subtype
 *
 * @param props - The event and the function to add a notification.
 */
export const handleNotifications = (props: Props) => {
  if (props.event.page !== "APPROVALS") return;
  // APPROVALS

  const service_data = (props.event as { service_data?: ServiceSummaryType })
    .service_data;
  const service_name = service_data?.id
    ? props.monitorData.current?.service[service_data.id]?.name ?? service_data.id
    : "Unknown";

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
            title: service_name,
            body: `New version: ${
              props.event.service_data?.status?.latest_version ?? "Unknown"
            }`,
            small:
              props.event.service_data?.status?.latest_version_timestamp ??
              new Date().toString(),
            delay: 0,
          });
          break;
        case "UPDATED":
          props.addNotification({
            type: "success",
            title: service_name,
            body: `Updated to version '${
              props.event.service_data?.status?.deployed_version ?? "Unknown"
            }'`,
            small:
              props.event.service_data?.status?.deployed_version_timestamp ??
              new Date().toString(),
            delay: 30000,
          });
          break;
        case "INIT":
          props.addNotification({
            type: "info",
            title: service_name,
            body: `Latest version: ${
              props.event.service_data?.status?.latest_version ?? "Unknown"
            }`,
            small:
              props.event.service_data?.status?.latest_version_timestamp ?? "",
            delay: 5000,
          });
          break;
        case "ACTION":
          // Notify on SKIP, ignore latest version being approved.
          if (props.event.service_data?.status?.approved_version) {
            if (
              props.event.service_data.status.approved_version.startsWith(
                "SKIP_"
              )
            )
              props.addNotification({
                type: "info",
                title: service_name,
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
              title: service_name,
              body: `'${key}' Command ran successfully`,
              small: new Date().toString(),
              delay: 30000,
            })
          : props.addNotification({
              type: "danger",
              title: service_name,
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
              title: service_name,
              body: `'${key}' WebHook sent successfully`,
              small: new Date().toString(),
              delay: 30000,
            })
          : props.addNotification({
              type: "danger",
              title: service_name,
              body: `'${key}' WebHook failed to send`,
              small: new Date().toString(),
              delay: 30000,
            });
      }
      break;
  }
};
