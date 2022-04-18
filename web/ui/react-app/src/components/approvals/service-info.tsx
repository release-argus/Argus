import {
  Button,
  Card,
  Container,
  ListGroup,
  OverlayTrigger,
  Tooltip,
} from "react-bootstrap";
import { ModalType, ServiceSummaryType } from "types/summary";
import { ReactElement, useCallback, useContext } from "react";
import {
  faAngleDoubleUp,
  faCheck,
  faInfo,
  faTimes,
} from "@fortawesome/free-solid-svg-icons";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { ModalContext } from "contexts/modal";
import { formatRelative } from "date-fns";

interface ServiceInfoData {
  service: ServiceSummaryType;
  updateAvailable: boolean;
  updateSkipped: boolean;
  setShowUpdateInfo: () => void;
}

export const ServiceInfo = ({
  service,
  updateAvailable,
  updateSkipped,
  setShowUpdateInfo,
}: ServiceInfoData): ReactElement => {
  const { handleModal } = useContext(ModalContext);

  const showModal = useCallback(
    (type: ModalType, service: ServiceSummaryType) => {
      handleModal(type, service);
    },
    [handleModal]
  );

  // If version hasn't been found or a new version has been found
  const serviceWarning =
    service?.status?.current_version === undefined ||
    service?.status?.current_version === "" ||
    (updateAvailable && !updateSkipped);

  return (
    <Container
      style={{
        padding: "0px",
      }}
      className={serviceWarning ? "alert-warning" : "default"}
    >
      <ListGroup className="list-group-flush">
        {service.webhook && updateAvailable && !updateSkipped ? (
          <>
            <ListGroup.Item
              key="update-available"
              className={"service-item update-options alert-warning"}
              variant="secondary"
            >
              Update available:
            </ListGroup.Item>
            <ListGroup.Item
              key="update-buttons"
              className={"service-item update-options alert-warning"}
              variant="secondary"
            >
              <Button
                key="details"
                className="btn-flex"
                variant="primary"
                onClick={() => setShowUpdateInfo()}
              >
                <FontAwesomeIcon icon={faInfo} />
              </Button>
              <Button
                key="approve"
                className="btn-flex"
                variant="success"
                onClick={() => showModal("SEND", service)}
              >
                <FontAwesomeIcon icon={faCheck} />
              </Button>
              <Button
                key="reject"
                className="btn-flex"
                variant="danger"
                onClick={() => showModal("SKIP", service)}
              >
                <FontAwesomeIcon icon={faTimes} color="white" />
              </Button>
            </ListGroup.Item>
          </>
        ) : (
          <ListGroup.Item
            key="current_v"
            variant={serviceWarning ? "warning" : "secondary"}
            className={
              "service-item" + (service.webhook ? "" : " justify-left")
            }
          >
            <OverlayTrigger
              key="current-version"
              placement="top"
              delay={{ show: 500, hide: 500 }}
              overlay={
                service?.status?.current_version_timestamp ? (
                  <Tooltip id={`tooltip-current-version`}>
                    <>
                      {formatRelative(
                        new Date(service.status.current_version_timestamp),
                        new Date()
                      )}
                    </>
                  </Tooltip>
                ) : (
                  <>Unknown</>
                )
              }
            >
              <p style={{ margin: 0 }}>
                Current version:
                <br />
                {service?.status?.current_version
                  ? service.status.current_version
                  : "Unknown"}{" "}
              </p>
            </OverlayTrigger>
            {service.webhook && (!updateAvailable || updateSkipped) && (
              <OverlayTrigger
                key="resend"
                placement="top"
                delay={{ show: 500, hide: 500 }}
                overlay={
                  <Tooltip id={`tooltip-resend`}>
                    {updateSkipped
                      ? "Approve this release"
                      : "Resend the WebHooks"}
                  </Tooltip>
                }
              >
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() =>
                    showModal(updateSkipped ? "SEND" : "RESEND", service)
                  }
                  disabled={service.loading}
                >
                  <FontAwesomeIcon
                    icon={updateSkipped ? faCheck : faAngleDoubleUp}
                  />
                </Button>
              </OverlayTrigger>
            )}
          </ListGroup.Item>
        )}
      </ListGroup>
      <Card.Footer
        className={
          serviceWarning || !service?.status?.last_queried
            ? "alert-warning"
            : ""
        }
      >
        <small
          className={
            "text-muted same-color" + (serviceWarning ? " alert-warning" : "")
          }
        >
          {service?.status?.last_queried ? (
            <>
              Queried{" "}
              {formatRelative(
                new Date(service.status.last_queried),
                new Date()
              )}
            </>
          ) : service.loading ? (
            "loading"
          ) : (
            "no successful queries"
          )}
        </small>
      </Card.Footer>
    </Container>
  );
};
