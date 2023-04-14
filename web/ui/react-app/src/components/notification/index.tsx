import { Button, Toast } from "react-bootstrap";
import { FC, useEffect } from "react";
import {
  faCheckCircle,
  faExclamationCircle,
  faExclamationTriangle,
  faInfoCircle,
  faQuestionCircle,
  faXmark,
} from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { NotificationType } from "types/notification";
import { formatRelative } from "date-fns";
import useNotification from "hooks/notifications";

const Notification: FC<NotificationType> = ({
  id,
  title,
  type,
  body,
  small,
  delay,
}) => {
  const { removeNotification } = useNotification();

  useEffect(() => {
    if (delay !== 0) {
      const timer = setTimeout(
        () => {
          removeNotification(id);
        },
        delay ? delay : 10000
      );

      return () => {
        clearTimeout(timer);
      };
    }
  }, [delay, id, removeNotification]);

  return (
    <Toast
      id={`${id}`}
      className={`m-1 text-white alert-${type} `}
      bg={type}
      key={`notification-${id}`}
      onClose={() => removeNotification(id)}
    >
      <Toast.Header
        className={`alert-${type} text-white`}
        style={{ padding: "0.5em" }}
        closeVariant="white"
        closeButton={false}
      >
        <FontAwesomeIcon
          icon={
            type === "info"
              ? faInfoCircle
              : type === "success"
              ? faCheckCircle
              : type === "warning"
              ? faExclamationTriangle
              : type === "danger"
              ? faExclamationCircle
              : faQuestionCircle
          }
          style={{ paddingRight: "0.5em", height: "1.25em" }}
        />
        <strong className="me-auto">{title}</strong>

        <small style={{ paddingLeft: "1rem", fontSize: "0.7em" }}>
          <>{formatRelative(new Date(small), new Date())}</>
        </small>
        <Button
          key="details"
          className=""
          variant="none"
          onClick={() => removeNotification(id)}
          style={{
            display: "flex",
            padding: "0px",
            paddingLeft: "0.5em",
            height: "1.5em",
          }}
        >
          <FontAwesomeIcon
            icon={faXmark}
            className={`alert-${type}`}
            style={{ height: "100%", width: "100%" }}
          />
        </Button>
      </Toast.Header>
      <Toast.Body className={`notification-${type}`}>{body}</Toast.Body>
    </Toast>
  );
};

export default Notification;
