import {
  createContext,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";

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

/**
 * The notification context, which provides notifications to the application.
 *
 * @param notifications - The notifications to display
 * @param addNotification - Function to add a notification
 * @param removeNotification - Function to remove a notification
 * @returns The notification context
 */
const NotificationContext = createContext<NotificationCtx>({
  notifications: [],
  // eslint-disable-next-line @typescript-eslint/no-empty-function, @typescript-eslint/no-unused-vars
  addNotification: (notification: NotificationType) => {},
  // eslint-disable-next-line @typescript-eslint/no-empty-function, @typescript-eslint/no-unused-vars
  removeNotification: (id: number | undefined) => {},
});

/**
 * @returns The notification provider, which provides notifications to the application.
 */
const NotificationProvider = () => {
  const [notifications, setNotifications] = useState<NotificationType[]>([]);

  const addNotification = (notification: NotificationType) => {
    // Don't repeat notifications
    setNotifications((prevState: NotificationType[]) => [
      ...prevState,
      {
        ...notification,
        id:
          prevState.length === 0
            ? 0
            : (prevState[prevState.length - 1].id as number) + 1,
      },
    ]);
  };

  useEffect(() => {
    addMessageHandler("notifications", handleNotifications, {
      addNotification: addNotification,
    });
  }, []);

  const removeNotification = useCallback((id?: number) => {
    id !== undefined &&
      setNotifications((prevState) => prevState.filter((n) => n.id !== id));
  }, []);

  const contextValue = useMemo(
    () => ({
      notifications,
      addNotification,
      removeNotification,
    }),
    [notifications, addNotification, removeNotification]
  );

  return (
    <NotificationContext.Provider value={contextValue}>
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
