import { Alert } from "react-bootstrap";
import { BooleanType } from "types/boolean";
import { FC } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { WS_ADDRESS } from "config";
import { faCircleNotch } from "@fortawesome/free-solid-svg-icons";
import { useDelayedRender } from "hooks/delayed-render";

interface Props {
  connected: BooleanType;
}

export const WebSocketStatus: FC<Props> = ({ connected }) => {
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
