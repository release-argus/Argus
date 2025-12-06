import type { FC } from 'react';
import { Item } from '@/components/modals/action-release/item';
import type {
	CommandSummaryListType,
	ModalType,
	WebHookSummaryListType,
} from '@/utils/api/types/config/summary';

type ModalListCommandProps = {
	/* The type of item to render. */
	itemType: 'COMMAND';
	/* The data to render for that type. */
	data: CommandSummaryListType;
};
type ModalListWebHookProps = {
	/* The type of item to render. */
	itemType: 'WEBHOOK';
	/* The data to render for that type. */
	data: WebHookSummaryListType;
};

type ModalListTypeProps = ModalListCommandProps | ModalListWebHookProps;

type ModalListProps = ModalListTypeProps & {
	/* Defines the kind of modal to render. */
	modalType: ModalType;
	/* The ID of the service. */
	serviceID: string;
	/* The list of sent items. */
	sent: string[];
	/* Function called when the user acknowledges an item. */
	onClickAcknowledge: (target: string) => void;
};

/**
 * A list of items to approve.
 *
 * @param itemType - The type of item to render.
 * @param modalType - Defines the kind of modal to render.
 * @param serviceID - The ID of the service.
 * @param data - The data to render for that type.
 * @param sent - The list of sent items.
 * @param onClickAcknowledge - Function called when the user acknowledges an item.
 * @returns A list of `itemType` items to approve.
 */
export const ModalList: FC<ModalListProps> = ({
	itemType,
	modalType,
	serviceID,
	data,
	sent,
	onClickAcknowledge,
}) => {
	return (
		<div className="flex flex-col gap-1">
			{Object.entries(data).map(([title, item]) => {
				const sending = sent.includes(`${serviceID} ${title}`);
				return (
					<Item
						ack={onClickAcknowledge}
						failed={item.failed}
						itemType={itemType}
						key={title}
						loading={item.loading}
						modalType={modalType}
						next_runnable={item.next_runnable ?? ''}
						sending={sending}
						title={title}
					/>
				);
			})}
		</div>
	);
};
