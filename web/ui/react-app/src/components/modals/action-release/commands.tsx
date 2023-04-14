import { ActionModalData, ModalType } from "types/summary";

import { Command } from "components/modals/action-release/command";
import { Container } from "react-bootstrap";
import { FC } from "react";
import { Loading } from "components/modals/action-release/loading";

interface Props {
  modalType: ModalType;
  data: ActionModalData;
  onClickAcknowledge: (target: string) => void;
  setSendingThisService: React.Dispatch<React.SetStateAction<boolean>>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  delayedRender: any;
}

export const Commands: FC<Props> = ({
  modalType,
  data,
  onClickAcknowledge,
  setSendingThisService,
  delayedRender,
}) => {
  return (
    <>
      <strong>Command(s):</strong>
      <Container fluid className="command">
        {Object.keys(data.commands ? data.commands : {}).length === 0
          ? [...Array.from(Array(data.commands).keys())].map((id) => (
              <Loading
                key={id}
                modalType={modalType}
                delayedRender={delayedRender}
              />
            ))
          : Object.entries(data.commands).map(([id, command]) => {
              let sending = false;
              if (data.sentC.includes(`${data.service_id} ${id}`)) {
                setSendingThisService(true);
                sending = true;
              }
              return (
                <Command
                  key={id}
                  modalType={modalType}
                  name={id}
                  command={command}
                  sending={sending}
                  ack={onClickAcknowledge}
                />
              );
            })}
      </Container>
    </>
  );
};
