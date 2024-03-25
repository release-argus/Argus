import {
  CommandSummaryListType,
  ModalType,
  WebHookSummaryListType,
} from "types/summary";

import { Container } from "react-bootstrap";
import { FC } from "react";
import { Item } from "components/modals/action-release/item";
import { Loading } from "components/modals/action-release/loading";

interface Props {
  itemType: "COMMAND" | "WEBHOOK";
  modalType: ModalType;
  serviceID: string;
  data: CommandSummaryListType | WebHookSummaryListType;
  sent: string[];
  onClickAcknowledge: (target: string) => void;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  delayedRender: any;
}

/**
 * ModalList renders a list of items to approve
 *
 * @param itemType - The type of item to render
 * @param modalType - The type of modal to render
 * @param serviceID - The ID of the service
 * @param data - The data to render for that type
 * @param sent - The list of sent items
 * @param onClickAcknowledge - The function to call when an item is acknowledged
 * @param delayedRender - The delayed render function
 * @returns A list of `itemType` items to approve
 */
export const ModalList: FC<Props> = ({
  itemType,
  modalType,
  serviceID,
  data,
  sent,
  onClickAcknowledge,
  delayedRender,
}) => {
  // LOADING
  if (Object.keys(data ? data : {}).length === 0)
    return (
      <Container fluid className="list">
        {[...Array.from(Array(data).keys())].map((id) => (
          <Loading
            key={id}
            modalType={modalType}
            delayedRender={delayedRender}
          />
        ))}
      </Container>
    );

  return (
    <Container fluid className="list">
      {Object.entries(data).map(([title, item]) => {
        const sending = sent.includes(`${serviceID} ${title}`);
        return (
          <Item
            itemType={itemType}
            key={title}
            modalType={modalType}
            title={title}
            failed={item.failed}
            sending={sending}
            next_runnable={item.next_runnable ?? ""}
            ack={onClickAcknowledge}
          />
        );
      })}
    </Container>
  );
};
