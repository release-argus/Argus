import { ReactElement, useEffect, useMemo, useState } from "react";
import { sendMessage, useWebSocket } from "contexts/websocket";

import ApprovalsToolbar from "components/approvals/toolbar";
import { ApprovalsToolbarOptions } from "types/util";
import { Container } from "react-bootstrap";
import Service from "components/approvals/service";
import useLocalStorage from "hooks/local-storage";

export const Approvals = (): ReactElement => {
  const { monitorData } = useWebSocket();
  const toolbarDefaults: ApprovalsToolbarOptions = {
    search: "",
    editMode: false,
    hide: [3],
  };
  const [toolbarOptionsLS, setLSToolbarOptions] =
    useLocalStorage<ApprovalsToolbarOptions>("toolbarOptions", toolbarDefaults);
  const [toolbarOptions, setToolbarOptions] = useState(toolbarOptionsLS);

  // Keep local storage and state in sync
  useEffect(() => {
    setLSToolbarOptions({
      search: toolbarDefaults.search,
      editMode: toolbarOptions.editMode,
      hide: toolbarOptions.hide,
    });
  }, [toolbarOptions]);

  useEffect(() => {
    // Starting on /status and switching to /approvals never has data
    //
    // INIT for approvals is requested on all pages, but the message fails
    // to parse correctly
    monitorData.order.length === 1 &&
      monitorData.order[0] === "monitorData_loading" &&
      sendMessage(
        JSON.stringify({
          version: 1,
          page: "APPROVALS",
          type: "INIT",
        })
      );
  }, [monitorData.order]);

  const filteredServices = useMemo(
    () =>
      Object.values(monitorData.order)
        .filter((service) => {
          if (
            service.includes(toolbarOptions.search) &&
            monitorData.service[service]
          ) {
            const svc = monitorData.service[service];
            const skipped =
              `SKIP_${svc.status?.latest_version}` ===
              svc.status?.approved_version;
            const upToDate =
              svc.status?.deployed_version === svc.status?.latest_version;
            return (
              // hideUpToDate - deployed_version NOT latest_version
              (!toolbarOptions.hide.includes(0) || !upToDate) &&
              // hideUpdatable - deployed_version IS latest_version AND approved_version IS NOT "SKIP_"+latest_version
              (!toolbarOptions.hide.includes(1) || upToDate || skipped) &&
              // hideSkipped - approved_version NOT "SKIP_"+latest_version OR NO approved_version
              (!toolbarOptions.hide.includes(2) || !skipped) &&
              // hideInactive - active NOT false
              (!toolbarOptions.hide.includes(3) || svc.active !== false)
            );
          }
        })
        .map((service) => monitorData.service[service]),
    [toolbarOptions, monitorData.service, monitorData.order]
  );

  return (
    <>
      <ApprovalsToolbar values={toolbarOptions} setValues={setToolbarOptions} />
      <Container
        fluid
        className="services"
        style={{
          maxWidth: filteredServices.length === 1 ? "500px" : "",
          padding: 0,
        }}
      >
        {!(
          monitorData.order.length === 1 &&
          monitorData.order[0] === "monitorData_loading"
        ) &&
          filteredServices.map((service) => (
            <Service
              key={service.id}
              service={service}
              editable={toolbarOptions.editMode}
            />
          ))}
      </Container>
    </>
  );
};
