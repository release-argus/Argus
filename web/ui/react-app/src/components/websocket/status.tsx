import { LoaderCircle } from 'lucide-react';
import type { FC } from 'react';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { WS_ADDRESS } from '@/config';
import { useDelayedRender } from '@/hooks/use-delayed-render';

type WebSocketStatusProps = {
	/* The connection status of the WebSocket. */
	connected?: boolean;
};

/**
 * @param connected - The connection status of the WebSocket.
 * @returns A warning 'Alert' if not connected.
 */
export const WebSocketStatus: FC<WebSocketStatusProps> = ({ connected }) => {
	const delayedRender = useDelayedRender(1000);
	const fallback = (
		<Alert variant={connected === false ? 'destructive' : 'default'}>
			<AlertTitle>
				WebSocket{' '}
				{connected === false ? 'disconnected! Reconnecting' : 'connecting'}
			</AlertTitle>
			<AlertDescription className="flex flex-row items-center gap-2 pt-1">
				<LoaderCircle className="animate-spin" />
				Connecting to {WS_ADDRESS}...
			</AlertDescription>
		</Alert>
	);
	return connected !== true && delayedRender(() => fallback);
};
