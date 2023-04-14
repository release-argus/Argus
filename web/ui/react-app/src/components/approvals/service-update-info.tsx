import {
  Card,
  Container,
  ListGroup,
  OverlayTrigger,
  Tooltip,
} from "react-bootstrap";
import { FC, memo } from "react";

import { ServiceSummaryType } from "types/summary";
import { formatRelative } from "date-fns";

interface Props {
  service: ServiceSummaryType;
  visible: boolean;
}

const UpdateInfo: FC<Props> = ({ service, visible }) => (
  <Container
    fluid
    style={{ display: visible ? "block" : "none", padding: "0px" }}
  >
    <ListGroup.Item
      key="latest_v"
      variant="secondary"
      className="service-item"
      style={{ height: "6rem" }}
    >
      <Container style={{ padding: "0px" }}>
        <OverlayTrigger
          key="from-version"
          placement="top"
          delay={{ show: 500, hide: 500 }}
          overlay={
            <Tooltip id={`tooltip-deployed-version`}>
              {service.status?.deployed_version_timestamp ? (
                <>
                  {formatRelative(
                    new Date(service.status.deployed_version_timestamp),
                    new Date()
                  )}
                </>
              ) : (
                <>Unknown</>
              )}
            </Tooltip>
          }
        >
          <p style={{ marginTop: 5, marginBottom: 5, maxWidth: "fit-content" }}>
            <strong>From:</strong>{" "}
            {service?.status?.deployed_version
              ? service.status.deployed_version
              : "Unknown"}
          </p>
        </OverlayTrigger>
        <OverlayTrigger
          key="to-version"
          placement="bottom"
          delay={{ show: 500, hide: 500 }}
          overlay={
            <Tooltip id={`tooltip-latest-version`}>
              {service.status?.latest_version_timestamp ? (
                <>
                  {formatRelative(
                    new Date(service.status.latest_version_timestamp),
                    new Date()
                  )}
                </>
              ) : (
                <>Unknown</>
              )}
            </Tooltip>
          }
        >
          <p style={{ marginBottom: 0, maxWidth: "fit-content" }}>
            <strong>To:</strong>{" "}
            {service?.status?.latest_version
              ? service.status.latest_version
              : "Unknown"}
          </p>
        </OverlayTrigger>
      </Container>
    </ListGroup.Item>
    <Card.Footer style={{ height: "1rem", paddingBottom: 0 }}>
      <small className="text-muted same-color">
        {service?.status && service.status?.latest_version_timestamp ? (
          <>
            Found{" "}
            {formatRelative(
              new Date(service.status.latest_version_timestamp),
              new Date()
            )}
          </>
        ) : (
          <>Loading</>
        )}
      </small>
    </Card.Footer>
  </Container>
);

export default memo(UpdateInfo);
