import {
  ActionAPIType,
  CommandSummaryListType,
  WebHookSummaryListType,
} from "types/summary";
import {
  Button,
  Container,
  Modal,
  OverlayTrigger,
  Tooltip,
} from "react-bootstrap";
import { dateIsAfterNow, fetchJSON } from "utils";
import { useCallback, useContext, useEffect, useMemo, useReducer } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";

import { ModalContext } from "contexts/modal";
import { ModalList } from "components/modals/action-release/list";
import { WebSocketResponse } from "types/websocket";
import { addMessageHandler } from "contexts/websocket";
import { formatRelative } from "date-fns";
import reducerActionModal from "reducers/action-release";
import { useDelayedRender } from "hooks/delayed-render";
import { useTheme } from "contexts/theme";

const isSendingService = (
  serviceName: string,
  sentCommands: string[],
  sentWebHooks: string[]
) => {
  const prefixStr = `${serviceName} `;
  for (const id of sentCommands) {
    if (id.startsWith(prefixStr)) return true;
  }
  for (const id of sentWebHooks) {
    if (id.startsWith(prefixStr)) return true;
  }
  return false;
};

const ActionReleaseModal = () => {
  // modal.actionType:
  // RESEND - 0 WebHooks failed. Resend Modal
  // SEND   - Send WebHooks for this new version. New release Modal
  // SKIP   - Release not wanted. Skip release Modal
  // RETRY  - 1+ WebHooks failed to send. Retry send Modal
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
    setModalData({ page: "APPROVALS", type: "ACTION", sub_type: "RESET" });
    handleModal("", { id: "", loading: true });
  }, []);

  // Handle deployed version becoming latest when there's no deployed version check
  // (close the modal)
  useEffect(() => {
    if (
      // Allow resend and edit/create modals to stay open
      !["RESEND", "EDIT"].includes(modal.actionType) &&
      modal.service?.status?.deployed_version &&
      modal.service?.status?.latest_version &&
      modal.service?.status?.deployed_version ===
        modal.service?.status?.latest_version
    )
      hideModal();
  }, [modal.actionType, modal.service?.status]);

  const isSendingThisService = useMemo(
    () => isSendingService(modal.service.id, modalData.sentC, modalData.sentWH),
    [modal.service.id, modalData]
  );
  const canSendUnspecific = useMemo(() => {
    // Currently sending/running an action for this service
    if (isSendingThisService) return false;
    // has no actions - allow unspecific (SKIP)
    if (
      Object.keys(modalData.commands).length === 0 &&
      Object.keys(modalData.webhooks).length === 0
    )
      return true;
    // has an action that's runnable
    return (
      Object.keys(modalData.commands).find((command_id) =>
        modalData.commands[command_id].next_runnable
          ? // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
            !dateIsAfterNow(modalData.commands[command_id].next_runnable!)
          : true
      ) ||
      Object.keys(modalData.webhooks).find((webhook_id) =>
        modalData.webhooks[webhook_id].next_runnable
          ? // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
            !dateIsAfterNow(modalData.webhooks[webhook_id].next_runnable!)
          : true
      ) ||
      false
    );
  }, [isSendingThisService, modalData]);

  const { mutate } = useMutation({
    mutationFn: (data: {
      target: string;
      service: string;
      isWebHook: boolean;
      unspecificTarget: boolean;
    }) =>
      fetch(`api/v1/service/actions/${encodeURIComponent(data.service)}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ target: data.target }),
      }),
    onMutate: (data) => {
      if (data.target === "ARGUS_SKIP") return;

      let command_data: CommandSummaryListType | undefined = {};
      let webhook_data: WebHookSummaryListType | undefined = {};
      if (!data.unspecificTarget) {
        // Targeting specific command/webhook
        if (data.isWebHook)
          webhook_data = { [data.target.slice("webhook_".length)]: {} };
        else command_data = { [data.target.slice("command_".length)]: {} };
      } else {
        // All Commands/WebHooks have been sent successfully
        const allSuccessful =
          Object.keys(modalData.commands).every(
            (command_id) => modalData.commands[command_id].failed === false
          ) &&
          Object.keys(modalData.webhooks).every(
            (webhook_id) => modalData.webhooks[webhook_id].failed === false
          );

        // sending these commands
        for (const command_id in modalData.commands) {
          // skip commands that aren't after next_runnable
          // and commands that have already succeeded if some commands haven't
          if (
            (modalData.commands[command_id].next_runnable !== undefined &&
              dateIsAfterNow(
                // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                modalData.commands[command_id].next_runnable!
              )) ||
            (!allSuccessful && modalData.commands[command_id].failed === false)
          )
            continue;
          command_data[command_id] = {};
        }

        // sending these webhooks
        for (const webhook_id in modalData.webhooks) {
          // skip webhooks that aren't after next_runnable
          // and webhooks that have already succeeded if some webhooks haven't
          if (
            (modalData.webhooks[webhook_id].next_runnable !== undefined &&
              dateIsAfterNow(
                // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
                modalData.webhooks[webhook_id].next_runnable!
              )) ||
            (!allSuccessful && modalData.webhooks[webhook_id].failed === false)
          )
            continue;
          webhook_data[webhook_id] = {};
        }
      }

      setModalData({
        page: "APPROVALS",
        type: "ACTION",
        sub_type: "SENDING",
        service_data: { id: modal.service.id, loading: false },
        command_data: command_data,
        webhook_data: webhook_data,
      });
    },
  });

  const onClickAcknowledge = useCallback(
    (target: string, isWebHook?: boolean) => {
      const unspecificTarget = [
        "ARGUS_ALL",
        "ARGUS_FAILED",
        "ARGUS_SKIP",
      ].includes(target);

      // don't allow unspecific non-skip targets if currently sending this service
      if (
        !(!canSendUnspecific && unspecificTarget && target !== "ARGUS_SKIP")
      ) {
        console.log(`Approving ${modal.service.id} - ${target}`);
        let approveTarget = target;
        if (!unspecificTarget)
          if (isWebHook) approveTarget = `webhook_${target}`;
          else approveTarget = `command_${target}`;
        mutate({
          service: modal.service.id,
          target: approveTarget,
          isWebHook: isWebHook === true,
          unspecificTarget: unspecificTarget,
        });
      }

      if (unspecificTarget) hideModal();
    },
    [modal.service, canSendUnspecific]
  );

  const { data } = useQuery<ActionAPIType>({
    queryKey: ["actions", { service: modal.service.id }],
    queryFn: () =>
      fetchJSON(
        `api/v1/service/actions/${encodeURIComponent(modal.service.id)}`
      ),
    enabled: modal.actionType !== "EDIT" && modal.service.id !== "",
    refetchOnMount: "always",
  });

  useEffect(
    () =>
      setModalData({
        page: "APPROVALS",
        type: "ACTION",
        sub_type: "REFRESH",
        service_data: { id: modal.service.id },

        webhook_data: data?.webhook,
        command_data: data?.command,
      }),
    [data]
  );

  useEffect(() => {
    if (modal.actionType !== "EDIT" && modal.service.id !== "") {
      // Handler to listen to WebSocket messages
      const handler = (event: WebSocketResponse) => {
        if (event && ["ACTIONS", "COMMAND", "WEBHOOK"].includes(event.type))
          setModalData(event);
      };
      addMessageHandler("action-modal", handler);
    }
  }, [modal.actionType, modal.service.id]);

  return (
    <Modal
      show={!["", "EDIT"].includes(modal.actionType)}
      onHide={() => hideModal()}
    >
      <Modal.Header
        closeButton
        closeVariant={themeCtx.theme === "theme-dark" ? "white" : undefined}
      >
        <Modal.Title>
          <strong>
            {modal.actionType === "RESEND"
              ? `Resend the ${
                  modal.service.webhook ? "WebHook" : "Command"
                }(s)?`
              : modal.actionType === "SEND"
              ? `Send the  ${
                  modal.service.webhook ? "WebHook" : "Command"
                }(s) to upgrade?`
              : modal.actionType === "SKIP"
              ? `Skip this release? (don't send any  ${
                  modal.service.webhook ? "WebHook" : "Command"
                }s)`
              : modal.actionType === "SKIP_NO_WH"
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
          {modal.actionType === "RESEND"
            ? modal.service?.status?.latest_version
              ? ` - ${modal.service?.status?.latest_version}`
              : " - Unknown"
            : ""}
        </Container>
        <>
          {modal.actionType !== "RESEND" && (
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
                  {`${modal.actionType === "SKIP" ? "Stay on" : "From"}: ${
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
                  {`${modal.actionType === "SKIP" ? "Skip" : "To"}: ${
                    modal.service?.status?.latest_version
                  }`}
                </p>
              </OverlayTrigger>
            </>
          )}
          {data?.webhook && Object.keys(data.webhook).length > 0 && (
            <>
              <br />
              <strong>WebHook(s):</strong>
              <ModalList
                itemType="WEBHOOK"
                modalType={modal.actionType}
                serviceID={modal.service.id}
                data={modalData.webhooks}
                sent={modalData.sentWH}
                onClickAcknowledge={onClickAcknowledge}
                delayedRender={delayedRender}
              />
            </>
          )}
          {data?.command && Object.keys(data.command).length > 0 && (
            <>
              <br />
              <strong>Command(s):</strong>
              <ModalList
                itemType="COMMAND"
                modalType={modal.actionType}
                serviceID={modal.service.id}
                data={modalData.commands}
                sent={modalData.sentC}
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
          hidden={!canSendUnspecific}
          onClick={() => hideModal()}
        >
          Cancel
        </Button>
        <Button
          id="modal-action"
          variant="primary"
          onClick={() => {
            if (
              !["SKIP", "SKIP_NO_WH"].includes(modal.actionType) &&
              !canSendUnspecific
            ) {
              hideModal();
              return;
            }
            switch (modal.actionType) {
              case "RESEND":
                onClickAcknowledge("ARGUS_ALL");
                break;
              case "SEND":
              case "RETRY":
                onClickAcknowledge("ARGUS_FAILED");
                break;
              case "SKIP":
              case "SKIP_NO_WH":
                onClickAcknowledge("ARGUS_SKIP");
                break;
            }
          }}
          disabled={modal.actionType !== "SKIP" && isSendingThisService}
        >
          {modal.actionType === "SKIP" || modal.actionType === "SKIP_NO_WH"
            ? "Skip release"
            : !canSendUnspecific
            ? "Done"
            : modal.actionType === "RESEND"
            ? "Resend all"
            : modal.actionType === "SEND"
            ? "Confirm"
            : modal.actionType === "RETRY"
            ? "Retry all failed"
            : ""}
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default ActionReleaseModal;
