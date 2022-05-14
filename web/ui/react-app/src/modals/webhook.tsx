import {
  Button,
  Card,
  Container,
  Modal,
  OverlayTrigger,
  Placeholder,
  Tooltip,
} from "react-bootstrap";
import { addMessageHandler, sendMessage } from "contexts/websocket";
import {
  faCheck,
  faCircleNotch,
  faPaperPlane,
  faRedo,
  faSquareFull,
  faTimes,
} from "@fortawesome/free-solid-svg-icons";
import { useCallback, useContext, useEffect, useReducer } from "react";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { ModalContext } from "contexts/modal";
import { WebHookSummaryListType } from "types/summary";
import formatRelative from "date-fns/formatRelative";
import reducerWebHookModal from "reducers/webhook";
import { useDelayedRender } from "hooks/delayed-render";
import { useTheme } from "contexts/theme";
import { websocketResponse } from "types/websocket";

const WebHookModal = () => {
  // modal.type:
  // RESEND - 0 WebHooks failed. Resend Modal
  // SEND   - Send WebHooks for this new version. New release Modal
  // SKIP   - Release not wanted. Skip release Modal
  // RETRY  - !+ WebHooks failed to send. Retry send Modal
  const { handleModal, modal } = useContext(ModalContext);
  const delayedRender = useDelayedRender(250);
  const [modalData, setModalData] = useReducer(reducerWebHookModal, {
    service_id: "",
    sent: [],
    webhooks: {},
  });
  const themeCtx = useTheme();

  const hideModal = useCallback(() => {
    setModalData({ page: "APPROVALS", type: "WEBHOOK", sub_type: "RESET" });
    handleModal("", { id: "", loading: true });
  }, [handleModal]);

  const onClickAcknowledge = useCallback(
    (target: string) => {
      const unspecificWebHook = [
        "ARGUS_ALL",
        "ARGUS_FAILED",
        "ARGUS_SKIP",
      ].includes(target);

      if (!(sendingThisService && unspecificWebHook)) {
        console.log(`Approving ${modal.service.id}`);
        sendMessage(
          JSON.stringify({
            version: 1,
            page: "APPROVALS",
            type: "VERSION",
            target: target,
            service_data: {
              id: `${modal.service.id}`,
              status: {
                latest_version: modal.service?.status?.latest_version
                  ? modal.service?.status?.latest_version
                  : "LATEST",
              },
            },
          })
        );

        setModalData({
          page: "APPROVALS",
          type: "WEBHOOK",
          sub_type: target === "ALL" ? "RESENDING" : "SENDING",
          service_data: { id: modal.service.id, loading: false },
          webhook_data: unspecificWebHook ? {} : { [target]: {} },
        });
      }

      if (unspecificWebHook) {
        hideModal();
      }
    },
    [modal.service, hideModal]
  );

  useEffect(() => {
    if (modal.service.id !== "") {
      // Handler to listen to WebSocket messages
      const handler = (event: websocketResponse) => {
        if (event) {
          if (event.type === "WEBHOOK") {
            setModalData(event);
          }
        }
      };
      addMessageHandler("webhook-modal", handler);

      sendMessage(
        JSON.stringify({
          version: 1,
          page: "APPROVALS",
          type: "WEBHOOK",
          sub_type: "SUMMARY",
          service_data: {
            id: modal.service.id,
          },
        })
      );
    }
  }, [modal.type, modal.service.id]);
  let sendingThisService = false;

  return (
    <Modal show={modal.type !== ""} onHide={() => hideModal()}>
      <Modal.Header
        closeButton
        closeVariant={themeCtx.theme === "theme-dark" ? "white" : undefined}
      >
        <Modal.Title>
          <strong>
            {modal.type === "RESEND"
              ? "Resend the WebHook(s)?"
              : modal.type === "SEND"
              ? "Send the WebHook(s) to upgrade?"
              : modal.type === "SKIP"
              ? "Skip this release? (don't send any WebHooks)"
              : ""}
          </strong>
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <Container
          fluid
          className="font-weight-bold"
          style={{ paddingLeft: "0px" }}
        >
          <strong>{modal.service.id}</strong>
          {modal.type === "RESEND"
            ? modal.service?.status?.latest_version
              ? ` - ${modal.service?.status?.latest_version}`
              : " - Unknown"
            : ""}
        </Container>
        <>
          {modal.type !== "RESEND" && (
            <>
              <OverlayTrigger
                key="from-version"
                placement="top"
                delay={{ show: 500, hide: 500 }}
                overlay={
                  <Tooltip id={`tooltip-current-version`}>
                    {modal.service?.status?.current_version_timestamp ? (
                      <>
                        {formatRelative(
                          new Date(
                            modal.service?.status?.current_version_timestamp
                          ),
                          new Date()
                        )}
                      </>
                    ) : (
                      <>Unknown</>
                    )}
                  </Tooltip>
                }
              >
                <p style={{ margin: 0 }}>
                  {`${modal.type === "SKIP" ? "Stay on" : "From"}: ${
                    modal.service?.status?.current_version
                  }`}
                </p>
              </OverlayTrigger>
              <OverlayTrigger
                key="to-version"
                placement="bottom"
                delay={{ show: 500, hide: 500 }}
                overlay={
                  <Tooltip id={`tooltip-latest-version`}>
                    {modal.service?.status?.latest_version_timestamp ? (
                      <>
                        {formatRelative(
                          new Date(
                            modal.service?.status?.latest_version_timestamp
                          ),
                          new Date()
                        )}
                      </>
                    ) : (
                      <>Unknown</>
                    )}
                  </Tooltip>
                }
              >
                <p style={{ margin: 0 }}>
                  {`${modal.type === "SKIP" ? "Skip" : "To"}: ${
                    modal.service?.status?.latest_version
                  }`}
                </p>
              </OverlayTrigger>
            </>
          )}
          <br />
          <strong>WebHook(s):</strong>
          <Container fluid className="webhooks">
            {Object.keys(modalData.webhooks ? modalData.webhooks : {})
              .length === 0
              ? [...Array.from(Array(modal.service.webhook).keys())].map(
                  (num) => (
                    <Card
                      key={num}
                      bg="secondary"
                      className={"no-margin service"}
                    >
                      <Card.Title className="webhook-title">
                        <Container fluid style={{ paddingLeft: "0px" }}>
                          {delayedRender(() => (
                            <Placeholder xs={4} />
                          ))}
                        </Container>

                        {modal.type !== "SKIP" && (
                          <Button
                            variant="secondary"
                            size="sm"
                            className="float-end"
                            // Disable if success or waiting send response
                            disabled
                          >
                            <FontAwesomeIcon icon={faSquareFull} />
                          </Button>
                        )}
                      </Card.Title>
                    </Card>
                  )
                )
              : Object.entries(
                  modalData.webhooks as WebHookSummaryListType
                ).map(([id, webhook]) => {
                  let sending = false;
                  if (modalData.sent.includes(`${modal.service.id} ${id}`)) {
                    sendingThisService = true;
                    sending = true;
                  }
                  // const sending = modalData.sent.includes(
                  //   `${modal.service.id} ${id}`
                  // );
                  return (
                    <Card
                      key={id}
                      bg="secondary"
                      className={"no-margin service"}
                    >
                      <Card.Title className="webhook-title" key={id + "-title"}>
                        <Container fluid style={{ paddingLeft: "0px" }}>
                          {id}
                        </Container>
                        {!sending && webhook.failed !== undefined && (
                          <OverlayTrigger
                            key="status"
                            placement="top"
                            delay={{ show: 500, hide: 500 }}
                            overlay={
                              <Tooltip id={`tooltip-status`}>
                                {webhook.failed === true
                                  ? "Send failed"
                                  : "Sent successfully"}
                              </Tooltip>
                            }
                          >
                            <Container
                              fluid
                              style={{
                                display: "flex",
                                justifyContent: "flex-end",
                                width: "auto",
                                paddingRight:
                                  modal.type === "SKIP" ? "0px" : "",
                              }}
                            >
                              <FontAwesomeIcon
                                icon={
                                  webhook.failed === true ? faTimes : faCheck
                                }
                                style={{
                                  height: "2rem",
                                }}
                                className={
                                  webhook.failed === true
                                    ? "icon-danger"
                                    : "icon-success"
                                }
                              />
                            </Container>
                          </OverlayTrigger>
                        )}

                        {/* Send WebHook button */}
                        {modal.type !== "SKIP" && (
                          <OverlayTrigger
                            key="send"
                            placement="top"
                            delay={{ show: 500, hide: 500 }}
                            overlay={
                              <Tooltip id={`tooltip-send`}>
                                {modal.type === "RESEND"
                                  ? "Resend"
                                  : modal.type === "SEND"
                                  ? webhook.failed === true
                                    ? "Resend"
                                    : "Send"
                                  : modal.type === "RETRY"
                                  ? "Retry"
                                  : ""}
                              </Tooltip>
                            }
                          >
                            <Button
                              key={id}
                              variant="secondary"
                              size="sm"
                              onClick={() => onClickAcknowledge(id)}
                              className="float-end"
                              // Disable if success or waiting send response
                              disabled={sending || webhook.failed === false}
                            >
                              <FontAwesomeIcon
                                icon={
                                  sending && webhook.failed === undefined
                                    ? faCircleNotch
                                    : modal.type === "SEND" &&
                                      webhook.failed !== true
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
                })}
          </Container>
        </>
      </Modal.Body>
      <Modal.Footer>
        <Button
          id="modal-cancel"
          variant="secondary"
          hidden={sendingThisService}
          onClick={() => hideModal()}
        >
          Cancel
        </Button>
        <Button
          id="modal-action"
          variant="primary"
          onClick={() =>
            sendingThisService
              ? hideModal()
              : modal.type === "RESEND"
              ? onClickAcknowledge("ARGUS_ALL")
              : modal.type === "SEND"
              ? onClickAcknowledge("ARGUS_FAILED")
              : modal.type === "RETRY"
              ? onClickAcknowledge("ARGUS_FAILED")
              : modal.type === "SKIP"
              ? onClickAcknowledge("ARGUS_SKIP")
              : hideModal()
          }
          disabled={modal.type === "SKIP" && sendingThisService}
        >
          {sendingThisService
            ? modal.type === "SKIP"
              ? "Still sending..."
              : "Done"
            : modal.type === "RESEND"
            ? "Resend all"
            : modal.type === "SEND"
            ? "Confirm"
            : modal.type === "RETRY"
            ? "Retry all failed"
            : modal.type === "SKIP"
            ? "Skip release"
            : ""}
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default WebHookModal;
