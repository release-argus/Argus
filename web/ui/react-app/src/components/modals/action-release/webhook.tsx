import {
  Button,
  Card,
  Container,
  OverlayTrigger,
  Tooltip,
} from "react-bootstrap";
import { ModalType, WebHookSummaryType } from "types/summary";
import {
  faCheck,
  faCircleNotch,
  faPaperPlane,
  faRedo,
  faTimes,
} from "@fortawesome/free-solid-svg-icons";

import { FC } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

interface Props {
  modalType: ModalType;
  name: string;
  webhook: WebHookSummaryType;
  sending: boolean;
  ack: (target: string, isWebHook: boolean) => void;
}

export const WebHook: FC<Props> = ({
  modalType,
  name,
  webhook,
  sending,
  ack,
}) => {
  return (
    <Card bg="secondary" className={"no-margin service"}>
      <Card.Title className="modal-item-title" key={name + "-title"}>
        <Container fluid style={{ paddingLeft: "0px" }}>
          {name}
        </Container>
        {!sending && webhook.failed !== undefined && (
          <OverlayTrigger
            key="status"
            placement="top"
            delay={{ show: 500, hide: 500 }}
            overlay={
              <Tooltip id={`tooltip-status`}>
                {webhook.failed === true ? "Send failed" : "Sent successfully"}
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
                icon={webhook.failed === true ? faTimes : faCheck}
                style={{
                  height: "2rem",
                }}
                className={
                  webhook.failed === true ? "icon-danger" : "icon-success"
                }
              />
            </Container>
          </OverlayTrigger>
        )}

        {/* Send WebHook button */}
        {modalType !== "SKIP" && (
          <OverlayTrigger
            key="send"
            placement="top"
            delay={{ show: 500, hide: 500 }}
            overlay={
              <Tooltip id={`tooltip-send`}>
                {modalType === "RESEND"
                  ? "Resend"
                  : modalType === "SEND"
                  ? webhook.failed === true
                    ? "Resend"
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
              onClick={() => ack(name, true)}
              className="float-end"
              // Disable if success or waiting send response
              disabled={sending || webhook.failed === false}
            >
              <FontAwesomeIcon
                icon={
                  sending && webhook.failed === undefined
                    ? faCircleNotch
                    : modalType === "SEND" && webhook.failed !== true
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
