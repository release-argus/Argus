import {
  Button,
  Card,
  Container,
  OverlayTrigger,
  Tooltip,
} from "react-bootstrap";
import { ReactElement, useEffect, useState } from "react";
import {
  faCheck,
  faCircleNotch,
  faHourglass,
  faPaperPlane,
  faRedo,
  faTimes,
} from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { ModalType } from "types/summary";
import differenceInMilliseconds from "date-fns/differenceInMilliseconds";
import formatRelative from "date-fns/formatRelative";

interface params {
  itemType: "COMMAND" | "WEBHOOK";
  modalType: ModalType;
  title: string;
  failed?: boolean;
  sending: boolean;
  next_runnable: string;
  ack: (target: string, isWebHook: boolean) => void;
}

export const Item = ({
  itemType,
  modalType,
  title,
  failed,
  sending,
  next_runnable,
  ack,
}: params): ReactElement => {
  const nextRunnable = new Date(next_runnable);
  const now = new Date();
  const [sendable, setSendable] = useState(!sending && nextRunnable <= now);

  // disable resend button until nextRunnable
  useEffect(() => {
    if (sending) {
      setSendable(false);
    } else if (!sendable) {
      let timeout = differenceInMilliseconds(nextRunnable, now);
      // if we're already after nextRunnable
      if (now > nextRunnable) {
        // just wait a second
        timeout = 1000;
      }
      const timer = setTimeout(function () {
        setSendable(true);
      }, timeout);
      return () => {
        clearTimeout(timer);
      };
    }
  }, [next_runnable, sending]);

  return (
    <Card bg="secondary" className={"no-margin service"}>
      <Card.Title className="modal-item-title" key={title + "-title"}>
        <Container fluid style={{ paddingLeft: "0px" }}>
          {title}
        </Container>
        {!sendable && !sending && (
          <OverlayTrigger
            key="resend-unavailable"
            placement="top"
            delay={{ show: 500, hide: 500 }}
            overlay={
              <Tooltip id={`tooltip-status`}>
                {`Can resend ${formatRelative(
                  new Date(next_runnable),
                  new Date()
                )}`}
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
                icon={faHourglass}
                style={{
                  height: "1.25rem",
                }}
                transform="right-8"
              />
            </Container>
          </OverlayTrigger>
        )}

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
                {modalType === "RESEND" || failed !== undefined
                  ? "Retry"
                  : "Send"}
              </Tooltip>
            }
          >
            <Button
              variant="secondary"
              size="sm"
              onClick={() => ack(title, itemType === "WEBHOOK")}
              className="float-end"
              // Disable if success or waiting send response
              disabled={!sendable}
            >
              <FontAwesomeIcon
                icon={
                  sending && failed === undefined
                    ? faCircleNotch
                    : modalType === "SEND" && failed === undefined
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
