import { createContext, useCallback, useEffect, useState } from "react";

import Notification from "components/notification";
import { NotificationType } from "types/notification";
import { ToastContainer } from "react-bootstrap";
import { addMessageHandler } from "./websocket";
import { handleNotifications } from "handlers/notifications";

interface NotificationCtx {
  notifications: NotificationType[];
  addNotification: (notification: NotificationType) => void;
  removeNotification: (id: number | undefined) => void;
}

const NotificationContext = createContext<NotificationCtx>({
  notifications: [],
  // eslint-disable-next-line @typescript-eslint/no-empty-function, @typescript-eslint/no-unused-vars
  addNotification: (notification: NotificationType) => {},
  // eslint-disable-next-line @typescript-eslint/no-empty-function, @typescript-eslint/no-unused-vars
  removeNotification: (id: number | undefined) => {},
});

let id = 0;
const NotificationProvider = () => {
  const [notifications, setNotifications] = useState<NotificationType[]>([]);

  const addNotification = useCallback(
    (notification: NotificationType) => {
      // Don't repeat notifications
      setNotifications((notifications) => [
        ...notifications,
        { ...notification, id: id++ },
      ]);
    },
    [setNotifications]
  );

  useEffect(() => {
    addMessageHandler("notifications", handleNotifications, {
      addNotification: addNotification,
    });
  }, [addNotification]);

  const removeNotification = useCallback(
    (id: number | undefined) => {
      id !== undefined &&
        setNotifications((notifications) =>
          notifications.filter((n) => n.id !== id)
        );
    },
    [setNotifications]
  );

  return (
    <NotificationContext.Provider
      value={{
        notifications,
        addNotification,
        removeNotification,
      }}
    >
      <ToastContainer
        className="p-3 position-fixed"
        position={"bottom-end"}
        style={{
          zIndex: 1056,
          display: "flex",
          flexDirection: "column",
          width: "max-content",
          overflow: "hidden",
        }}
        key="notifications"
      >
        {notifications.map((notification) => (
          <Notification
            id={notification.id}
            title={notification.title}
            type={notification.type}
            body={notification.body}
            small={notification.small}
            key={notification.id}
            delay={notification.delay}
          />
        ))}
      </ToastContainer>
    </NotificationContext.Provider>
  );
};

export { NotificationContext, NotificationProvider };
