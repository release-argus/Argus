import { differenceInMilliseconds, formatRelative } from 'date-fns';
import { Hourglass } from 'lucide-react';
import {
	type Dispatch,
	type FC,
	type SetStateAction,
	useEffect,
	useState,
} from 'react';
import {
	SendingIcon,
	StatusIcon,
} from '@/components/modals/action-release/icon';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import Tip from '@/components/ui/tip';
import { cn } from '@/lib/utils';
import type { ModalType } from '@/utils/api/types/config/summary';

type SendableTimeoutProps = {
	/* Item can send. */
	sendable: boolean;
	/* Item actively sending. */
	sending: boolean;
	/* Updates the item's ability to send. */
	setSendable: Dispatch<SetStateAction<boolean>>;
	/* Represents the current time. */
	now: Date;
	/* Specifies the earliest time the item may next send. */
	nextRunnable: Date;
};

/* Fallback delay in milliseconds if the `nextRunnable` time is in the past. */
const FALLBACK_SENDABLE_DELAY_MS = 1000;
/**
 * A timeout that permits the item to send after the current time exceeds its nextRunnable value,
 * or prevents sending while the item remains active.
 *
 * @param sendable - Marks the item as ready to send.
 * @param sending - Marks the item as actively sending.
 * @param setSendable - Updates the item's ability to send.
 * @param now - Represents the current time.
 * @param nextRunnable - Specifies the earliest time the item may send.
 */
const sendableTimeout = ({
	sendable,
	sending,
	setSendable,
	now,
	nextRunnable,
}: SendableTimeoutProps): (() => void) | undefined => {
	if (sending) {
		setSendable(false);
		return;
	}

	if (sendable) return;

	// If not sendable and not sending, set a timer to become sendable.
	const timeUntilNextRunnableMs = differenceInMilliseconds(nextRunnable, now);

	const timeoutDuration =
		timeUntilNextRunnableMs > 0
			? timeUntilNextRunnableMs
			: FALLBACK_SENDABLE_DELAY_MS;

	const timer = setTimeout(() => {
		setSendable(true);
	}, timeoutDuration);

	return () => {
		clearTimeout(timer);
	};
};

type ItemProps = {
	itemType: 'COMMAND' | 'WEBHOOK';
	modalType: ModalType;
	title: string;
	loading?: boolean;
	failed?: boolean;
	sending: boolean;
	next_runnable: string;
	ack: (target: string, isWebHook: boolean) => void;
};

/**
 * Renders the item's information with buttons based on the modal type.
 *
 * @param itemType - The item type (e.g. COMMAND/WEBHOOK).
 * @param modalType - The modal type.
 * @param title - The title of the item.
 * @param loading - Display a loading skeleton rather than the title.
 * @param failed - Whether the item failed.
 * @param sending - Marks the item as actively sending.
 * @param next_runnable - Defines when the item becomes eligible for sending.
 * @param ack - The action to send this item.
 * @returns Displays the item's information with buttons based on the modal type.
 */
export const Item: FC<ItemProps> = ({
	itemType,
	modalType,
	title,
	loading,
	failed,
	sending,
	next_runnable,
	ack,
}) => {
	const nextRunnable = new Date(next_runnable);
	const now = new Date();
	const [sendable, setSendable] = useState(!sending && nextRunnable <= now);

	// Disable resend button until nextRunnable.
	useEffect(() => {
		const nextRunnable = new Date(next_runnable);
		const now = new Date();

		return sendableTimeout({
			nextRunnable: nextRunnable,
			now: now,
			sendable: sendable,
			sending: sending,
			setSendable: setSendable,
		});
	}, [sendable, sending, next_runnable]);

	const id = title.replace(' ', '_').toLowerCase();

	return (
		<Card className="flex min-h-12 flex-row items-center justify-between gap-4 px-4 py-2">
			<CardHeader className="flex w-full items-center p-0">
				<CardTitle className="w-full" key={`${title}-title`}>
					{loading ? <Skeleton className={cn('h-6 w-full')} /> : title}
				</CardTitle>
			</CardHeader>
			<CardContent className="justify-none flex flex-row items-center gap-2 px-0">
				{!sendable && !sending && (
					<Tip
						content={
							<p>
								{`Can resend ${formatRelative(
									new Date(next_runnable),
									new Date(),
								)}`}
							</p>
						}
					>
						<Hourglass className="size-5" />
					</Tip>
				)}
				{!sending && failed !== undefined && <StatusIcon failed={failed} />}
				{/* Send button */}
				{modalType !== 'SKIP' && (
					<Button
						aria-describedby={cn(
							`${id}-send`,
							!sendable && !sending && `${id}-resend-date`,
							!sending && failed !== undefined && `${id}-result`,
						)}
						aria-label={
							modalType === 'RESEND' || failed !== undefined ? 'Retry' : 'Send'
						}
						disabled={loading || !sendable}
						onClick={() => ack(title, itemType === 'WEBHOOK')}
						size="sm"
						// Disable if success or waiting send response.
						variant="secondary"
					>
						<SendingIcon
							failed={failed}
							modalType={modalType}
							sending={sending}
						/>
					</Button>
				)}
			</CardContent>
		</Card>
	);
};
