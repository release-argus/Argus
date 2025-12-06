import { useQueryClient } from '@tanstack/react-query';
import { useTheme } from 'next-themes';
import { type CSSProperties, useEffect } from 'react';
import { Toaster as Sonner, type ToasterProps } from 'sonner';
import { addMessageHandler, removeMessageHandler } from '@/contexts/websocket';
import { handleNotifications } from '@/handlers/notifications';

const Toaster = ({ ...props }: ToasterProps) => {
	const { theme = 'system' } = useTheme();

	// WS Notification handler.
	const queryClient = useQueryClient();
	useEffect(() => {
		addMessageHandler('notifications', {
			handler: handleNotifications,
			params: {
				queryClient: queryClient,
			},
		});

		return () => removeMessageHandler('notifications');
	}, [queryClient]);

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
