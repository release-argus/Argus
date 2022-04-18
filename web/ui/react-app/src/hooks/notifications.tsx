import { NotificationContext } from "contexts/notification";
import { useContext } from "react";

const useNotifications = () => {
  return useContext(NotificationContext);
};

export default useNotifications;
