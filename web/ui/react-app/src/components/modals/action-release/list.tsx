import {
  CommandSummaryListType,
  ModalType,
  WebHookSummaryListType,
} from "types/summary";

import { Container } from "react-bootstrap";
import { Item } from "components/modals/action-release/item";
import { Loading } from "components/modals/action-release/loading";
import { ReactElement } from "react";

interface params {
  itemType: "COMMAND" | "WEBHOOK";
  modalType: ModalType;
  serviceID: string;
  data: CommandSummaryListType | WebHookSummaryListType;
  sent: string[];
  onClickAcknowledge: (target: string) => void;
  setSendingThisService: React.Dispatch<React.SetStateAction<boolean>>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  delayedRender: any;
}

export const ModalList = ({
  itemType,
  modalType,
  serviceID,
  data,
  sent,
  onClickAcknowledge,
  setSendingThisService,
  delayedRender,
}: params) => {
  return (
    <Container fluid className="list">
      {Object.keys(data ? data : {}).length === 0
        ? [...Array.from(Array(data).keys())].map((id) => (
            <Loading
              key={id}
              modalType={modalType}
              delayedRender={delayedRender}
            />
          ))
        : Object.entries(data).map(([title, item]): ReactElement => {
            let sending = false;
            if (sent.includes(`${serviceID} ${title}`)) {
              setSendingThisService(true);
              sending = true;
            }
            return (
              <Item
                itemType={itemType}
                key={title}
                modalType={modalType}
                title={title}
                failed={item.failed}
                sending={sending}
                next_runnable={item.next_runnable || ""}
                ack={onClickAcknowledge}
              />
            );
          })}
    </Container>
  );
};
