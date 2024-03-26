import { NotificationContext } from "contexts/notification";
import { useContext } from "react";

/**
 * useNotifications is a hook to use the NotificationContext
 *
 * @returns The NotificationContext
 */
const useNotifications = () => {
  return useContext(NotificationContext);
};

export default useNotifications;
