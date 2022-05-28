import { ActionModalData, ModalType } from "types/summary";

import { Container } from "react-bootstrap";
import { Loading } from "components/modals/action-release/loading";
import { ReactElement } from "react";
import { WebHook } from "components/modals/action-release/webhook";

interface params {
  modalType: ModalType;
  data: ActionModalData;
  onClickAcknowledge: (target: string, isWebHook: boolean) => void;
  setSendingThisService: React.Dispatch<React.SetStateAction<boolean>>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  delayedRender: any;
}

export const WebHooks = ({
  modalType,
  data,
  onClickAcknowledge,
  setSendingThisService,
  delayedRender,
}: params) => {
  return (
    <>
      <strong>WebHook(s):</strong>
      <Container fluid className="webhooks">
        {Object.keys(data.webhooks ? data.webhooks : {}).length === 0
          ? [...Array.from(Array(data.webhooks).keys())].map((id) => (
              <Loading
                key={id}
                modalType={modalType}
                delayedRender={delayedRender}
              />
            ))
          : Object.entries(data.webhooks).map(([id, webhook]): ReactElement => {
              let sending = false;
              if (data.sentWH.includes(`${data.service_id} ${id}`)) {
                setSendingThisService(true);
                sending = true;
              }
              return (
                <WebHook
                  key={id}
                  modalType={modalType}
                  name={id}
                  webhook={webhook}
                  sending={sending}
                  ack={onClickAcknowledge}
                />
              );
            })}
      </Container>
    </>
  );
};
