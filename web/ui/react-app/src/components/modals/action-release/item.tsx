import {
  Button,
  Card,
  Container,
  OverlayTrigger,
  Tooltip,
} from "react-bootstrap";
import {
  faCheck,
  faCircleNotch,
  faPaperPlane,
  faRedo,
  faTimes,
} from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { ModalType } from "types/summary";
import { ReactElement } from "react";

interface params {
  itemType: "COMMAND" | "WEBHOOK";
  modalType: ModalType;
  title: string;
  failed?: boolean;
  sending: boolean;
  ack: (target: string, isWebHook: boolean) => void;
}

export const Item = ({
  itemType,
  modalType,
  title,
  failed,
  sending,
  ack,
}: params): ReactElement => {
  return (
    <Card bg="secondary" className={"no-margin service"}>
      <Card.Title className="modal-item-title" key={title + "-title"}>
        <Container fluid style={{ paddingLeft: "0px" }}>
          {title}
        </Container>
        {!sending && failed !== undefined && (
          <OverlayTrigger
            key="status"
            placement="top"
            delay={{ show: 500, hide: 500 }}
            overlay={
              <Tooltip id={`tooltip-status`}>
                {failed === true ? "Failed" : "Successful"}
              </Tooltip>
            }
          >
            <Container
              fluid
              style={{
                display: "flex",
                justifyContent: "flex-end",
                width: "auto",
                paddingRight: modalType === "SKIP" ? "0px" : "",
              }}
            >
              <FontAwesomeIcon
                icon={failed === true ? faTimes : faCheck}
                style={{
                  height: "2rem",
                }}
                className={failed === true ? "icon-danger" : "icon-success"}
              />
            </Container>
          </OverlayTrigger>
        )}

        {/* Send button */}
        {modalType !== "SKIP" && (
          <OverlayTrigger
            key="send"
            placement="top"
            delay={{ show: 500, hide: 500 }}
            overlay={
              <Tooltip id={`tooltip-send`}>
                {modalType === "RESEND"
                  ? "Retry"
                  : modalType === "SEND"
                  ? failed === true
                    ? "Retry"
                    : "Send"
                  : modalType === "RETRY"
                  ? "Retry"
                  : ""}
              </Tooltip>
            }
          >
            <Button
              variant="secondary"
              size="sm"
              onClick={() => ack(title, itemType === "WEBHOOK")}
              className="float-end"
              // Disable if success or waiting send response
              disabled={sending || failed === false}
            >
              <FontAwesomeIcon
                icon={
                  sending && failed === undefined
                    ? faCircleNotch
                    : modalType === "SEND" && failed !== true
                    ? faPaperPlane
                    : faRedo
                }
                className={sending ? "fa-spin" : ""}
              />
            </Button>
          </OverlayTrigger>
        )}
      </Card.Title>
    </Card>
  );
};
