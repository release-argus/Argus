import { Alert } from "react-bootstrap";
import { BooleanType } from "types/boolean";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { ReactElement } from "react";
import { WS_ADDRESS } from "config";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons";
import { useDelayedRender } from "hooks/delayed-render";

interface params {
  connected: BooleanType;
}

export const WebSocketStatus = ({ connected }: params): ReactElement => {
  const delayedRender = useDelayedRender(1000);
  const fallback = (
    <Alert variant={connected === false ? "danger" : "info"}>
      <Alert.Heading>
        WebSocket{" "}
        {connected === false ? "Disconnected! Reconnecting" : "connecting"}
      </Alert.Heading>
      <>
        <FontAwesomeIcon
          icon={faCircleNotch}
          className="fa-spin"
          style={{ marginRight: "0.5rem" }}
        />
        Connecting to {WS_ADDRESS}...
      </>
    </Alert>
  );
  return connected !== true && delayedRender(() => fallback);
};
