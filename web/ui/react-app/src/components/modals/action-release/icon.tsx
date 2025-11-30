import { Check, LoaderCircle, RotateCw, Send, X } from 'lucide-react';
import type { FC } from 'react';
import { cn } from '@/lib/utils';
import type { ModalType } from '@/utils/api/types/config/summary';

type StatusIconProps = {
	/* A boolean indicating whether the status represents a failure. */
	failed?: boolean;
};

/**
 * A status icon indicating success or failure.
 *
 * @param failed - A boolean indicating whether the status represents a failure.
 * If `true`, a "Failed" icon is displayed; if `false`, a "Successful" icon is shown.
 * If not provided, the component renders nothing.
 */
export const StatusIcon: FC<StatusIconProps> = ({ failed }) => {
	if (failed === undefined) return null;
	const Icon = failed ? X : Check;

	return (
		<Icon
			aria-label={failed ? 'Failed' : 'Successful'}
			className={cn('size-6', failed ? 'text-destructive' : 'text-success')}
		/>
	);
};

type SendingIconProps = {
	/* A boolean indicating whether the item is in a sending state. */
	sending: boolean;
	/* A boolean indicating whether the item has failed. */
	failed?: boolean;
	/* The 'type' of modal rendered. */
	modalType: ModalType;
};

/**
 * An icon that signifies a sending action or state.
 *
 * @param sending - A boolean indicating whether the item is in a sending state.
 * @param failed - A boolean indicating whether the item has failed.
 * @param modalType - The 'type' of modal rendered.
 */
export const SendingIcon: FC<SendingIconProps> = ({
	sending,
	failed,
	modalType,
}) => {
	const Icon = (() => {
		// Sending and hasn't failed.
		if (sending && failed === undefined) return LoaderCircle;
		// First time sending.
		if (modalType === 'SEND' && failed === undefined) return Send;
		// Resend.
		return RotateCw;
	})();

	return <Icon className={cn('size-5', sending && 'animate-spin')} />;
};
