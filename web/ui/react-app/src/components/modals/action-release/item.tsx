import {
  Button,
  Card,
  Container,
  OverlayTrigger,
  Tooltip,
} from "react-bootstrap";
import { FC, useEffect, useState } from "react";
import { differenceInMilliseconds, formatRelative } from "date-fns";
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

interface Props {
  itemType: "COMMAND" | "WEBHOOK";
  modalType: ModalType;
  title: string;
  failed?: boolean;
  sending: boolean;
  next_runnable: string;
  ack: (target: string, isWebHook: boolean) => void;
}

/**
 * Returns a function that sets a timeout to enable the item to be sent when it
 * is after the items nextRunnable time, or disables the item if it is sending.
 *
 * @param sendable - Whether the item can be sent
 * @param sending - Whether the item is being sent
 * @param setSendable - Function to set whether the item can be sent
 * @param now - The current time
 * @param nextRunnable - The time the item can be sent
 * @returns A function that sets a timeout to enable the item to be sent when it
 * is after the items nextRunnable time, or disables the item if it is sending
 */
const sendableTimeout = (
  sendable: boolean,
  sending: boolean,
  setSendable: React.Dispatch<React.SetStateAction<boolean>>,
  now: Date,
  nextRunnable: Date
) => {
  if (sending) setSendable(false);
  else if (!sendable) {
    let timeout = differenceInMilliseconds(nextRunnable, now);
    // if we're already after nextRunnable, just wait a second
    if (now > nextRunnable) timeout = 1000;
    const timer = setTimeout(function () {
      setSendable(true);
    }, timeout);
    return () => {
      clearTimeout(timer);
    };
  }
};

/**
 * Renders the item's information with buttons based on the modal type.
 *
 * @param itemType - The type of the item (e.g. COMMAND/WEBHOOK)
 * @param modalType - The type of the modal
 * @param title - The title of the item
 * @param failed - Whether the item failed
 * @param sending - Whether the item is being sent
 * @param next_runnable - The time the item can next be sent
 * @returns A component that displays the item's information with buttons based on the modal type
 */
export const Item: FC<Props> = ({
  itemType,
  modalType,
  title,
  failed,
  sending,
  next_runnable,
  ack,
}) => {
  const nextRunnable = new Date(next_runnable);
  const now = new Date();
  const [sendable, setSendable] = useState(!sending && nextRunnable <= now);

  // disable resend button until nextRunnable
  useEffect(() => {
    sendableTimeout(sendable, sending, setSendable, now, nextRunnable);
  }, [next_runnable, sending]);

  // add timeout if it wasn't sent by this user
  useEffect(() => {
    if (!sending && nextRunnable <= now)
      sendableTimeout(sendable, sending, setSendable, now, nextRunnable);
  }, []);

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
              }}
            >
              <FontAwesomeIcon
                icon={faHourglass}
                style={{
                  height: "1.25rem",
                }}
                transform={failed !== undefined ? "right-8" : ""}
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
                className={failed === true ? "text-danger" : "text-success"}
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
