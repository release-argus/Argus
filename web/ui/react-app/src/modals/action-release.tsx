import {
  Button,
  Container,
  Modal,
  OverlayTrigger,
  Tooltip,
} from "react-bootstrap";
import { addMessageHandler, sendMessage } from "contexts/websocket";
import {
  useCallback,
  useContext,
  useEffect,
  useReducer,
  useState,
} from "react";

import { ModalContext } from "contexts/modal";
import { ModalList } from "components/modals/action-release/list";
import formatRelative from "date-fns/formatRelative";
import reducerActionModal from "reducers/action-release";
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
  const [modalData, setModalData] = useReducer(reducerActionModal, {
    service_id: "",
    sentC: [],
    sentWH: [],
    webhooks: {},
    commands: {},
  });
  const themeCtx = useTheme();

  const hideModal = useCallback(() => {
    setSendingThisService(false);
    setModalData({ page: "APPROVALS", type: "ACTION", sub_type: "RESET" });
    handleModal("", { id: "", loading: true });
  }, [handleModal]);

  const onClickAcknowledge = useCallback(
    (target: string, isWebHook?: boolean) => {
      const unspecificTarget = [
        "ARGUS_ALL",
        "ARGUS_FAILED",
        "ARGUS_SKIP",
      ].includes(target);

      if (!(sendingThisService && unspecificTarget)) {
        console.log(`Approving ${modal.service.id}`);
        sendMessage(
          JSON.stringify({
            version: 1,
            page: "APPROVALS",
            type: "VERSION",
            target: unspecificTarget
              ? target
              : `${isWebHook ? "webhook" : "command"}_${target}`,
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

        unspecificTarget
          ? target === "ARGUS_SKIP"
            ? {}
            : // Sending ARGUS_ALL or ARGUS_FAILED
              setModalData({
                page: "APPROVALS",
                type: "ACTION",
                sub_type: "SENDING",
                service_data: { id: modal.service.id, loading: false },
                command_data: unspecificTarget ? {} : { [`${target}`]: {} },
              })
          : isWebHook
          ? // Sending WebHook X
            setModalData({
              page: "APPROVALS",
              type: "WEBHOOK",
              sub_type: "SENDING",
              service_data: { id: modal.service.id, loading: false },
              webhook_data: unspecificTarget ? {} : { [`${target}`]: {} },
            })
          : // Sending Command Y
            setModalData({
              page: "APPROVALS",
              type: "COMMAND",
              sub_type: "SENDING",
              service_data: { id: modal.service.id, loading: false },
              command_data: unspecificTarget ? {} : { [`${target}`]: {} },
            });
      }

      if (unspecificTarget) {
        hideModal();
      }
    },
    [modal.service, hideModal]
  );

  useEffect(() => {
    if (modal.service.id !== "") {
      // Handler to listen to WebSocket messages
      const handler = (event: websocketResponse) => {
        if (event && ["ACTIONS", "COMMAND", "WEBHOOK"].includes(event.type)) {
          setModalData(event);
        }
      };
      addMessageHandler("action-modal", handler);

      sendMessage(
        JSON.stringify({
          version: 1,
          page: "APPROVALS",
          type: "ACTIONS",
          sub_type: "SUMMARY",
          service_data: {
            id: modal.service.id,
          },
        })
      );
    }
  }, [modal.type, modal.service.id]);
  const [sendingThisService, setSendingThisService] = useState(false);

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
              : modal.type === "SKIP_NO_WH"
              ? "Skip this release?"
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
                  <Tooltip id={`tooltip-deployed-version`}>
                    {modal.service?.status?.deployed_version_timestamp ? (
                      <>
                        {formatRelative(
                          new Date(
                            modal.service?.status?.deployed_version_timestamp
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
                <p style={{ margin: 0, maxWidth: "fit-content" }}>
                  {`${modal.type === "SKIP" ? "Stay on" : "From"}: ${
                    modal.service?.status?.deployed_version
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
                <p style={{ margin: 0, maxWidth: "fit-content" }}>
                  {`${modal.type === "SKIP" ? "Skip" : "To"}: ${
                    modal.service?.status?.latest_version
                  }`}
                </p>
              </OverlayTrigger>
            </>
          )}
          {modal.type !== "SKIP_NO_WH" && modal.service.webhook && (
            <>
              <br />
              <strong>WebHook(s):</strong>
              <ModalList
                itemType="WEBHOOK"
                modalType={modal.type}
                serviceID={modalData.service_id}
                data={modalData.webhooks}
                sent={modalData.sentWH}
                setSendingThisService={setSendingThisService}
                onClickAcknowledge={onClickAcknowledge}
                delayedRender={delayedRender}
              />
            </>
          )}
          {modal.type !== "SKIP_NO_WH" && modal.service.command && (
            <>
              <br />
              <strong>Command(s):</strong>
              <ModalList
                itemType="COMMAND"
                modalType={modal.type}
                serviceID={modalData.service_id}
                data={modalData.commands}
                sent={modalData.sentC}
                setSendingThisService={setSendingThisService}
                onClickAcknowledge={onClickAcknowledge}
                delayedRender={delayedRender}
              />
            </>
          )}
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
              : modal.type === "SKIP" || modal.type === "SKIP_NO_WH"
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
            : modal.type === "SKIP" || modal.type === "SKIP_NO_WH"
            ? "Skip release"
            : ""}
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default WebHookModal;
