import { Button, Toast } from 'react-bootstrap';
import { FC, useEffect } from 'react';
import {
	IconDefinition,
	faCheckCircle,
	faExclamationCircle,
	faExclamationTriangle,
	faInfoCircle,
	faQuestionCircle,
	faXmark,
} from '@fortawesome/free-solid-svg-icons';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { NotificationType } from 'types/notification';
import { formatRelative } from 'date-fns';
import useNotification from 'hooks/notifications';

/**
 * A notification toast component.
 *
 * @param id - Unique identifier for the notification.
 * @param title - Title text displayed in the notification.
 * @param type - Category or style of the notification (e.g., success, error).
 * @param body - Main content of the notification.
 * @param small - Whether the notification is rendered in a smaller size.
 * @param delay - Time (in milliseconds) before the notification is removed.
 * @returns A notification component with its title, type, body and creation time.
 */
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
			const timer = setTimeout(() => removeNotification(id), delay ?? 10000);

			return () => clearTimeout(timer);
		}
	}, [delay, id, removeNotification]);

	const faIcon = () => {
		const iconMap: Record<string, IconDefinition> = {
			info: faInfoCircle,
			success: faCheckCircle,
			warning: faExclamationTriangle,
			danger: faExclamationCircle,
		};

		return iconMap[type] || faQuestionCircle;
	};

	return (
		<Toast
			id={`${id}`}
			className={`m-1 alert-${type}`}
			bg={type}
			key={`notification-${id}`}
			onClose={() => removeNotification(id)}
		>
			<Toast.Header
				className={`alert-${type}`}
				style={{ padding: '0.5em' }}
				closeButton={false}
			>
				<FontAwesomeIcon
					icon={faIcon()}
					style={{ paddingRight: '0.5em', height: '1.25em' }}
				/>
				<strong className="me-auto">{title}</strong>

				<small style={{ paddingLeft: '1rem', fontSize: '0.7em' }}>
					{formatRelative(new Date(small), new Date())}
				</small>
				<Button
					key="details"
					className=""
					variant="none"
					onClick={() => removeNotification(id)}
					style={{
						display: 'flex',
						padding: '0px',
						paddingLeft: '0.5em',
						height: '1.5em',
					}}
				>
					<FontAwesomeIcon
						icon={faXmark}
						className={`alert-${type}`}
						style={{ height: '100%', width: '100%' }}
					/>
				</Button>
			</Toast.Header>
			<Toast.Body className={`notification-${type}`}>{body}</Toast.Body>
		</Toast>
	);
};

export default Notification;
