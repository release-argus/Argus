import { addMessageHandler, useWebSocket } from './websocket';
import {
	createContext,
	useCallback,
	useEffect,
	useMemo,
	useRef,
	useState,
} from 'react';

import Notification from 'components/notification';
import { NotificationType } from 'types/notification';
import { ToastContainer } from 'react-bootstrap';
import { handleNotifications } from 'handlers/notifications';
import { isEmptyArray } from 'utils';

interface NotificationCtx {
	notifications: NotificationType[];
	addNotification: (notification: NotificationType) => void;
	removeNotification: (id: number | undefined) => void;
}

/**
 * Provides notifications to the application and functions to add and remove them.
 *
 * @param notifications - The notifications to display
 * @param addNotification - Function to add a notification
 * @param removeNotification - Function to remove a notification
 * @returns A context to view, add, and remove notifications
 */
const NotificationContext = createContext<NotificationCtx>({
	notifications: [],
	addNotification: (_notification: NotificationType) => {},
	removeNotification: (_id: number | undefined) => {},
});

/**
 * @returns A provider of notifications to the application.
 */
const NotificationProvider = () => {
	const [notifications, setNotifications] = useState<NotificationType[]>([]);
	const { monitorData } = useWebSocket();
	const monitorDataRef = useRef(monitorData);

	// Update the ref whenever monitorData changes
	useEffect(() => {
		monitorDataRef.current = monitorData;
	}, [monitorData]);

	const addNotification = (notification: NotificationType) => {
		// Don't repeat notifications.
		setNotifications((prevState: NotificationType[]) => [
			...prevState,
			{
				...notification,
				id: isEmptyArray(prevState)
					? 0
					: (prevState[prevState.length - 1].id as number) + 1,
			},
		]);
	};

	useEffect(() => {
		addMessageHandler('notifications', handleNotifications, {
			addNotification: addNotification,
			monitorData: monitorDataRef,
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
		[notifications, addNotification, removeNotification],
	);

	return (
		<NotificationContext.Provider value={contextValue}>
			<ToastContainer
				className="p-3 position-fixed"
				position="bottom-end"
				style={{
					zIndex: 1056,
					display: 'flex',
					flexDirection: 'column',
					width: 'max-content',
					overflow: 'hidden',
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
