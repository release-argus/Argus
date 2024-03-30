import { NotificationContext } from "contexts/notification";
import { useContext } from "react";

/**
 * @returns The notifications context
 */
const useNotifications = () => {
  return useContext(NotificationContext);
};

export default useNotifications;
