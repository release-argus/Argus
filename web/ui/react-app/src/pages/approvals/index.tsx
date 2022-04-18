import { ReactElement, useEffect } from "react";
import { sendMessage, useWebSocket } from "contexts/websocket";

import { Container } from "react-bootstrap";
import { Service } from "components/approvals/service";

export const Approvals = (): ReactElement => {
  const { monitorData } = useWebSocket();

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

  return (
    <Container
      fluid
      className="services"
      style={{
        maxWidth:
          Object.entries(monitorData.service).length === 1 ? "500px" : "",
        padding: 0,
      }}
    >
      {!(
        monitorData.order.length === 1 &&
        monitorData.order[0] === "monitorData_loading"
      ) &&
        Object.entries(monitorData.service).map(([id, service]) => (
          <Service key={id} service={service} />
        ))}
    </Container>
  );
};
