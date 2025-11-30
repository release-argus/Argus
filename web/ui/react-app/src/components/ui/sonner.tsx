import { useTheme } from 'next-themes';
import { type CSSProperties, useEffect, useRef } from 'react';
import { Toaster as Sonner, type ToasterProps } from 'sonner';
import { addMessageHandler, useWebSocket } from '@/contexts/websocket';
import { handleNotifications } from '@/handlers/notifications';

const Toaster = ({ ...props }: ToasterProps) => {
	const { theme = 'system' } = useTheme();

	// WS Notification handler.
	const { monitorData } = useWebSocket();
	const monitorDataRef = useRef(monitorData);
	useEffect(() => {
		addMessageHandler('notifications', {
			handler: handleNotifications,
			params: {
				monitorData: monitorDataRef,
			},
		});
	}, []);

	return (
		<Sonner
			className="toaster group"
			style={
				{
					'--normal-bg': 'var(--popover)',
					'--normal-border': 'var(--border)',
					'--normal-text': 'var(--popover-foreground)',
				} as CSSProperties
			}
			theme={theme as ToasterProps['theme']}
			{...props}
		/>
	);
};

export { Toaster };
